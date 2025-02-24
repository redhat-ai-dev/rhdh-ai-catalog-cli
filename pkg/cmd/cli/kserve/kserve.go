package kserve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	serverapiv1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	"github.com/spf13/cobra"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
	"strings"
)

const (
	kserveExamples = `
# Both Owner and Lifecycle are required parameters.  Examine Backstage Catalog documentation for details.
# This will query all the InferenceService instances in the current namespace and build Catalog Component, Resource, and
# API Entities from the data.
$ %s new-model kserve <Owner> <Lifecycle> <args...>

# This example shows the flags that will set the URL, Token, and Skip TLS when accessing the cluster for InferenceService instances 
$ %s new-model kserve <Owner> <Lifecycle> --model-metadata-url=https://my-kubeflow.com --model-metadata-token=my-token --model-metadata-skip-tls=true

# The '--kubeconfig' flag can be used to set the location of the configuration file used for accessing the credentials
# used for interacting with the Kubernetes cluster.
$ %s new-model kserve <Owner> <Lifecycle> --kubeconfig=/home/myid/my-kube.json

# This form will pull in only the InferenceService instances with the names 'inferenceservice1' and 'inferenceservice2'
# in the 'my-datascience-project'namespace in order to build Catalog Component, Resource, and API Entities.
$ %s new-model kserve Owner Lifecycle inferenceservice1 inferenceservice2 --namespace my-datascience-project
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

type CommonPopulator struct {
	Owner     string
	Lifecycle string
	InferSvc  *serverapiv1beta1.InferenceService
}

func (pop *CommonPopulator) GetOwner() string {
	return pop.Owner
}

func (pop *CommonPopulator) GetLifecycle() string {
	return pop.Lifecycle
}

func (pop *CommonPopulator) GetName() string {
	if pop.InferSvc == nil {
		return ""
	}
	return fmt.Sprintf("%s_%s", pop.InferSvc.Namespace, pop.InferSvc.Name)
}
func (pop *CommonPopulator) GetDescription() string {
	return fmt.Sprintf("KServe instance %s:%s", pop.InferSvc.Namespace, pop.InferSvc.Name)
}

func (pop *CommonPopulator) GetLinks() []backstage.EntityLink {
	links := []backstage.EntityLink{}
	if pop.InferSvc == nil {
		return links
	}
	if pop.InferSvc.Status.URL != nil {
		links = append(links, backstage.EntityLink{
			URL:   pop.InferSvc.Status.URL.String(),
			Title: backstage.LINK_API_URL,
			Type:  backstage.LINK_TYPE_WEBSITE,
			Icon:  backstage.LINK_ICON_WEBASSET,
		})
	}
	for componentType, componentStatus := range pop.InferSvc.Status.Components {

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

func (pop *CommonPopulator) GetTags() []string {
	tags := []string{}
	if pop.InferSvc == nil {
		return tags
	}
	predictor := pop.InferSvc.Spec.Predictor
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
	explainer := pop.InferSvc.Spec.Explainer
	if explainer != nil && explainer.ART != nil {
		tags = append(tags, strings.ToLower(string(explainer.ART.Type)))
	}
	return tags
}

func (pop *CommonPopulator) GetProvidedAPIs() []string {
	if pop.InferSvc == nil {
		return []string{}
	}
	return []string{fmt.Sprintf("%s_%s", pop.InferSvc.Namespace, pop.InferSvc.Name)}
}

type ComponentPopulator struct {
	CommonPopulator
}

func (pop *ComponentPopulator) GetDependsOn() []string {
	return []string{fmt.Sprintf("resource:%s_%s", pop.InferSvc.Namespace, pop.InferSvc.Name), fmt.Sprintf("api:%s_%s", pop.InferSvc.Namespace, pop.InferSvc.Name)}
}

func (pop *ComponentPopulator) GetTechdocRef() string {
	return "./"
}

func (pop *ComponentPopulator) GetDisplayName() string {
	return pop.GetName()
}

type ResourcePopulator struct {
	CommonPopulator
}

func (pop *ResourcePopulator) GetDependencyOf() []string {
	return []string{fmt.Sprintf("component:%s_%s", pop.InferSvc.Namespace, pop.InferSvc.Name)}
}

func (pop *ResourcePopulator) GetTechdocRef() string {
	return "resource/"
}

func (pop *ResourcePopulator) GetDisplayName() string {
	return pop.GetName()
}

type ApiPopulator struct {
	CommonPopulator
}

func (pop *ApiPopulator) GetDependencyOf() []string {
	if pop.InferSvc == nil {
		return []string{}
	}
	return []string{fmt.Sprintf("component:%s_%s", pop.InferSvc.Namespace, pop.InferSvc.Name)}
}

func (pop *ApiPopulator) GetDefinition() string {
	if pop.InferSvc.Status.URL == nil {
		return ""
	}
	defBytes, _ := util.FetchURL(pop.InferSvc.Status.URL.String() + "/openapi.json")
	dst := bytes.Buffer{}
	json.Indent(&dst, defBytes, "", "    ")
	return dst.String()
}

func (pop *ApiPopulator) GetTechdocRef() string {
	return "api/"
}

func (pop *ApiPopulator) GetDisplayName() string {
	return pop.GetName()
}

func SetupKServeClient(cfg *config.Config) {
	if cfg == nil {
		klog.Error("Command config InferSvc nil")
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

					err = CallBackstagePrinters(owner, lifecycle, is, cmd.OutOrStdout())
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
					err = CallBackstagePrinters(owner, lifecycle, &is, cmd.OutOrStdout())
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

func CallBackstagePrinters(owner, lifecycle string, is *serverapiv1beta1.InferenceService, writer io.Writer) error {
	compPop := ComponentPopulator{}
	compPop.Owner = owner
	compPop.Lifecycle = lifecycle
	compPop.InferSvc = is
	err := backstage.PrintComponent(&compPop, writer)
	if err != nil {
		return err
	}

	resPop := ResourcePopulator{}
	resPop.Owner = owner
	resPop.Lifecycle = lifecycle
	resPop.InferSvc = is
	err = backstage.PrintResource(&resPop, writer)
	if err != nil {
		return err
	}

	apiPop := ApiPopulator{}
	apiPop.Owner = owner
	apiPop.Lifecycle = lifecycle
	apiPop.InferSvc = is
	err = backstage.PrintAPI(&apiPop, writer)
	return err
}
