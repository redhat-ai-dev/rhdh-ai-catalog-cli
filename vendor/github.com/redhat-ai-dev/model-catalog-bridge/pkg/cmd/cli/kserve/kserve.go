package kserve

import (
	"bytes"
	"encoding/json"
	"fmt"
	serverapiv1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/config"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/util"
	"io"
	"k8s.io/klog/v2"
	"os"
	"strings"
)

const (
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
