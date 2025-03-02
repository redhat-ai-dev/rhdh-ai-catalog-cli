package client

import (
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Artifacts) AddContent(key string, content []byte) error {
	restCfg, err := util.GetK8sConfig(a.cfg)
	if err != nil {
		return err
	}
	coreClient := util.GetCoreClient(restCfg)
	a.cm = &corev1.ConfigMap{}
	a.cm, err = coreClient.ConfigMaps(a.cfg.Namespace).Get(a.ctx, "bac-import-model", metav1.GetOptions{})
	notFound := false
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		notFound = true
		a.cm.Namespace = a.cfg.Namespace
		a.cm.Name = "bac-import-model"
	}
	if a.cm.BinaryData == nil {
		a.cm.BinaryData = map[string][]byte{}
	}
	// remember k8s CM keys can only contain alphanumerics and the '.', '-', and '_' symbols
	a.cm.BinaryData[key] = content
	if notFound {
		_, err = coreClient.ConfigMaps(a.cfg.Namespace).Create(a.ctx, a.cm, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	_, err = coreClient.ConfigMaps(a.cfg.Namespace).Update(a.ctx, a.cm, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
