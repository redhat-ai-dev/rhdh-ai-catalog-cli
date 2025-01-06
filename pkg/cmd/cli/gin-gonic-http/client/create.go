package client

import (
     "github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
     metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Artifacts) Create() error {
     restCfg, err := util.GetK8sConfig(a.cfg)
     if err != nil {
          return err
     }
     coreClient := util.GetCoreClient(restCfg)
     a.cm, err = coreClient.ConfigMaps(a.cfg.Namespace).Create(a.ctx, a.cm, metav1.CreateOptions{})
     if err != nil {
          return err
     }

     a.svc, err = coreClient.Services(a.cfg.Namespace).Create(a.ctx, a.svc, metav1.CreateOptions{})
     if err != nil {
          return err
     }

     routeClient := util.GetRouteClient(restCfg)
     a.route, err = routeClient.Routes(a.cfg.Namespace).Create(a.ctx, a.route, metav1.CreateOptions{})
     if err != nil {
          return err
     }

     appClient := util.GetAppClient(restCfg)
     a.dpm, err = appClient.Deployments(a.cfg.Namespace).Create(a.ctx, a.dpm, metav1.CreateOptions{})
     if err != nil {
          return err
     }
     return nil
}
