package client

import (
	"context"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/config"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TODO this entire bit of the POC most likely can be removed, and then revendor / client up the CLI repo
type Artifacts struct {
	ctx      context.Context
	cfg      *config.Config
	cm       *corev1.ConfigMap
	svc      *corev1.Service
	route    *routev1.Route
	dpm      *appv1.Deployment
	routeURL string
}

func NewArtifacts(ctx context.Context /*content []byte,*/, cfg *config.Config) *Artifacts {
	a := &Artifacts{}
	a.ctx = ctx
	a.cfg = cfg

	name := "bac-import-model"
	//a.cm = &corev1.ConfigMap{}
	//a.cm.Namespace = a.cfg.Namespace
	//a.cm.Name = "bac-import-model"
	//if len(content) > 0 {
	//	a.cm.BinaryData = map[string][]byte{}
	//	a.cm.BinaryData["catalog-info-yaml"] = content
	//}

	a.svc = &corev1.Service{}
	a.svc.Namespace = a.cfg.Namespace
	a.svc.Name = name
	a.svc.ObjectMeta.Labels = map[string]string{}
	a.svc.ObjectMeta.Labels["app"] = name
	a.svc.Spec.Selector = map[string]string{}
	a.svc.Spec.Selector["app"] = name
	a.svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:     "location",
			Protocol: corev1.ProtocolTCP,
			Port:     8080,
			//TargetPort: intstr.FromString("location"),
			TargetPort: intstr.FromInt32(8080),
		},
	}

	a.route = &routev1.Route{}
	a.route.Namespace = a.cfg.Namespace
	a.route.Name = name
	a.route.Spec = routev1.RouteSpec{
		To: routev1.RouteTargetReference{Kind: "Service", Name: a.svc.Name},
	}

	a.dpm = &appv1.Deployment{}
	a.dpm.Namespace = a.cfg.Namespace
	a.dpm.Name = name
	a.dpm.ObjectMeta.Labels = map[string]string{}
	a.dpm.ObjectMeta.Labels["app.kubernetes.io/name"] = name
	replicas := int32(1)
	defaultMode := int32(420)
	readOnlyFSnonRoot := true
	a.dpm.Spec = appv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": name},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": name},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            "location",
						Image:           "quay.io/gabemontero/import-location:latest",
						ImagePullPolicy: corev1.PullAlways,
						Ports: []corev1.ContainerPort{
							{
								Name:          "location",
								ContainerPort: 8080,
							},
						},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.Quantity{Format: resource.Format("500m")},
								corev1.ResourceMemory: resource.Quantity{Format: resource.Format("384Mi")},
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.Quantity{Format: resource.Format("5m")},
								corev1.ResourceMemory: resource.Quantity{Format: resource.Format("64Mi")},
							},
						},
						SecurityContext: &corev1.SecurityContext{ReadOnlyRootFilesystem: &readOnlyFSnonRoot},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "location",
								MountPath: "/data",
								ReadOnly:  true,
							},
						},
					},
				},
				SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: &readOnlyFSnonRoot},
				Volumes: []corev1.Volume{
					{
						Name: "location",
						VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: name},
							DefaultMode:          &defaultMode,
						}},
					},
				},
			},
		},
	}

	return a
}
