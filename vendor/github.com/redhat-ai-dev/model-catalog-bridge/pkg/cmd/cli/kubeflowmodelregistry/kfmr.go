package kubeflowmodelregistry

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	serverv1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/kubeflow/model-registry/pkg/openapi"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/cmd/cli/kserve"
	brdgtypes "github.com/redhat-ai-dev/model-catalog-bridge/pkg/types"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/util"
	"github.com/redhat-ai-dev/model-catalog-bridge/schema/types/golang"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (

	// pulled from makeValidator.ts in the catalog-model package in core backstage
	tagRegexp = "^[a-z0-9:+#]+(\\-[a-z0-9:+#]+)*$"
)

func LoopOverKFMR(ids []string, kfmr *KubeFlowRESTClientWrapper) ([]openapi.RegisteredModel, map[string][]openapi.ModelVersion, map[string]map[string][]openapi.ModelArtifact, error) {
	var err error
	rmArray := []openapi.RegisteredModel{}
	mvsMap := map[string][]openapi.ModelVersion{}
	masMap := map[string]map[string][]openapi.ModelArtifact{}

	if len(ids) == 0 {
		var rms []openapi.RegisteredModel
		rms, err = kfmr.ListRegisteredModels()
		if err != nil {
			klog.Errorf("list registered models error: %s", err.Error())
			klog.Flush()
			return nil, nil, nil, err
		}
		for _, rm := range rms {
			if rm.State != nil && *rm.State == openapi.REGISTEREDMODELSTATE_ARCHIVED {
				klog.V(4).Infof("LoopOverKFMR skipping archived registered model %s", rm.Name)
				continue
			}
			var mvs []openapi.ModelVersion
			var mas map[string][]openapi.ModelArtifact
			mvs, mas, err = callKubeflowREST(*rm.Id, kfmr)
			if err != nil {
				klog.Errorf("%s", err.Error())
				klog.Flush()
				return nil, nil, nil, err
			}

			rmArray = append(rmArray, rm)
			mvsMap[util.SanitizeName(rm.Name)] = mvs
			masMap[util.SanitizeName(rm.Name)] = mas
		}
	} else {
		for _, id := range ids {
			var rm *openapi.RegisteredModel
			rm, err = kfmr.GetRegisteredModel(id)
			if err != nil {
				klog.Errorf("get registered model error for %s: %s", id, err.Error())
				klog.Flush()
				return nil, nil, nil, err
			}
			if rm.State != nil && *rm.State == openapi.REGISTEREDMODELSTATE_ARCHIVED {
				klog.V(4).Infof("LoopOverKFMR skipping archived registered model %s", rm.Name)
				continue
			}
			var mvs []openapi.ModelVersion
			var mas map[string][]openapi.ModelArtifact
			mvs, mas, err = callKubeflowREST(*rm.Id, kfmr)
			if err != nil {
				klog.Errorf("get model version/artifact error for %s: %s", id, err.Error())
				klog.Flush()
				return nil, nil, nil, err
			}
			rmArray = append(rmArray, *rm)
			mvsMap[util.SanitizeName(rm.Name)] = mvs
			masMap[util.SanitizeName(rm.Name)] = mas
		}
	}
	return rmArray, mvsMap, masMap, nil
}

func GetKubeFlowInferenceServicesForModelVersion(kfrm *KubeFlowRESTClientWrapper,
	mv *openapi.ModelVersion) ([]openapi.InferenceService, error) {
	isl, err := kfrm.ListInferenceServices()
	if err != nil {
		return nil, err
	}
	// only include inference services that correspond to this model version
	mvISL := []openapi.InferenceService{}
	for _, is := range isl {
		if is.GetModelVersionId() == mv.GetId() && is.ModelVersionId != nil {
			mvISL = append(mvISL, is)
		}
	}
	klog.V(4).Infof("getKubeFlowInferenceServicesForModelVersion total num kubeflow infsvc %d num matched to model version %d", len(isl), len(mvISL))
	return mvISL, nil
}

func callKubeflowREST(id string, kfmr *KubeFlowRESTClientWrapper) ([]openapi.ModelVersion, map[string][]openapi.ModelArtifact, error) {
	finalMVS := []openapi.ModelVersion{}
	mvs, err := kfmr.ListModelVersions(id)
	if err != nil {
		klog.Errorf("ERROR: error list model versions for %s: %s", id, err.Error())
		return nil, nil, err
	}
	ma := map[string][]openapi.ModelArtifact{}
	for _, mv := range mvs {
		if mv.State != nil && *mv.State == openapi.MODELVERSIONSTATE_ARCHIVED {
			klog.V(4).Infof("callKubeflowREST skipping archived model version %s", mv.Name)
			continue
		}
		finalMVS = append(finalMVS, mv)
		var v []openapi.ModelArtifact
		v, err = kfmr.ListModelArtifacts(*mv.Id)
		if err != nil {
			klog.Errorf("ERROR error list model artifacts for %s:%s: %s", id, *mv.Id, err.Error())
			return finalMVS, ma, err
		}
		if len(v) == 0 {
			v, err = kfmr.ListModelArtifacts(id)
			if err != nil {
				klog.Errorf("ERROR error list model artifacts for %s:%s: %s", id, *mv.Id, err.Error())
				return finalMVS, ma, err
			}
		}
		ma[*mv.Id] = v
	}
	return finalMVS, ma, nil
}

func getTagsFromCustomProps(lastMod bool, props map[string]openapi.MetadataValue) map[string]string {
	tags := map[string]string{}
	regex, _ := regexp.Compile(tagRegexp)
	for cpk, cpv := range props {
		switch {
		case cpk == brdgtypes.LicenseKey:
			fallthrough
		case cpk == brdgtypes.TechDocsKey:
			klog.V(4).Infof("Skip adding TechDocs or License key to tags")
		case cpk == brdgtypes.RHOAIModelCatalogSourceModelVersion:
			fallthrough
		case cpk == brdgtypes.RHOAIModelCatalogSourceModelKey:
			fallthrough
		case cpk == brdgtypes.RHOAIModelCatalogRegisteredFromKey:
			fallthrough
		case cpk == brdgtypes.RHOAIModelCatalogProviderKey:
			fallthrough
		case cpk == brdgtypes.APITypeKey:
			fallthrough
		case cpk == brdgtypes.RHOAIModelRegistryRegisteredFromCatalogRepositoryName:
			v := ""
			if cpv.MetadataStringValue != nil {
				v = strings.ToLower(cpv.MetadataStringValue.StringValue)
			}
			if len(v) > 0 && regex.MatchString(v) && len(v) <= 63 {
				tags[cpk] = v
			}
		case cpk == brdgtypes.RHOAIModelRegistryLastModified && lastMod:
			v := ""
			replacerColon := strings.NewReplacer(":", "-")
			replacerDot := strings.NewReplacer(".", "-")
			replacerT := strings.NewReplacer("T", "-")
			replacerZ := strings.NewReplacer("Z", "")
			if cpv.MetadataStringValue != nil {
				v = replacerColon.Replace(cpv.MetadataStringValue.StringValue)
				v = replacerDot.Replace(v)
				v = replacerT.Replace(v)
				v = replacerZ.Replace(v)
				v = fmt.Sprintf("last-modified-time-%s", v)
			}
			if len(v) > 0 && regex.MatchString(v) && len(v) <= 63 {
				v = strings.ToLower(v)
				tags[cpk] = v
			}
		default:
			v := cpk
			if cpv.MetadataStringValue != nil && len(cpv.MetadataStringValue.StringValue) > 0 {
				v = v + "-" + strings.ToLower(cpv.MetadataStringValue.StringValue)
			}
			if len(v) > 0 && regex.MatchString(v) && len(v) <= 63 {
				tags[cpk] = v
			}
		}
	}
	return tags
}

func commonGetStringPropVal(key string, mvIndex int, mv *openapi.ModelVersion, rm *openapi.RegisteredModel) *string {
	var vmap map[string]openapi.MetadataValue
	var retString *string

	if mv.HasCustomProperties() {
		vmap = mv.GetCustomProperties()
		retString = innerGetStringPropVal(key, &vmap)
		if retString != nil {
			return retString
		}
	}

	if rm.HasCustomProperties() {
		vmap = rm.GetCustomProperties()
		retString = innerGetStringPropVal(key, &vmap)
	}
	return retString
}

func innerGetStringPropVal(key string, vmap *map[string]openapi.MetadataValue) *string {
	v, ok := (*vmap)[key]
	if !ok {
		return nil
	}

	if v.MetadataStringValue != nil {
		return &v.MetadataStringValue.StringValue
	}
	return nil
}

// json array schema populator

type CommonSchemaPopulator struct {
	// reuse the component populator as it houses all the KFMR artifacts of noew
	ComponentPopulator
}

type ModelCatalogPopulator struct {
	CommonSchemaPopulator
	MSPop *ModelServerPopulator
	MPops []*ModelPopulator
}

func (m *ModelCatalogPopulator) GetModels() []golang.Model {
	models := []golang.Model{}
	mPop := ModelPopulator{CommonSchemaPopulator: CommonSchemaPopulator{m.ComponentPopulator}}
	m.MPops = append(m.MPops, &mPop)
	for maidx, ma := range m.ModelArtifacts {
		if ma.GetId() == m.RegisteredModel.GetId() {
			mPop.MAIndex = maidx
			break
		}
	}

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

	if model.Annotations == nil {
		model.Annotations = make(map[string]string)
	}

	model.Annotations[backstage.MODEL_NAME] = model.Name
	// see if there are overriddes in the custom props
	props := mPop.RegisteredModel.GetCustomProperties()
	modelName, ok := props[backstage.MODEL_NAME]
	if ok && modelName.MetadataStringValue != nil {
		model.Annotations[backstage.MODEL_NAME] = modelName.MetadataStringValue.StringValue
	}
	props = mPop.ModelVersion.GetCustomProperties()
	modelName, ok = props[backstage.MODEL_NAME]
	if ok && modelName.MetadataStringValue != nil {
		model.Annotations[backstage.MODEL_NAME] = modelName.MetadataStringValue.StringValue
	}

	techDocsUrl := mPop.GetTechDocs()
	if techDocsUrl != nil && *techDocsUrl != "" {
		model.Annotations[brdgtypes.TechDocsKey] = *techDocsUrl
	}
	models = append(models, model)

	return models
}

func (m *ModelCatalogPopulator) GetModelServer() *golang.ModelServer {
	klog.V(4).Infof("kfmr:GetModelServer kubeflow infsvc found %v kserve infsvc found %v", m.InferenceService != nil, m.Kis != nil)

	if m.InferenceService == nil && m.Kis == nil {
		m.Kis = m.GetInferenceServerByRegModelModelVersionName()
		if m.Kis == nil {
			return nil
		}
	}

	foundInferenceService := false
	if m.InferenceService != nil {
		klog.V(4).Infof("kfmr:GetModelServer m.Kis != nil %v mv.RegisteredModelId == m.RegisteredModel.GetId() %v mv.GetId() == kfmrIS.GetModelVersionId() %v mv.RegisteredModelId == m.RegisteredModel.GetId() %v mv.GetId() == kfmrIS.GetModelVersionId() %v",
			m.Kis != nil, m.ModelVersion.RegisteredModelId == m.RegisteredModel.GetId(), m.ModelVersion.GetId() == m.InferenceService.GetModelVersionId(), m.ModelVersion.RegisteredModelId == m.RegisteredModel.GetId(), m.ModelVersion.GetId() == m.InferenceService.GetModelVersionId())
	} else {
		klog.V(4).Infof("kfmr:GetModelServer m.Kis != nil %v mv.RegisteredModelId == m.RegisteredModel.GetId() %v mv.RegisteredModelId == m.RegisteredModel.GetId() %v",
			m.Kis != nil, m.ModelVersion.RegisteredModelId == m.RegisteredModel.GetId(), m.ModelVersion.RegisteredModelId == m.RegisteredModel.GetId())
	}
	switch {
	// in case kubeflow/kserve reconciliation is not working
	case m.Kis != nil && util.KServeInferenceServiceMapping(m.RegisteredModel.GetId(), m.ModelVersion.GetId(), m.Kis):
		fallthrough
	case m.ModelVersion.RegisteredModelId == m.RegisteredModel.GetId() && m.InferenceService != nil && m.ModelVersion.GetId() == m.InferenceService.GetModelVersionId():
		foundInferenceService = true
		break
	}

	klog.V(4).Infof("kfmr:GetModelServer found infsvcs %v", foundInferenceService)
	if !foundInferenceService {
		return nil
	}

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
		ms.Annotations = make(map[string]string)
	}
	return ms
}

type ModelPopulator struct {
	CommonSchemaPopulator
	MVIndex int
	MAIndex int
}

func (m *ModelPopulator) GetName() string {
	return util.SanitizeName(m.RegisteredModel.Name) + "-" + util.SanitizeModelVersion(m.ModelVersion.GetName())
}

func (m *ModelPopulator) GetOwner() string {
	owner := m.getStringPropVal(brdgtypes.Owner)
	if owner != nil {
		return util.SanitizeName(*owner)
	}
	if m.RegisteredModel.Owner != nil {
		return util.SanitizeName(*m.RegisteredModel.Owner)
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
	return m.RegisteredModel.GetDescription() + "\n" + m.ModelVersion.GetDescription()
}

func (m *ModelPopulator) GetTags() []string {
	tags := getTagsFromCustomProps(false, m.RegisteredModel.GetCustomProperties())
	if m.ModelVersion.HasCustomProperties() {
		tagsMV := getTagsFromCustomProps(true, m.ModelVersion.GetCustomProperties())
		for k, v := range tagsMV {
			tags[k] = v
		}
	}
	// any MA custom props will be user defined so just add
	for _, ma := range m.ModelArtifacts {
		if ma.HasCustomProperties() {
			tagsMA := getTagsFromCustomProps(true, ma.GetCustomProperties())
			for k, v := range tagsMA {
				tags[k] = v
			}
		}
	}

	finalTags := []string{}
	for _, v := range tags {
		finalTags = append(finalTags, v)
	}
	return finalTags
}

func (m *ModelPopulator) GetArtifactLocationURL() *string {
	for _, ma := range m.ModelArtifacts {
		return ma.Uri
	}
	return nil
}

func (m *ModelPopulator) getStringPropVal(key string) *string {
	return commonGetStringPropVal(key, m.MVIndex, m.ModelVersion, m.RegisteredModel)
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
	replacer := strings.NewReplacer(" ", "")

	if techdocsUrl != nil {
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
	for _, ma := range m.ModelArtifacts {
		s := fmt.Sprintf("http://localhost:9090%s?key=%s", util.ModelCardURI, replacer.Replace(ma.GetModelSourceClass())+replacer.Replace(ma.GetModelSourceGroup())+replacer.Replace(ma.GetModelSourceName()))
		techdocsUrl = &s
		break
	}

	return techdocsUrl
}

type ModelServerPopulator struct {
	CommonSchemaPopulator
	ApiPop  ModelServerAPIPopulator
	MVIndex int
	MAIndex int
}

func (m *ModelServerPopulator) getStringPropVal(key string) *string {
	return commonGetStringPropVal(key, m.MVIndex, m.ModelVersion, m.RegisteredModel)
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

	if m.Kis == nil {
		m.Kis = m.GetInferenceServerByRegModelModelVersionName()
		if m.Kis == nil {
			return &auth
		}
	}
	listOptions := &client.ListOptions{Namespace: m.Kis.Namespace}
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
				o.Name == m.Kis.Name {
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
	if m.InferenceService != nil {
		sanitizedName := util.SanitizeName(m.InferenceService.GetName())
		return sanitizedName
	}
	// if kubeflow/kserve reconciliation is not working, let's use the kserve inference service name
	if m.Kis != nil {
		return util.SanitizeName(m.Kis.Name)
	}
	return ""
}

func (m *ModelServerPopulator) GetTags() []string {
	tags := getTagsFromCustomProps(false, m.RegisteredModel.GetCustomProperties())
	if m.ModelVersion.HasCustomProperties() {
		tagsMV := getTagsFromCustomProps(true, m.ModelVersion.GetCustomProperties())
		for k, v := range tagsMV {
			tags[k] = v
		}
	}
	// any MA custom props will be user defined so just add
	for _, ma := range m.ModelArtifacts {
		if ma.HasCustomProperties() {
			tagsMA := getTagsFromCustomProps(true, ma.GetCustomProperties())
			for k, v := range tagsMA {
				tags[k] = v
			}
		}
	}

	finalTags := []string{}
	for _, v := range tags {
		finalTags = append(finalTags, v)
	}
	return finalTags
}

func (m *ModelServerPopulator) GetAPI() *golang.API {
	m.ApiPop.MVIndex = m.MVIndex
	m.ApiPop.MAIndex = m.MAIndex
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
		// model exposed via route
		api.Annotations[backstage.EXTERNAL_ROUTE_URL] = routeExternalURL

	}
	return api
}

func (m *ModelServerPopulator) GetOwner() string {
	owner := m.getStringPropVal(brdgtypes.Owner)
	if owner != nil {
		return util.SanitizeName(*owner)
	}
	if m.RegisteredModel.Owner != nil {
		return util.SanitizeName(*m.RegisteredModel.Owner)
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
	return m.RegisteredModel.GetDescription() + "\n" + m.ModelVersion.GetDescription()
}

type ModelServerAPIPopulator struct {
	CommonSchemaPopulator
	MVIndex int
	MAIndex int
}

func (m *ModelServerAPIPopulator) getStringPropVal(key string) *string {
	return commonGetStringPropVal(key, m.MVIndex, m.ModelVersion, m.RegisteredModel)
}

func (m *ModelServerAPIPopulator) GetSpec() string {
	ret := m.getStringPropVal(brdgtypes.APISpecKey)
	if ret == nil {
		return "TBD"
	}
	return *ret
}

func (m *ModelServerAPIPopulator) GetTags() []string {
	tags := getTagsFromCustomProps(false, m.RegisteredModel.GetCustomProperties())
	if m.ModelVersion.HasCustomProperties() {
		tagsMV := getTagsFromCustomProps(true, m.ModelVersion.GetCustomProperties())
		for k, v := range tagsMV {
			tags[k] = v
		}
	}
	// any MA custom props will be user defined so just add
	for _, ma := range m.ModelArtifacts {
		if ma.HasCustomProperties() {
			tagsMA := getTagsFromCustomProps(true, ma.GetCustomProperties())
			for k, v := range tagsMA {
				tags[k] = v
			}
		}
	}

	finalTags := []string{}
	for _, v := range tags {
		finalTags = append(finalTags, v)
	}
	return finalTags
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
	// prior testing with chatbot confirmed we needed to add the target port to the service URL if the port is 80
	// and the target port is 8080; otherwise, if the port itself is 8080, the odh/rhoai consoles seem the append
	// the port correctly; so we will find the corresponding service and add the port
	listOptions := &client.ListOptions{Namespace: m.Kis.Namespace}
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
				o.Name == m.Kis.Name &&
				strings.HasSuffix(svc.Name, "-predictor") {
				// prior testing with chatbot confirmed we needed to add the target port to the service URL if the port is 80
				// and the target port is 8080; otherwise, if the port itself is 8080, the odh/rhoai consoles seem the append
				// the port correctly
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
	if m.Kis == nil {
		m.Kis = m.GetInferenceServerByRegModelModelVersionName()
		if m.Kis == nil {
			return "", ""
		}
	}
	if m.Kis.Status.URL != nil && m.Kis.Status.URL.URL() != nil {
		kisUrl := m.Kis.Status.URL.URL().String()
		if strings.Contains(kisUrl, "svc.cluster.local") {
			// only the service URL is exposed
			svcURL := m.getFullSvcURL()
			return svcURL, svcURL
		}
		// return the KServe InferenceService Route and Service URL
		return m.Kis.Status.URL.URL().String(), m.getFullSvcURL()
	}

	return "", ""
}

// catalog-info.yaml populators

func CallBackstagePrinters(ctx context.Context, owner, lifecycle string, rm *openapi.RegisteredModel, mvs *openapi.ModelVersion, mas []openapi.ModelArtifact, isl *openapi.InferenceService, is *serverv1beta1.InferenceService, kfmr *KubeFlowRESTClientWrapper, client client.Client, writer io.Writer, format brdgtypes.NormalizerFormat) error {
	compPop := ComponentPopulator{}
	compPop.Owner = owner
	compPop.Lifecycle = lifecycle
	compPop.Kfmr = kfmr
	compPop.RegisteredModel = rm
	compPop.ModelVersion = mvs
	compPop.ModelArtifacts = mas
	compPop.InferenceService = isl
	compPop.Kis = is
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
		fallthrough
	default:
		err := backstage.PrintComponent(&compPop, writer)
		if err != nil {
			return err
		}

		resPop := ResourcePopulator{}
		resPop.Owner = owner
		resPop.Lifecycle = lifecycle
		resPop.Kfmr = kfmr
		resPop.RegisteredModel = rm
		resPop.ModelVersion = mvs
		resPop.Kis = is
		resPop.CtrlClient = client
		resPop.Ctx = ctx
		resPop.ModelVersion = mvs
		resPop.ModelArtifacts = mas
		err = backstage.PrintResource(&resPop, writer)
		if err != nil {
			return err
		}

		apiPop := ApiPopulator{}
		apiPop.Owner = owner
		apiPop.Lifecycle = lifecycle
		apiPop.Kfmr = kfmr
		apiPop.RegisteredModel = rm
		apiPop.ModelVersion = mvs
		apiPop.InferenceService = isl
		apiPop.Kis = is
		apiPop.CtrlClient = client
		apiPop.Ctx = ctx
		return backstage.PrintAPI(&apiPop, writer)
	}

	return nil

}

type CommonPopulator struct {
	Owner            string
	Lifecycle        string
	RegisteredModel  *openapi.RegisteredModel
	ModelVersion     *openapi.ModelVersion
	InferenceService *openapi.InferenceService
	Kfmr             *KubeFlowRESTClientWrapper
	Kis              *serverv1beta1.InferenceService
	CtrlClient       client.Client
	Ctx              context.Context
}

func (pop *CommonPopulator) GetOwner() string {
	if len(pop.Owner) != 0 {
		return pop.Owner
	}
	if pop.RegisteredModel.Owner != nil {
		return util.SanitizeName(*pop.RegisteredModel.Owner)
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

func (pop *CommonPopulator) GetProvidedAPIs() []string {
	return []string{}
}

type ComponentPopulator struct {
	CommonPopulator
	ModelArtifacts []openapi.ModelArtifact
}

func (pop *ComponentPopulator) GetName() string {
	return util.SanitizeName(pop.RegisteredModel.Name)
}

func (pop *ComponentPopulator) GetLinks() []backstage.EntityLink {
	links := pop.GetLinksFromInferenceServices()
	//TODO maybe multi resource / multi model indication
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

func (pop *CommonPopulator) GetInferenceServerByRegModelModelVersionName() *serverv1beta1.InferenceService {
	iss := []serverv1beta1.InferenceService{}
	switch {
	case pop.CtrlClient != nil:
		isList := &serverv1beta1.InferenceServiceList{}
		err := pop.CtrlClient.List(pop.Ctx, isList)
		if err != nil {
			klog.Errorf("getLinksFromInferenceServices list all inferenceservices error: %s", err.Error())
			return nil
		}
		iss = append(iss, isList.Items...)

	case pop.Kfmr != nil && pop.Kfmr.Config != nil && pop.Kfmr.Config.ServingClient != nil:
		isList, err := pop.Kfmr.Config.ServingClient.InferenceServices(metav1.NamespaceAll).List(pop.Ctx, metav1.ListOptions{})
		if err != nil {
			klog.Errorf("getLinksFromInferenceServices list all inferenceservices error: %s", err.Error())
			return nil
		}
		if isList != nil {
			iss = append(iss, isList.Items...)
		}
	}
	klog.V(4).Infof("commonPop:GetInferenceServerByRegModelModelVersionName found %d kserve infsvcs", len(iss))
	for _, is := range iss {
		if util.KServeInferenceServiceMapping(pop.RegisteredModel.GetId(), pop.ModelVersion.GetId(), &is) {
			return &is
		}
	}

	return nil
}

func (pop *CommonPopulator) GetLinksFromInferenceServices() []backstage.EntityLink {
	links := []backstage.EntityLink{}
	// if for some reason kserve/kubeflow reconciliation is not working and there are no kubeflow inference services,
	// let's match up based on registered model / model version name
	if pop.InferenceService == nil {
		if pop.Kis != nil {
			kpop := kserve.CommonPopulator{InferSvc: pop.Kis}
			links = append(links, kpop.GetLinks()...)
			return links
		}
		pop.Kis = pop.GetInferenceServerByRegModelModelVersionName()
		if pop.Kis != nil {
			kpop := kserve.CommonPopulator{InferSvc: pop.Kis}
			links = append(links, kpop.GetLinks()...)
			return links
		}
		return []backstage.EntityLink{}
	}

	rmid := pop.RegisteredModel.GetId()
	if pop.InferenceService.RegisteredModelId != rmid {
		return []backstage.EntityLink{}
	}
	var iss *openapi.InferenceServiceState
	ok := false
	iss, ok = pop.InferenceService.GetDesiredStateOk()
	if !ok {
		return []backstage.EntityLink{}
	}
	if *iss != openapi.INFERENCESERVICESTATE_DEPLOYED {
		return []backstage.EntityLink{}
	}
	se, err := pop.Kfmr.GetServingEnvironment(pop.InferenceService.ServingEnvironmentId)
	if err != nil {
		klog.Errorf("ComponentPopulator GetLinks: %s", err.Error())
		return []backstage.EntityLink{}
	}
	if pop.Kis == nil {
		kisns := se.GetName()
		kisnm := pop.InferenceService.GetRuntime()
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
			return []backstage.EntityLink{}
		}

		pop.Kis = kis
	}
	kpop := kserve.CommonPopulator{InferSvc: pop.Kis}
	links = append(links, kpop.GetLinks()...)

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
	depends = append(depends, "resource:"+pop.ModelVersion.Name)

	for _, ma := range pop.ModelArtifacts {
		depends = append(depends, "api:"+*ma.Name)
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
	//TODO maybe multi resource / multi model indication
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
	return []string{fmt.Sprintf("component:%s", util.SanitizeName(pop.RegisteredModel.Name))}
}

func (pop *ResourcePopulator) GetDisplayName() string {
	return pop.GetName()
}

type ApiPopulator struct {
	CommonPopulator
}

func (pop *ApiPopulator) GetName() string {
	return util.SanitizeName(pop.RegisteredModel.Name)
}

func (pop *ApiPopulator) GetDependencyOf() []string {
	return []string{fmt.Sprintf("component:%s", util.SanitizeName(pop.RegisteredModel.Name))}
}

func (pop *ApiPopulator) GetDefinition() string {
	// definition must be set to something to pass backstage validation
	return "no-definition-yet"
}

func (pop *ApiPopulator) GetTechdocRef() string {
	// TODO in theory the Kfmr modelcard support when it arrives will replace this
	return "api/"
}

func (pop *ApiPopulator) GetTags() []string {
	return []string{}
}

func (pop *ApiPopulator) GetLinks() []backstage.EntityLink {
	return pop.GetLinksFromInferenceServices()
}

func (pop *ApiPopulator) GetDisplayName() string {
	return pop.GetName()
}
