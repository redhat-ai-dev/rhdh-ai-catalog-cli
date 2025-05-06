package types

import "k8s.io/client-go/rest"

type BridgeStorage interface {
	Initialize(cfg *rest.Config) error
	Fetch(key string) (StorageBody, error)
	Upsert(key string, value StorageBody) error
	Remove(key string) error
	List() ([]string, error)
}

type StorageBody struct {
	Body            []byte `json:"body"`
	LocationId      string `json:"locationId"`
	LocationTarget  string `json:"locationTarget"`
	LocationIDValid bool   `json:"locationIDValid"`
}

type BridgeStorageType string

const (
	ConfigMapBridgeStorage BridgeStorageType = "ConfigMap"
	GithubBridgeStorage    BridgeStorageType = "Github"

	StorageUrlEnvVar  = "STORAGE_URL"
	StorageTypeEnvVar = "STORAGE_TYPE"

	PushToRHDHEnvVar = "PUSH_TO_RHDH"
)
