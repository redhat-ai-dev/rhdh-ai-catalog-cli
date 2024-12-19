package kubeflowmodelregistry

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/model-registry/pkg/openapi"
)

func (k *KubeFlowRESTClientWrapper) GetServingEnvironment(id string) (*openapi.ServingEnvironment, error) {
	buf, err := k.getFromModelRegistry(k.RootURL + fmt.Sprintf(GET_SERVING_ENV_URI, id))
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
