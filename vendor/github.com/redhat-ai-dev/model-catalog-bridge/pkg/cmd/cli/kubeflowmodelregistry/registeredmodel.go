package kubeflowmodelregistry

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/model-registry/pkg/openapi"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/rest"
)

func (k *KubeFlowRESTClientWrapper) ListRegisteredModels() ([]openapi.RegisteredModel, error) {
	buf, err := k.getFromModelRegistry(k.RootURL + rest.LIST_REG_MODEL_URI)
	if err != nil {
		return nil, err
	}

	rmList := openapi.RegisteredModelList{}
	err = json.Unmarshal(buf, &rmList)
	if err != nil {
		return nil, err
	}
	return rmList.Items, err
}

func (k *KubeFlowRESTClientWrapper) GetRegisteredModel(registeredModelID string) (*openapi.RegisteredModel, error) {
	buf, err := k.getFromModelRegistry(k.RootURL + fmt.Sprintf(rest.GET_REG_MODEL_URI, registeredModelID))
	if err != nil {
		return nil, err
	}

	rm := openapi.RegisteredModel{}
	err = json.Unmarshal(buf, &rm)
	if err != nil {
		return nil, err
	}
	return &rm, err
}
