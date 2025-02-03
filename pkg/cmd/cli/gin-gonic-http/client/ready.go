package client

import (
	"fmt"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (a *Artifacts) Ready() error {
	restCfg, err := util.GetK8sConfig(a.cfg)
	if err != nil {
		return err
	}
	routeClient := util.GetRouteClient(restCfg)

	watchTimeOut := int64(120)
	routeWatch, routeWatchErr := routeClient.Routes(a.cfg.Namespace).Watch(a.ctx, metav1.ListOptions{TimeoutSeconds: &watchTimeOut})
	if routeWatchErr != nil {
		return routeWatchErr
	}
	for event := range routeWatch.ResultChan() {
		switch event.Type {
		case watch.Error:
			return fmt.Errorf("route watch received error event with obj %#v", event.Object)
		case watch.Deleted:
			item := event.Object.(*routev1.Route)
			if item.Name != a.route.Name {
				continue
			}
			routeWatch.Stop()
			return fmt.Errorf("route watch detected our route %s was deleted", a.route.Name)
		}

		item := event.Object.(*routev1.Route)
		if item.Name != a.route.Name {
			continue
		}

		if len(item.Status.Ingress) == 0 {
			continue
		}
		if len(item.Status.Ingress[0].Conditions) == 0 {
			continue
		}

		if item.Status.Ingress[0].Conditions[0].Status == corev1.ConditionTrue {
			a.routeURL = "http://" + item.Status.Ingress[0].Host
			return nil
		}

	}
	return fmt.Errorf("route watch did not find desired condition in time: %s", a.route.String())
}
