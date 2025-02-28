package client

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Artifacts) Delete() error {
	restCfg, err := util.GetK8sConfig(a.cfg)
	if err != nil {
		return err
	}

	appClient := util.GetAppClient(restCfg)
	name := "bac-import-model"
	err = appClient.Deployments(a.cfg.Namespace).Delete(a.ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	coreClient := util.GetCoreClient(restCfg)
	//err = coreClient.ConfigMaps(a.cfg.Namespace).Delete(a.ctx, name, metav1.DeleteOptions{})
	//if err != nil && !errors.IsNotFound(err) {
	//	return err
	//}

	err = coreClient.Services(a.cfg.Namespace).Delete(a.ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	routeClient := util.GetRouteClient(restCfg)
	err = routeClient.Routes(a.cfg.Namespace).Delete(a.ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}
