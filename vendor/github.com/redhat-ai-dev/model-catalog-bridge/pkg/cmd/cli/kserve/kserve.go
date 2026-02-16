package kserve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	serverapiv1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/config"
	brdgtypes "github.com/redhat-ai-dev/model-catalog-bridge/pkg/types"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/util"
	"github.com/redhat-ai-dev/model-catalog-bridge/schema/types/golang"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	Owner      string
	Lifecycle  string
	InferSvc   *serverapiv1beta1.InferenceService
	CtrlClient client.Client
	Ctx        context.Context
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

func CallBackstagePrinters(ctx context.Context, owner, lifecycle string, is *serverapiv1beta1.InferenceService, client client.Client, writer io.Writer, format brdgtypes.NormalizerFormat) error {
	compPop := ComponentPopulator{}
	compPop.Owner = owner
	compPop.Lifecycle = lifecycle
	compPop.InferSvc = is
	compPop.CtrlClient = client
	compPop.Ctx = ctx

	switch format {
	case brdgtypes.JsonArrayForamt:
		mcPop := ModelCatalogPopulator{CommonSchemaPopulator: CommonSchemaPopulator{compPop}}
		msPop := ModelServerPopulator{
			CommonSchemaPopulator: CommonSchemaPopulator{compPop},
			ApiPop:                ModelServerAPIPopulator{CommonSchemaPopulator: CommonSchemaPopulator{compPop}},
		}
		mcPop.MSPop = &msPop
		return backstage.PrintModelCatalogPopulator(&mcPop, writer)

	case brdgtypes.CatalogInfoYamlFormat:
	default:
		err := backstage.PrintComponent(&compPop, writer)
		if err != nil {
			return err
		}

		resPop := ResourcePopulator{}
		resPop.Owner = owner
		resPop.Lifecycle = lifecycle
		resPop.InferSvc = is
		resPop.CtrlClient = client
		resPop.Ctx = ctx
		err = backstage.PrintResource(&resPop, writer)
		if err != nil {
			return err
		}

		apiPop := ApiPopulator{}
		apiPop.Owner = owner
		apiPop.Lifecycle = lifecycle
		apiPop.InferSvc = is
		apiPop.CtrlClient = client
		apiPop.Ctx = ctx
		err = backstage.PrintAPI(&apiPop, writer)
		return err
	}
	return nil
}

// json array schema populator

type CommonSchemaPopulator struct {
	// reuse the component populator as it houses all the KFMR artifacts of noew
	ComponentPopulator
}

func fixKeyForAnnotation(key string) string {
	key = strings.ToLower(key)
	replacer := strings.NewReplacer(" ", "")
	key = replacer.Replace(key)
	return key
}

func commonGetStringPropVal(key string, is *serverapiv1beta1.InferenceService) *string {
	if is == nil {
		return nil
	}
	retString := ""

	if is.Annotations == nil {
		return nil
	}

	val, ok := is.Annotations[fmt.Sprintf("%s%s", brdgtypes.AnnotationPrefix, fixKeyForAnnotation(key))]
	if !ok {
		return nil
	}
	retString = val

	return &retString
}

type ModelServerPopulator struct {
	CommonSchemaPopulator
	ApiPop ModelServerAPIPopulator
}

func (m *ModelServerPopulator) getStringPropVal(key string) *string {
	return commonGetStringPropVal(key, m.InferSvc)
}

func (m *ModelServerPopulator) GetUsage() *string {
	return m.getStringPropVal(brdgtypes.UsageKey)
}

func (m *ModelServerPopulator) GetHomepageURL() *string {
	return m.getStringPropVal(brdgtypes.HomepageURLKey)
}

func (m *ModelServerPopulator) GetAuthentication() *bool {
	auth := false
	// when auth is configured, a service account is created whose name is prefixed with the inference service's name, and with the
	// inference service set as an owner reference

	listOptions := &client.ListOptions{Namespace: m.InferSvc.Namespace}
	saList := &corev1.ServiceAccountList{}
	err := m.CtrlClient.List(m.Ctx, saList, listOptions)
	if err != nil {
		return &auth
	}
	for _, sa := range saList.Items {
		if sa.OwnerReferences == nil {
			continue
		}
		for _, o := range sa.OwnerReferences {
			if o.Kind == "InferenceService" &&
				o.Name == m.InferSvc.Name {
				auth = true
				break
			}
		}
	}
	return &auth
}

// GetName returns the inference server name, sanitized to meet the following criteria
// "a string that is sequences of [a-zA-Z0-9] separated by any of [-_.], at most 63 characters in total"
func (m *ModelServerPopulator) GetName() string {
	name := fmt.Sprintf("%s-%s", util.SanitizeName(m.InferSvc.Namespace), util.SanitizeName(m.InferSvc.Name))
	return util.SanitizeName(name)
}

func (m *ModelServerPopulator) GetTags() []string {
	tags := []string{}
	for k, v := range m.InferSvc.Labels {
		tag := fmt.Sprintf("%s-%s", util.SanitizeName(k), util.SanitizeName(v))
		tags = append(tags, util.SanitizeName(tag))
	}
	return tags
}

func (m *ModelServerPopulator) GetAPI() *golang.API {
	m.ApiPop.Ctx = m.Ctx
	routeExternalURL, svcInternalURL := m.ApiPop.GetURL()
	api := &golang.API{
		Spec: m.ApiPop.GetSpec(),
		Tags: m.ApiPop.GetTags(),
		Type: m.ApiPop.GetType(),
		URL:  routeExternalURL,
	}
	if api.Annotations == nil {
		api.Annotations = map[string]string{}
	}
	if len(svcInternalURL) > 0 {
		api.Annotations[backstage.INTERNAL_SVC_URL] = svcInternalURL
	}
	if routeExternalURL != svcInternalURL && len(routeExternalURL) > 0 {
		api.Annotations[backstage.EXTERNAL_ROUTE_URL] = routeExternalURL
	}
	return api
}

func (m *ModelServerPopulator) GetOwner() string {
	owner := m.getStringPropVal(brdgtypes.Owner)
	if owner != nil {
		return util.SanitizeName(*owner)
	}
	return m.Owner
}

func (m *ModelServerPopulator) GetLifecycle() string {
	lifecycle := m.getStringPropVal(brdgtypes.Lifecycle)
	if lifecycle != nil {
		return *lifecycle
	}
	return m.Lifecycle
}

func (m *ModelServerPopulator) GetDescription() string {
	desc := ""
	d := m.getStringPropVal(brdgtypes.DescriptionKey)
	if d != nil {
		return *d
	}
	return desc
}

type ModelServerAPIPopulator struct {
	CommonSchemaPopulator
}

func (m *ModelServerAPIPopulator) getStringPropVal(key string) *string {
	return commonGetStringPropVal(key, m.InferSvc)
}

func (m *ModelServerAPIPopulator) GetSpec() string {
	ret := m.getStringPropVal(brdgtypes.APISpecKey)
	if ret == nil {
		return "TBD"
	}
	return *ret
}

func (m *ModelServerAPIPopulator) GetTags() []string {
	tags := []string{}
	for k, v := range m.InferSvc.Labels {
		tag := fmt.Sprintf("%s-%s", util.SanitizeName(k), util.SanitizeName(v))
		tags = append(tags, util.SanitizeName(tag))
	}
	return tags
}

func (m *ModelServerAPIPopulator) GetType() golang.Type {
	t := m.getStringPropVal(brdgtypes.APITypeKey)
	if t == nil {
		// assume open api
		return golang.Openapi
	}
	switch {
	case golang.Type(*t) == golang.Graphql:
		return golang.Graphql
	case golang.Type(*t) == golang.Asyncapi:
		return golang.Asyncapi
	case golang.Type(*t) == golang.Grpc:
		return golang.Grpc
	}
	return golang.Openapi
}

func (m *ModelServerAPIPopulator) getFullSvcURL() string {
	listOptions := &client.ListOptions{Namespace: m.InferSvc.Namespace}
	svcList := &corev1.ServiceList{}
	err := m.CtrlClient.List(m.Ctx, svcList, listOptions)
	if err != nil {
		return ""
	}
	for _, svc := range svcList.Items {
		if svc.OwnerReferences == nil {
			continue
		}
		for _, o := range svc.OwnerReferences {
			if o.Kind == "InferenceService" &&
				o.Name == m.InferSvc.Name &&
				strings.HasSuffix(svc.Name, "-predictor") {
				var port int32
				port = 0
				for _, sp := range svc.Spec.Ports {
					port = sp.Port
					if sp.TargetPort.Type == intstr.Int {
						port = sp.TargetPort.IntVal
					}
					break
				}
				portStr := ""
				if port != 0 && port != 80 {
					portStr = fmt.Sprintf(":%d", port)
				}
				return fmt.Sprintf("http://%s.%s.svc.cluster.local%s", svc.Name, svc.Namespace, portStr)
			}
		}
	}
	return ""

}

func (m *ModelServerAPIPopulator) GetURL() (string, string) {
	if m.InferSvc.Status.URL != nil && m.InferSvc.Status.URL.URL() != nil {
		// return the KServe InferenceService Route or Service URL
		kisUrl := m.InferSvc.Status.URL.URL().String()
		if strings.Contains(kisUrl, "svc.cluster.local") {
			svcURL := m.getFullSvcURL()
			return svcURL, svcURL
		}
		return m.InferSvc.Status.URL.URL().String(), m.getFullSvcURL()
	}
	return "", ""
}

type ModelPopulator struct {
	CommonSchemaPopulator
}

func (m *ModelPopulator) GetName() string {
	name := fmt.Sprintf("%s-%s", util.SanitizeName(m.InferSvc.Namespace), util.SanitizeName(m.InferSvc.Name))
	return util.SanitizeName(name)
}

func (m *ModelPopulator) GetOwner() string {
	owner := m.getStringPropVal(brdgtypes.Owner)
	if owner != nil {
		return util.SanitizeName(*owner)
	}
	return m.Owner
}

func (m *ModelPopulator) GetLifecycle() string {
	lifecycle := m.getStringPropVal(brdgtypes.Lifecycle)
	if lifecycle != nil {
		return util.SanitizeName(*lifecycle)
	}
	return m.Lifecycle
}

func (m *ModelPopulator) GetDescription() string {
	desc := ""
	d := m.getStringPropVal(brdgtypes.DescriptionKey)
	if d != nil {
		return *d
	}
	return desc
}

func (m *ModelPopulator) GetTags() []string {
	tags := []string{}
	for k, v := range m.InferSvc.Labels {
		tag := fmt.Sprintf("%s-%s", util.SanitizeName(k), util.SanitizeName(v))
		tags = append(tags, util.SanitizeName(tag))
	}
	return tags
}

func (m *ModelPopulator) GetArtifactLocationURL() *string {
	model := m.InferSvc.Spec.Predictor.Model
	if model != nil && model.StorageURI != nil {
		return model.StorageURI
	}
	if model != nil && model.Storage != nil && model.Storage.Path != nil {
		url := fmt.Sprintf("s3://%s", *model.Storage.Path)
		return &url
	}
	return nil
}

func (m *ModelPopulator) getStringPropVal(key string) *string {
	return commonGetStringPropVal(key, m.InferSvc)
}

func (m *ModelPopulator) GetEthics() *string {
	return m.getStringPropVal(brdgtypes.EthicsKey)
}

func (m *ModelPopulator) GetHowToUseURL() *string {
	return m.getStringPropVal(brdgtypes.HowToUseKey)
}

func (m *ModelPopulator) GetSupport() *string {
	return m.getStringPropVal(brdgtypes.SupportKey)
}

func (m *ModelPopulator) GetTraining() *string {
	return m.getStringPropVal(brdgtypes.TrainingKey)
}

func (m *ModelPopulator) GetUsage() *string {
	return m.getStringPropVal(brdgtypes.UsageKey)
}

func (m *ModelPopulator) GetLicense() *string {
	return m.getStringPropVal(brdgtypes.LicenseKey)
}

func (m *ModelPopulator) GetTechDocs() *string {
	techdocsUrl := m.getStringPropVal(brdgtypes.TechDocsKey)
	if techdocsUrl == nil && strings.Contains(m.GetName(), brdgtypes.Granite318bLabName) {
		granite31TechDocs := brdgtypes.Granite318bLabTechDocs
		return &granite31TechDocs
	} else if techdocsUrl != nil {
		u, err := url.Parse(*techdocsUrl)
		switch {
		case err != nil:
			fallthrough
		case u == nil:
			fallthrough
		case u != nil && (u.Scheme != "http" && u.Scheme != "https"):
			klog.Errorf("ignoring techdoc URL since there is either an error or bad scheme for techdoc url %v: err %v, url %v", techdocsUrl, err, u)
			return nil
		}
	}
	return techdocsUrl
}

type ModelCatalogPopulator struct {
	CommonSchemaPopulator
	MSPop *ModelServerPopulator
	MPops []*ModelPopulator
}

func (m *ModelCatalogPopulator) GetModels() []golang.Model {
	models := []golang.Model{}
	mPop := ModelPopulator{CommonSchemaPopulator: CommonSchemaPopulator{m.ComponentPopulator}}
	mPop.InferSvc = m.InferSvc
	m.MPops = append(m.MPops, &mPop)

	model := golang.Model{
		ArtifactLocationURL: mPop.GetArtifactLocationURL(),
		Description:         mPop.GetDescription(),
		Ethics:              mPop.GetEthics(),
		HowToUseURL:         mPop.GetHowToUseURL(),
		Lifecycle:           mPop.GetLifecycle(),
		Name:                mPop.GetName(),
		Owner:               mPop.GetOwner(),
		Support:             mPop.GetSupport(),
		Tags:                mPop.GetTags(),
		Training:            mPop.GetTraining(),
		Usage:               mPop.GetUsage(),
		License:             mPop.GetLicense(),
	}

	model.Annotations = make(map[string]string)
	// avoid namespace prefix
	modelNameAlreadySet := m.InferSvc.Annotations[backstage.MODEL_NAME]
	modelNameAlreadySet = strings.TrimSpace(modelNameAlreadySet)
	model.Annotations[backstage.MODEL_NAME] = modelNameAlreadySet
	if len(modelNameAlreadySet) == 0 {
		model.Annotations[backstage.MODEL_NAME] = m.InferSvc.Name
	}
	techDocsUrl := mPop.GetTechDocs()
	if techDocsUrl != nil && *techDocsUrl != "" {
		model.Annotations[brdgtypes.TechDocsKey] = *techDocsUrl
	}
	models = append(models, model)

	return models
}

func (m *ModelCatalogPopulator) GetModelServer() *golang.ModelServer {

	m.MSPop.InferSvc = m.InferSvc

	ms := &golang.ModelServer{
		API:            m.MSPop.GetAPI(),
		Authentication: m.MSPop.GetAuthentication(),
		Description:    m.MSPop.GetDescription(),
		HomepageURL:    m.MSPop.GetHomepageURL(),
		Lifecycle:      m.MSPop.GetLifecycle(),
		Name:           m.MSPop.GetName(),
		Owner:          m.MSPop.GetOwner(),
		Tags:           m.MSPop.GetTags(),
		Usage:          m.MSPop.GetUsage(),
	}
	if ms.Annotations == nil {
		ms.Annotations = map[string]string{}
	}
	return ms
}
