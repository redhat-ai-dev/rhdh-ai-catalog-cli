package kserve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	serverapiv1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
	"strings"
)

const (
	kserveExamples = `
# Both owner and lifecycle are required parameters.  Examine Backstage Catalog documentation for details.
# This will query all the InferenceService instances in the current namespace and build Catalog Component, Resource, and
# API Entities from the data.
$ %s new-model kserve <owner> <lifecycle> <args...>

# This will set the URL, Token, and Skip TLS when accessing the cluster for InferenceService instances 
$ %s new-model kserve <owner> <lifecycle> --model-metadata-url=https://my-kubeflow.com --model-metadata-token=my-token --model-metadata-skip-tls=true

# This will set the Kubeconfig file to use when accessing the cluster for InferenceService instances
$ %s new-model kserve <owner> <lifecycle> --kubeconfig=/home/myid/my-kube.json

# This form will pull in only the InferenceService instances with the names 'inferenceservice1' and 'inferenceservice2'
# in the 'my-datascience-project'namespace in order to build Catalog Component, Resource, and API Entities.
$ %s new-model kserve owner lifecycle inferenceservice1 inferenceservice2 --namespace my-datascience-project
`
	sklearn     = "sklearn"
	xgboost     = "xgboost"
	tensorflow  = "tensorflow"
	pytorch     = "pytorch"
	triton      = "triton"
	onnx        = "onnx"
	huggingface = "huggingface"
	pmml        = "pmml"
	lightgbm    = "lightgbm"
	paddle      = "paddle"
)

type commonPopulator struct {
	owner     string
	lifecycle string
	is        *serverapiv1beta1.InferenceService
}

func (pop *commonPopulator) GetOwner() string {
	return pop.owner
}

func (pop *commonPopulator) GetLifecycle() string {
	return pop.lifecycle
}

func (pop *commonPopulator) GetName() string {
	return fmt.Sprintf("%s_%s", pop.is.Namespace, pop.is.Name)
}
func (pop *commonPopulator) GetDescription() string {
	return fmt.Sprintf("KServe instance %s:%s", pop.is.Namespace, pop.is.Name)
}

func (pop *commonPopulator) GetLinks() []backstage.EntityLink {
	links := []backstage.EntityLink{}
	if pop.is.Status.URL != nil {
		links = append(links, backstage.EntityLink{
			URL:   pop.is.Status.URL.String(),
			Title: backstage.LINK_API_URL,
			Type:  backstage.LINK_TYPE_WEBSITE,
			Icon:  backstage.LINK_ICON_WEBASSET,
		})
	}
	for componentType, componentStatus := range pop.is.Status.Components {

		if componentStatus.URL != nil {
			links = append(links, backstage.EntityLink{
				URL:   componentStatus.URL.String() + "/docs",
				Title: string(componentType) + " FastAPI URL",
				Icon:  backstage.LINK_ICON_WEBASSET,
				Type:  backstage.LINK_TYPE_WEBSITE,
			})
			links = append(links, backstage.EntityLink{
				URL:   componentStatus.URL.String(),
				Title: string(componentType) + " model serving URL",
				Icon:  backstage.LINK_ICON_WEBASSET,
				Type:  backstage.LINK_TYPE_WEBSITE,
			})
		}
		if componentStatus.RestURL != nil {
			links = append(links, backstage.EntityLink{
				URL:   componentStatus.RestURL.String(),
				Title: string(componentType) + " REST model serving URL",
				Icon:  backstage.LINK_ICON_WEBASSET,
				Type:  backstage.LINK_TYPE_WEBSITE,
			})
		}
		if componentStatus.GrpcURL != nil {
			links = append(links, backstage.EntityLink{
				URL:   componentStatus.GrpcURL.String(),
				Title: string(componentType) + " GRPC model serving URL",
				Icon:  backstage.LINK_ICON_WEBASSET,
				Type:  backstage.LINK_TYPE_WEBSITE,
			})
		}
	}
	return links
}

func (pop *commonPopulator) GetTags() []string {
	tags := []string{}
	predictor := pop.is.Spec.Predictor
	// one and only one predictor spec can be set
	switch {
	case predictor.SKLearn != nil:
		tags = append(tags, sklearn)
		fallthrough
	case predictor.XGBoost != nil:
		tags = append(tags, xgboost)
		fallthrough
	case predictor.Tensorflow != nil:
		tags = append(tags, tensorflow)
		fallthrough
	case predictor.PyTorch != nil:
		tags = append(tags, pytorch)
		fallthrough
	case predictor.Triton != nil:
		tags = append(tags, triton)
		fallthrough
	case predictor.ONNX != nil:
		tags = append(tags, onnx)
		fallthrough
	case predictor.HuggingFace != nil:
		tags = append(tags, huggingface)
		fallthrough
	case predictor.PMML != nil:
		tags = append(tags, pmml)
		fallthrough
	case predictor.LightGBM != nil:
		tags = append(tags, lightgbm)
		fallthrough
	case predictor.Paddle != nil:
		tags = append(tags, paddle)
		fallthrough
	case predictor.Model != nil:
		modelFormat := predictor.Model.ModelFormat
		tag := modelFormat.Name
		if modelFormat.Version != nil {
			tag = tag + "-" + *modelFormat.Version
		}
		tag = strings.ToLower(tag)
		tags = append(tags, tag)
	}
	explainer := pop.is.Spec.Explainer
	if explainer != nil && explainer.ART != nil {
		tags = append(tags, strings.ToLower(string(explainer.ART.Type)))
	}
	return tags
}

func (pop *commonPopulator) GetProvidedAPIs() []string {
	return []string{fmt.Sprintf("%s_%s", pop.is.Namespace, pop.is.Name)}
}

type componentPopulator struct {
	commonPopulator
}

func (pop *componentPopulator) GetDependsOn() []string {
	return []string{fmt.Sprintf("resource:%s_%s", pop.is.Namespace, pop.is.Name), fmt.Sprintf("api:%s_%s", pop.is.Namespace, pop.is.Name)}
}

func (pop *componentPopulator) GetTechdocRef() string {
	return "./"
}

func (pop *componentPopulator) GetDisplayName() string {
	return fmt.Sprintf("The %s model server", pop.GetName())
}

type resourcePopulator struct {
	commonPopulator
}

func (pop *resourcePopulator) GetDependencyOf() []string {
	return []string{fmt.Sprintf("component:%s_%s", pop.is.Namespace, pop.is.Name)}
}

func (pop *resourcePopulator) GetTechdocRef() string {
	return "resource/"
}

func (pop *resourcePopulator) GetDisplayName() string {
	return fmt.Sprintf("The %s ai model", pop.GetName())
}

type apiPopulator struct {
	commonPopulator
}

func (pop *apiPopulator) GetDependencyOf() []string {
	return []string{fmt.Sprintf("component:%s_%s", pop.is.Namespace, pop.is.Name)}
}

func (pop *apiPopulator) GetDefinition() string {
	if pop.is.Status.URL == nil {
		return ""
	}
	defBytes, _ := util.FetchURL(pop.is.Status.URL.String() + "/openapi.json")
	dst := bytes.Buffer{}
	json.Indent(&dst, defBytes, "", "    ")
	return dst.String()
}

func (pop *apiPopulator) GetTechdocRef() string {
	return "api/"
}

func (pop *apiPopulator) GetDisplayName() string {
	return fmt.Sprintf("The %s openapi", pop.GetName())
}

func SetupKServeClient(cfg *config.Config) {
	if cfg == nil {
		klog.Error("Command config is nil")
		klog.Flush()
		os.Exit(1)
	}
	if cfg.ServingClient != nil {
		return
	}
	if kubeconfig, err := util.GetK8sConfig(cfg); err != nil {
		err = fmt.Errorf("problem with kubeconfig: %s", err.Error())
		klog.Errorf("%s", err.Error())
		klog.Flush()
		os.Exit(1)
	} else {
		cfg.ServingClient = util.GetKServeClient(kubeconfig)
	}

	namespace := cfg.Namespace
	if len(namespace) == 0 {
		cfg.Namespace = util.GetCurrentProject()
	}

}

func NewCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kserve",
		Short:   "KServe related API",
		Long:    "Interact with KServe related instances on a K8s cluster to manage AI related catalog entities in a Backstage instance.",
		Example: strings.ReplaceAll(kserveExamples, "%s", util.ApplicationName),
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

			SetupKServeClient(cfg)
			namespace := cfg.Namespace
			servingClient := cfg.ServingClient

			if len(ids) != 0 {
				for _, id := range ids {
					is, err := servingClient.InferenceServices(namespace).Get(context.Background(), id, metav1.GetOptions{})
					if err != nil {
						klog.Errorf("inference service retrieval error for %s:%s: %s", namespace, id, err.Error())
						klog.Flush()
						return err
					}

					err = callBackstagePrinters(owner, lifecycle, is, cmd)
					if err != nil {
						return err
					}
				}
			} else {
				isl, err := servingClient.InferenceServices(namespace).List(context.Background(), metav1.ListOptions{})
				if err != nil {
					klog.Errorf("inference service retrieval error for %s: %s", namespace, err.Error())
					klog.Flush()
					return err
				}
				for _, is := range isl.Items {
					err = callBackstagePrinters(owner, lifecycle, &is, cmd)
					if err != nil {
						klog.Errorf("%s", err.Error())
						klog.Flush()
						return err
					}
				}
			}
			return nil
		},
	}

	return cmd
}

func callBackstagePrinters(owner, lifecycle string, is *serverapiv1beta1.InferenceService, cmd *cobra.Command) error {
	compPop := componentPopulator{}
	compPop.owner = owner
	compPop.lifecycle = lifecycle
	compPop.is = is
	err := backstage.PrintComponent(&compPop, cmd)
	if err != nil {
		return err
	}

	resPop := resourcePopulator{}
	resPop.owner = owner
	resPop.lifecycle = lifecycle
	resPop.is = is
	err = backstage.PrintResource(&resPop, cmd)
	if err != nil {
		return err
	}

	apiPop := apiPopulator{}
	apiPop.owner = owner
	apiPop.lifecycle = lifecycle
	apiPop.is = is
	err = backstage.PrintAPI(&apiPop, cmd)
	return err
}
