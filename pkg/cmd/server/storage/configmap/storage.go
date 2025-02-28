package configmap

import (
	"context"
	"encoding/json"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/types"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/rest"
)

type ConfigMapBridgeStorage struct {
	cfg *rest.Config
	cl  corev1client.CoreV1Interface
	ns  string
}

func (c *ConfigMapBridgeStorage) Initialize(cfg *rest.Config) error {
	c.cfg = cfg
	c.cl = util.GetCoreClient(c.cfg)
	c.ns = util.GetCurrentProject()
	_, err := c.cl.ConfigMaps(c.ns).Get(context.Background(), "bac-import-model", metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if err != nil {
		cm := &corev1.ConfigMap{}
		cm.Name = "bac-import-model"
		_, err = c.cl.ConfigMaps(c.ns).Create(context.Background(), cm, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ConfigMapBridgeStorage) Upsert(key string, value types.StorageBody) error {
	cm, err := c.cl.ConfigMaps(c.ns).Get(context.Background(), "bac-import-model", metav1.GetOptions{})
	if err != nil {
		return err
	}
	if cm.BinaryData == nil {
		cm.BinaryData = map[string][]byte{}
	}
	buf := []byte{}
	buf, err = json.Marshal(value)
	if err != nil {
		return err
	}
	cm.BinaryData[key] = buf
	_, err = c.cl.ConfigMaps(c.ns).Update(context.Background(), cm, metav1.UpdateOptions{})
	return err
}

func (c *ConfigMapBridgeStorage) Fetch(key string) (types.StorageBody, error) {
	cm, err := c.cl.ConfigMaps(c.ns).Get(context.Background(), "bac-import-model", metav1.GetOptions{})
	sb := types.StorageBody{}
	if err != nil {
		return sb, err
	}
	if cm.BinaryData == nil {
		return sb, nil
	}

	buf, ok := cm.BinaryData[key]
	if !ok {
		return sb, nil
	}
	err = json.Unmarshal(buf, &sb)
	return sb, err
}

func (c *ConfigMapBridgeStorage) Remove(key string) error {
	cm, err := c.cl.ConfigMaps(c.ns).Get(context.Background(), "bac-import-model", metav1.GetOptions{})
	if err != nil {
		return err
	}
	if cm.BinaryData == nil {
		return nil
	}
	delete(cm.BinaryData, key)
	_, err = c.cl.ConfigMaps(c.ns).Update(context.Background(), cm, metav1.UpdateOptions{})
	return err
}
