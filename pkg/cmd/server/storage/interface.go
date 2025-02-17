package storage

import "k8s.io/client-go/rest"

type BridgeStorage interface {
	Initialize(cfg *rest.Config)
	Fetch(key string) (any, error)
	Upsert(key string, value any) error
	Remove(key string) error
}

type BridgeStorageType string

const ConfigMapBridgeStorage BridgeStorageType = "ConfigMap"
const GithubBridgeStorage BridgeStorageType = "Github"

func NewBridgeStorage(storageType BridgeStorageType) BridgeStorage {
	switch storageType {
	case ConfigMapBridgeStorage:
	case GithubBridgeStorage:
	}
	return nil
}
