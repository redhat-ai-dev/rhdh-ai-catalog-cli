package kubeflowmodelregistry

import (
	"context"
	"fmt"
	serverv1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/kubeflow/model-registry/pkg/openapi"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/kserve"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	"github.com/spf13/cobra"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	kubeflowExample = `
# Both Owner and Lifecycle are required parameters.  Examine Backstage Catalog documentation for details.
# This will query all the RegisteredModel, ModelVersion, ModelArtifact, and InferenceService instances in the Kubeflow Model Registry and build Catalog Component, Resource, and
# API Entities from the data.
$ %s new-model kubeflow <Owner> <Lifecycle> <args...>

# This will set the URL, Token, and Skip TLS when accessing Kubeflow
$ %s new-model kubeflow <Owner> <Lifecycle> --model-metadata-url=https://my-kubeflow.com --model-metadata-token=my-token --model-metadata-skip-tls=true

# This form will pull in only the RegisteredModels with the specified IDs '1' and '2' and the ModelVersion, ModelArtifact, and InferenceService
# artifacts that are linked to those RegisteredModels in order to build Catalog Component, Resource, and API Entities.
$ %s new-model kubeflow <Owner> <Lifecycle> 1 2 
`

	// pulled from makeValidator.ts in the catalog-model package in core backstage
	tagRegexp = "^[a-z0-9:+#]+(\\-[a-z0-9:+#]+)*$"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kubeflow",
		Aliases: []string{"kf"},
		Short:   "Kubeflow Model Registry related API",
		Long:    "Interact with the Kubeflow Model Registry REST API as part of managing AI related catalog entities in a Backstage instance.",
		Example: strings.ReplaceAll(kubeflowExample, "%s", util.ApplicationName),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := []string{}

			if len(args) < 2 {
				err := fmt.Errorf("need to specify an Owner and Lifecycle setting")
				klog.Errorf("%s", err.Error())
				klog.Flush()
				return err
			}
			owner := args[0]
			lifecycle := args[1]

			if len(args) > 2 {
				ids = args[2:]
			}

			kfmr := SetupKubeflowRESTClient(cfg)

			_, _, err := LoopOverKFMR(owner, lifecycle, ids, cmd.OutOrStdout(), kfmr, nil)
			return err

		},
	}

	return cmd
}

func LoopOverKFMR(owner, lifecycle string, ids []string, writer io.Writer, kfmr *KubeFlowRESTClientWrapper, client client.Client) ([]openapi.RegisteredModel, map[string][]openapi.ModelVersion, error) {
	var err error
	var isl []openapi.InferenceService
	rmArray := []openapi.RegisteredModel{}
	mvsMap := map[string][]openapi.ModelVersion{}

	isl, err = kfmr.ListInferenceServices()

	if len(ids) == 0 {
		var rms []openapi.RegisteredModel
		rms, err = kfmr.ListRegisteredModels()
		if err != nil {
			klog.Errorf("list registered models error: %s", err.Error())
			klog.Flush()
			return nil, nil, err
		}
		for _, rm := range rms {
			var mvs []openapi.ModelVersion
			var mas map[string][]openapi.ModelArtifact
			mvs, mas, err = callKubeflowREST(*rm.Id, kfmr)
			if err != nil {
				klog.Errorf("%s", err.Error())
				klog.Flush()
				return nil, nil, err
			}
			err = CallBackstagePrinters(owner, lifecycle, &rm, mvs, mas, isl, nil, kfmr, client, writer)
			if err != nil {
				klog.Errorf("print model catalog: %s", err.Error())
				klog.Flush()
				return nil, nil, err
			}
			rmArray = append(rmArray, rm)
			mvsMap[rm.Name] = mvs
		}
	} else {
		for _, id := range ids {
			var rm *openapi.RegisteredModel
			rm, err = kfmr.GetRegisteredModel(id)
			if err != nil {
				klog.Errorf("get registered model error for %s: %s", id, err.Error())
				klog.Flush()
				return nil, nil, err
			}
			var mvs []openapi.ModelVersion
			var mas map[string][]openapi.ModelArtifact
			mvs, mas, err = callKubeflowREST(*rm.Id, kfmr)
			if err != nil {
				klog.Errorf("get model version/artifact error for %s: %s", id, err.Error())
				klog.Flush()
				return nil, nil, err
			}
			err = CallBackstagePrinters(owner, lifecycle, rm, mvs, mas, isl, nil, kfmr, client, writer)
			rmArray = append(rmArray, *rm)
			mvsMap[rm.Name] = mvs
		}
	}
	return rmArray, mvsMap, nil
}

func callKubeflowREST(id string, kfmr *KubeFlowRESTClientWrapper) (mvs []openapi.ModelVersion, ma map[string][]openapi.ModelArtifact, err error) {
	mvs, err = kfmr.ListModelVersions(id)
	if err != nil {
		klog.Errorf("ERROR: error list model versions for %s: %s", id, err.Error())
		return
	}
	ma = map[string][]openapi.ModelArtifact{}
	for _, mv := range mvs {
		var v []openapi.ModelArtifact
		v, err = kfmr.ListModelArtifacts(*mv.Id)
		if err != nil {
			klog.Errorf("ERROR error list model artifacts for %s:%s: %s", id, *mv.Id, err.Error())
			return
		}
		if len(v) == 0 {
			v, err = kfmr.ListModelArtifacts(id)
			if err != nil {
				klog.Errorf("ERROR error list model artifacts for %s:%s: %s", id, *mv.Id, err.Error())
				return
			}
		}
		ma[*mv.Id] = v
	}
	return
}

func CallBackstagePrinters(owner, lifecycle string, rm *openapi.RegisteredModel, mvs []openapi.ModelVersion, mas map[string][]openapi.ModelArtifact, isl []openapi.InferenceService, is *serverv1beta1.InferenceService, kfmr *KubeFlowRESTClientWrapper, client client.Client, writer io.Writer) error {
	compPop := ComponentPopulator{}
	compPop.Owner = owner
	compPop.Lifecycle = lifecycle
	compPop.Kfmr = kfmr
	compPop.RegisteredModel = rm
	compPop.ModelVersions = mvs
	compPop.ModelArtifacts = mas
	compPop.InferenceServices = isl
	compPop.Kis = is
	compPop.CtrlClient = client
	err := backstage.PrintComponent(&compPop, writer)
	if err != nil {
		return err
	}

	resPop := ResourcePopulator{}
	resPop.Owner = owner
	resPop.Lifecycle = lifecycle
	resPop.Kfmr = kfmr
	resPop.RegisteredModel = rm
	resPop.Kis = is
	resPop.CtrlClient = client
	for _, mv := range mvs {
		resPop.ModelVersion = &mv
		m, _ := mas[*mv.Id]
		resPop.ModelArtifacts = m
		err = backstage.PrintResource(&resPop, writer)
		if err != nil {
			return err
		}
	}

	apiPop := ApiPopulator{}
	apiPop.Owner = owner
	apiPop.Lifecycle = lifecycle
	apiPop.Kfmr = kfmr
	apiPop.RegisteredModel = rm
	apiPop.InferenceServices = isl
	apiPop.Kis = is
	apiPop.CtrlClient = client
	return backstage.PrintAPI(&apiPop, writer)
}

type CommonPopulator struct {
	Owner             string
	Lifecycle         string
	RegisteredModel   *openapi.RegisteredModel
	InferenceServices []openapi.InferenceService
	Kfmr              *KubeFlowRESTClientWrapper
	Kis               *serverv1beta1.InferenceService
	CtrlClient        client.Client
}

func (pop *CommonPopulator) GetOwner() string {
	if pop.RegisteredModel.Owner != nil {
		return *pop.RegisteredModel.Owner
	}
	return pop.Owner
}

func (pop *CommonPopulator) GetLifecycle() string {
	return pop.Lifecycle
}

func (pop *CommonPopulator) GetDescription() string {
	if pop.RegisteredModel.Description != nil {
		return *pop.RegisteredModel.Description
	}
	return ""
}

// TODO won't have API until KubeFlow Model Registry gets us inferenceservice endpoints
func (pop *CommonPopulator) GetProvidedAPIs() []string {
	return []string{}
}

type ComponentPopulator struct {
	CommonPopulator
	ModelVersions  []openapi.ModelVersion
	ModelArtifacts map[string][]openapi.ModelArtifact
}

func (pop *ComponentPopulator) GetName() string {
	return pop.RegisteredModel.Name
}

func (pop *ComponentPopulator) GetLinks() []backstage.EntityLink {
	links := pop.getLinksFromInferenceServices()
	// GGM maybe multi resource / multi model indication
	for _, maa := range pop.ModelArtifacts {
		for _, ma := range maa {
			if ma.Uri != nil {
				links = append(links, backstage.EntityLink{
					URL:   *ma.Uri,
					Title: ma.GetDescription(),
					Icon:  backstage.LINK_ICON_WEBASSET,
					Type:  backstage.LINK_TYPE_WEBSITE,
				})
			}
		}
	}

	return links
}

func (pop *CommonPopulator) getLinksFromInferenceServices() []backstage.EntityLink {
	links := []backstage.EntityLink{}
	for _, is := range pop.InferenceServices {
		var rmid *string
		var ok bool
		rmid, ok = pop.RegisteredModel.GetIdOk()
		if !ok {
			continue
		}
		if is.RegisteredModelId != *rmid {
			continue
		}
		var iss *openapi.InferenceServiceState
		iss, ok = is.GetDesiredStateOk()
		if !ok {
			continue
		}
		if *iss != openapi.INFERENCESERVICESTATE_DEPLOYED {
			continue
		}
		se, err := pop.Kfmr.GetServingEnvironment(is.ServingEnvironmentId)
		if err != nil {
			klog.Errorf("ComponentPopulator GetLinks: %s", err.Error())
			continue
		}
		if pop.Kis == nil {
			kisns := se.GetName()
			kisnm := is.GetRuntime()
			var kis *serverv1beta1.InferenceService
			if pop.Kfmr != nil && pop.Kfmr.Config != nil && pop.Kfmr.Config.ServingClient != nil {
				kis, err = pop.Kfmr.Config.ServingClient.InferenceServices(kisns).Get(context.Background(), kisnm, metav1.GetOptions{})
			}
			if kis == nil && pop.CtrlClient != nil {
				kis = &serverv1beta1.InferenceService{}
				err = pop.CtrlClient.Get(context.Background(), types.NamespacedName{Namespace: kisns, Name: kisnm}, kis)
			}

			if err != nil {
				klog.Errorf("ComponentPopulator GetLinks: %s", err.Error())
				continue
			}
			pop.Kis = kis
		}
		kpop := kserve.CommonPopulator{InferSvc: pop.Kis}
		links = append(links, kpop.GetLinks()...)
	}
	return links
}

func (pop *ComponentPopulator) GetTags() []string {
	tags := []string{}
	regex, _ := regexp.Compile(tagRegexp)
	for key, value := range pop.RegisteredModel.GetCustomProperties() {
		if !regex.MatchString(key) {
			klog.Infof("skipping custom prop %s for tags", key)
			continue
		}
		tag := key
		if value.MetadataStringValue != nil {
			strVal := value.MetadataStringValue.StringValue
			if !regex.MatchString(fmt.Sprintf("%v", strVal)) {
				klog.Infof("skipping custom prop value %v for tags", value.GetActualInstance())
				continue
			}
			tag = fmt.Sprintf("%s-%s", tag, strVal)
		}

		if len(tag) > 63 {
			klog.Infof("skipping tag %s because its length is greater than 63", tag)
		}

		tags = append(tags, tag)
	}

	return tags
}

func (pop *ComponentPopulator) GetDependsOn() []string {
	depends := []string{}
	for _, mv := range pop.ModelVersions {
		depends = append(depends, "resource:"+mv.Name)
	}
	for _, mas := range pop.ModelArtifacts {
		for _, ma := range mas {
			depends = append(depends, "api:"+*ma.Name)
		}
	}
	return depends
}

func (pop *ComponentPopulator) GetTechdocRef() string {
	return "./"
}

func (pop *ComponentPopulator) GetDisplayName() string {
	return pop.GetName()
}

type ResourcePopulator struct {
	CommonPopulator
	ModelVersion   *openapi.ModelVersion
	ModelArtifacts []openapi.ModelArtifact
}

func (pop *ResourcePopulator) GetName() string {
	return pop.ModelVersion.Name
}

func (pop *ResourcePopulator) GetTechdocRef() string {
	return "resource/"
}

func (pop *ResourcePopulator) GetLinks() []backstage.EntityLink {
	links := []backstage.EntityLink{}
	// GGM maybe multi resource / multi model indication
	for _, ma := range pop.ModelArtifacts {
		if ma.Uri != nil {
			links = append(links, backstage.EntityLink{
				URL:   *ma.Uri,
				Title: ma.GetDescription(),
				Icon:  backstage.LINK_ICON_WEBASSET,
				Type:  backstage.LINK_TYPE_WEBSITE,
			})
		}
	}
	return links
}

func (pop *ResourcePopulator) GetTags() []string {
	tags := []string{}
	regex, _ := regexp.Compile(tagRegexp)
	for key, value := range pop.ModelVersion.GetCustomProperties() {
		if !regex.MatchString(key) {
			klog.Infof("skipping custom prop %s for tags", key)
			continue
		}
		tag := key
		if value.MetadataStringValue != nil {
			strVal := value.MetadataStringValue.StringValue
			if !regex.MatchString(fmt.Sprintf("%v", strVal)) {
				klog.Infof("skipping custom prop value %v for tags", value.GetActualInstance())
				continue
			}
			tag = fmt.Sprintf("%s-%s", tag, strVal)
		}
		if len(tag) > 63 {
			klog.Infof("skipping tag %s because its length is greater than 63", tag)
		}

		tags = append(tags, tag)
	}

	for _, ma := range pop.ModelArtifacts {
		for k, v := range ma.GetCustomProperties() {
			if !regex.MatchString(k) {
				klog.Infof("skipping custom prop %s for tags", k)
				continue
			}
			tag := k
			if v.MetadataStringValue != nil {
				strVal := v.MetadataStringValue.StringValue
				if !regex.MatchString(fmt.Sprintf("%v", strVal)) {
					klog.Infof("skipping custom prop value %v for tags", v.GetActualInstance())
					continue
				}
				tag = fmt.Sprintf("%s-%s", tag, strVal)
			}

			if len(tag) > 63 {
				klog.Infof("skipping tag %s because its length is greater than 63", tag)
			}

			tags = append(tags, tag)
		}
	}
	return tags
}

func (pop *ResourcePopulator) GetDependencyOf() []string {
	return []string{fmt.Sprintf("component:%s", pop.RegisteredModel.Name)}
}

func (pop *ResourcePopulator) GetDisplayName() string {
	return pop.GetName()
}

// TODO Until we get the inferenceservice endpoint URL associated with the model registry related API won't have much for Backstage API here
type ApiPopulator struct {
	CommonPopulator
}

func (pop *ApiPopulator) GetName() string {
	return pop.RegisteredModel.Name
}

func (pop *ApiPopulator) GetDependencyOf() []string {
	return []string{fmt.Sprintf("component:%s", pop.RegisteredModel.Name)}
}

func (pop *ApiPopulator) GetDefinition() string {
	// definition must be set to something to pass backstage validation
	return "no-definition-yet"
}

func (pop *ApiPopulator) GetTechdocRef() string {
	// TODO in theory the Kfmr modelcard support when it arrives will influcen this
	return "api/"
}

func (pop *ApiPopulator) GetTags() []string {
	return []string{}
}

func (pop *ApiPopulator) GetLinks() []backstage.EntityLink {
	return pop.getLinksFromInferenceServices()
}

func (pop *ApiPopulator) GetDisplayName() string {
	return pop.GetName()
}
