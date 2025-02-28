package storage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/server/storage/configmap"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/types"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
)

func NewBridgeStorage(storageType types.BridgeStorageType) types.BridgeStorage {
	switch storageType {
	case types.ConfigMapBridgeStorage:
		st := configmap.ConfigMapBridgeStorage{}
		cfg, err := util.GetK8sConfig(&config.Config{})
		if err != nil {
			return nil
		}
		st.Initialize(cfg)
		return &st
	case types.GithubBridgeStorage:
	}
	return nil
}
