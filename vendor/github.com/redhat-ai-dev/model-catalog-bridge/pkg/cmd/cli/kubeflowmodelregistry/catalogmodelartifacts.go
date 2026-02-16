package kubeflowmodelregistry

import (
     "encoding/json"
     "fmt"

     "github.com/kubeflow/model-registry/catalog/pkg/openapi"
     "github.com/redhat-ai-dev/model-catalog-bridge/pkg/rest"
)

type CatalogModelArtifactsWrapper struct {
     artifacts []openapi.CatalogModelArtifact
}

func (k *KubeFlowRESTClientWrapper) ListCatalogModelArtifacts(sourceId, modelName string) ([]openapi.CatalogModelArtifact, error) {
     buf, err := k.getFromModelRegistry(k.RootCatalogURL + fmt.Sprintf(rest.LIST_CATALOG_MODEL_ARTIFACTS_URI, sourceId, modelName))
     if err != nil {
          return nil, err
     }

     cmal := openapi.CatalogModelArtifactList{}
     err = json.Unmarshal(buf, &cmal)
     if err != nil {
          return nil, err
     }
     return cmal.Items, nil
}