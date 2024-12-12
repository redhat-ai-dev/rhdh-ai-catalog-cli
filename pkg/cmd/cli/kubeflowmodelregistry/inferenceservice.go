package kubeflowmodelregistry

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/model-registry/pkg/openapi"
)

func (k *KubeFlowRESTClientWrapper) ListInferenceServices() ([]openapi.InferenceService, error) {
	buf, err := k.getFromModelRegistry(k.RootURL + fmt.Sprintf(LIST_INFERENCE_SERVICES_URI))
	if err != nil {
		return nil, err
	}

	mas := openapi.InferenceServiceList{}
	err = json.Unmarshal(buf, &mas)
	if err != nil {
		return nil, err
	}
	return mas.Items, err
}
