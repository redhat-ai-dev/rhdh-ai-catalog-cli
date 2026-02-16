package kubeflowmodelregistry

import (
     "encoding/json"

     "github.com/kubeflow/model-registry/catalog/pkg/openapi"
     "github.com/redhat-ai-dev/model-catalog-bridge/pkg/rest"
)

func (k *KubeFlowRESTClientWrapper) ListCatalogSources() ([]openapi.CatalogSource, error) {
     buf, err := k.getFromModelRegistry(k.RootCatalogURL + rest.LIST_CATALOG_SOURECES_URI)
     if err != nil {
          return nil, err
     }

     csl := openapi.CatalogSourceList{}
     err = json.Unmarshal(buf, &csl)
     if err != nil {
          return nil, err
     }
     return csl.Items, nil
}