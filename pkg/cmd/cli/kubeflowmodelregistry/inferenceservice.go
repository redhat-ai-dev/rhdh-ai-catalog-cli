package kubeflowmodelregistry

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/model-registry/pkg/openapi"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
)

func (k *KubeFlowRESTClientWrapper) ListInferenceServices() ([]openapi.InferenceService, error) {
	buf, err := k.getFromModelRegistry(k.RootURL + fmt.Sprintf(rest.LIST_INFERENCE_SERVICES_URI))
	if err != nil {
		return nil, err
	}

	mas := openapi.InferenceServiceList{}
	if len(buf) == 0 {
		return mas.Items, nil
	}
	err = json.Unmarshal(buf, &mas)
	if err != nil {
		return nil, err
	}
	return mas.Items, err
}
