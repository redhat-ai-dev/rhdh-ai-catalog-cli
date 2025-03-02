package kubeflowmodelregistry

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/model-registry/pkg/openapi"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/rest"
)

func (k *KubeFlowRESTClientWrapper) GetServingEnvironment(id string) (*openapi.ServingEnvironment, error) {
	buf, err := k.getFromModelRegistry(k.RootURL + fmt.Sprintf(rest.GET_SERVING_ENV_URI, id))
	if err != nil {
		return nil, err
	}

	se := &openapi.ServingEnvironment{}
	err = json.Unmarshal(buf, se)
	if err != nil {
		return nil, err
	}
	return se, err
}
