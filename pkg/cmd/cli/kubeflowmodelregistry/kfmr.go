package kubeflowmodelregistry

import (
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	"github.com/kubeflow/model-registry/pkg/openapi"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"strings"
)

const (
	kubeflowExample = `
# Both owner and lifecycle are required parameters.  Examine Backstage Catalog documentation for details.
# This will query all the RegisteredModel, ModelVersion, and ModelArtifact instances in the Kubeflow Model Registry and build Catalog Component, Resource, and
# API Entities from the data.
$ %s new-model kubeflow <owner> <lifecycle> <args...>

# This will set the URL, Token, and Skip TLS when accessing Kubeflow
$ %s new-model kubeflow <owner> <lifecycle> --model-metadata-url=https://my-kubeflow.com --model-metadata-token=my-token --model-metadata-skip-tls=true

# This form will pull in only the RegisteredModels with the specified IDs '1' and '2' and their ModelVersion and ModelArtifact
# children in order to build Catalog Component, Resource, and API Entities.
$ %s new-model kubeflow <owner> <lifecycle> 1 2 
`
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
				err := fmt.Errorf("need to specify an owner and lifecycle setting")
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

			if len(ids) == 0 {
				var err error
				var rms []openapi.RegisteredModel
				rms, err = kfmr.ListRegisteredModels()
				if err != nil {
					klog.Errorf("list registered models error: %s", err.Error())
					klog.Flush()
					return err
				}
				for _, rm := range rms {
					var mvs []openapi.ModelVersion
					var mas map[string][]openapi.ModelArtifact
					mvs, mas, err = callKubeflowREST(*rm.Id, kfmr)
					if err != nil {
						klog.Errorf("%s", err.Error())
						klog.Flush()
						return err
					}
					err = callBackstagePrinters(owner, lifecycle, &rm, mvs, mas, cmd)
					if err != nil {
						klog.Errorf("print model catalog: %s", err.Error())
						klog.Flush()
						return err
					}
				}
			} else {
				for _, id := range ids {
					rm, err := kfmr.GetRegisteredModel(id)
					if err != nil {
						klog.Errorf("get registered model error for %s: %s", id, err.Error())
						klog.Flush()
						return err
					}
					var mvs []openapi.ModelVersion
					var mas map[string][]openapi.ModelArtifact
					mvs, mas, err = callKubeflowREST(*rm.Id, kfmr)
					if err != nil {
						klog.Errorf("get model version/artifact error for %s: %s", id, err.Error())
						klog.Flush()
						return err
					}
					err = callBackstagePrinters(owner, lifecycle, rm, mvs, mas, cmd)
				}
			}
			return nil
		},
	}

	return cmd
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

func callBackstagePrinters(owner, lifecycle string, rm *openapi.RegisteredModel, mvs []openapi.ModelVersion, mas map[string][]openapi.ModelArtifact, cmd *cobra.Command) error {
	compPop := componentPopulator{}
	compPop.owner = owner
	compPop.lifecycle = lifecycle
	compPop.registeredModel = rm
	compPop.modelVersions = mvs
	compPop.modelArtifacts = mas
	err := backstage.PrintComponent(&compPop, cmd)
	if err != nil {
		return err
	}

	resPop := resourcePopulator{}
	resPop.owner = owner
	resPop.lifecycle = lifecycle
	resPop.registeredModel = rm
	for _, mv := range mvs {
		resPop.modelVersion = &mv
		m, _ := mas[*mv.Id]
		resPop.modelArtifacts = m
		err = backstage.PrintResource(&resPop, cmd)
		if err != nil {
			return err
		}
	}

	apiPop := apiPopulator{}
	apiPop.owner = owner
	apiPop.lifecycle = lifecycle
	apiPop.registeredModel = rm
	for _, arr := range mas {
		for _, ma := range arr {
			apiPop.modelArtifact = &ma
			err = backstage.PrintAPI(&apiPop, cmd)
			return err
		}
	}
	return nil
}

type commonPopulator struct {
	owner           string
	lifecycle       string
	registeredModel *openapi.RegisteredModel
}

func (pop *commonPopulator) GetOwner() string {
	if pop.registeredModel.Owner != nil {
		return *pop.registeredModel.Owner
	}
	return pop.owner
}

func (pop *commonPopulator) GetLifecycle() string {
	return pop.lifecycle
}

func (pop *commonPopulator) GetDescription() string {
	if pop.registeredModel.Description != nil {
		return *pop.registeredModel.Description
	}
	return ""
}

// TODO won't have API until KubeFlow Model Registry gets us inferenceservice endpoints
func (pop *commonPopulator) GetProvidedAPIs() []string {
	return []string{}
}

type componentPopulator struct {
	commonPopulator
	modelVersions  []openapi.ModelVersion
	modelArtifacts map[string][]openapi.ModelArtifact
}

func (pop *componentPopulator) GetName() string {
	return pop.registeredModel.Name
}

// TODO Until we get the inferenceservice endpoint URL associated with the model registry related API won't have component links
func (pop *componentPopulator) GetLinks() []backstage.EntityLink {
	return []backstage.EntityLink{}
}

func (pop *componentPopulator) GetTags() []string {
	tags := []string{}
	for key, value := range pop.registeredModel.GetCustomProperties() {
		tags = append(tags, fmt.Sprintf("%s:%v", key, value.GetActualInstance()))
	}

	return tags
}

func (pop *componentPopulator) GetDependsOn() []string {
	depends := []string{}
	for _, mv := range pop.modelVersions {
		depends = append(depends, "resource:"+mv.Name)
	}
	for _, mas := range pop.modelArtifacts {
		for _, ma := range mas {
			depends = append(depends, "api:"+*ma.Name)
		}
	}
	return depends
}

func (pop *componentPopulator) GetTechdocRef() string {
	return "./"
}

func (pop *componentPopulator) GetDisplayName() string {
	return fmt.Sprintf("The %s model server", pop.GetName())
}

type resourcePopulator struct {
	commonPopulator
	modelVersion   *openapi.ModelVersion
	modelArtifacts []openapi.ModelArtifact
}

func (pop *resourcePopulator) GetName() string {
	return pop.modelVersion.Name
}

func (pop *resourcePopulator) GetTechdocRef() string {
	return "resource/"
}

func (pop *resourcePopulator) GetLinks() []backstage.EntityLink {
	links := []backstage.EntityLink{}
	for _, ma := range pop.modelArtifacts {
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

func (pop *resourcePopulator) GetTags() []string {
	tags := []string{}
	for key := range pop.modelVersion.GetCustomProperties() {
		tags = append(tags, key)
	}

	for _, ma := range pop.modelArtifacts {
		for k := range ma.GetCustomProperties() {
			tags = append(tags, k)
		}
	}
	return tags
}

func (pop *resourcePopulator) GetDependencyOf() []string {
	return []string{fmt.Sprintf("component:%s", pop.registeredModel.Name)}
}

func (pop *resourcePopulator) GetDisplayName() string {
	return fmt.Sprintf("The %s ai model", pop.GetName())
}

// TODO Until we get the inferenceservice endpoint URL associated with the model registry related API won't have much for Backstage API here
type apiPopulator struct {
	commonPopulator
	modelArtifact *openapi.ModelArtifact
}

func (pop *apiPopulator) GetName() string {
	return *pop.modelArtifact.Name
}

func (pop *apiPopulator) GetDependencyOf() []string {
	return []string{fmt.Sprintf("component:%s", pop.registeredModel.Name)}
}

func (pop *apiPopulator) GetDefinition() string {
	// definition must be set to something to pass backstage validation
	return "no-definition-yet"
}

func (pop *apiPopulator) GetTechdocRef() string {
	return "api/"
}

func (pop *apiPopulator) GetTags() []string {
	return []string{}
}

func (pop *apiPopulator) GetLinks() []backstage.EntityLink {
	return []backstage.EntityLink{}
}

func (pop *apiPopulator) GetDisplayName() string {
	return fmt.Sprintf("The %s openapi", pop.GetName())
}
