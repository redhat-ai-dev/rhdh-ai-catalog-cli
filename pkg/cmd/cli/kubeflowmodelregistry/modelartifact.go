package kubeflowmodelregistry

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/model-registry/pkg/openapi"
)

func (k *KubeFlowRESTClientWrapper) ListModelArtifacts(id string) ([]openapi.ModelArtifact, error) {
	buf, err := k.getFromModelRegistry(k.RootURL + fmt.Sprintf(LIST_ARTFIACTS_OFF_VERSIONS_URI, id))
	if err != nil {
		return nil, err
	}

	mas := openapi.ModelArtifactList{}
	err = json.Unmarshal(buf, &mas)
	if err != nil {
		return nil, err
	}
	return mas.Items, err
}

func (k *KubeFlowRESTClientWrapper) GetModelArtifact(id string) (*openapi.ModelArtifact, error) {
	buf, err := k.getFromModelRegistry(k.RootURL + fmt.Sprintf(GET_MODEL_ARTIFACT_URI, id))
	if err != nil {
		return nil, err
	}

	ma := openapi.ModelArtifact{}
	err = json.Unmarshal(buf, &ma)
	if err != nil {
		return nil, err
	}
	return &ma, err
}
