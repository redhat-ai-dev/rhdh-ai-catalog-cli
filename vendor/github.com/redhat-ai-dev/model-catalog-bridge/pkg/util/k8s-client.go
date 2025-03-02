package util

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	servingset "github.com/kserve/kserve/pkg/client/clientset/versioned"
	servingv1beta1 "github.com/kserve/kserve/pkg/client/clientset/versioned/typed/serving/v1beta1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/config"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	appv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func GetK8sConfig(cfg *config.Config) (*rest.Config, error) {
	if len(cfg.Kubeconfig) > 0 {
		return clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	}
	// If an env variable is specified with the config locaiton, use that
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	// If no explicit location, try the in-cluster config
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}

func GetKServeClient(cfg *rest.Config) servingv1beta1.ServingV1beta1Interface {
	return servingset.NewForConfigOrDie(cfg).ServingV1beta1()
}

func GetCoreClient(cfg *rest.Config) corev1.CoreV1Interface {
	return corev1.NewForConfigOrDie(cfg)
}

func GetAppClient(cfg *rest.Config) appv1.AppsV1Interface {
	return appv1.NewForConfigOrDie(cfg)
}

func GetRouteClient(cfg *rest.Config) routev1.RouteV1Interface {
	return routev1.NewForConfigOrDie(cfg)
}

func GetCurrentProject() string {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	matchVersionKubeConfigFlags := kcmdutil.NewMatchVersionFlags(kubeConfigFlags)
	f := kcmdutil.NewFactory(matchVersionKubeConfigFlags)
	cfg, err := f.ToRawKubeConfigLoader().RawConfig()
	currentProject := ""
	if err == nil {
		currentContext := cfg.Contexts[cfg.CurrentContext]
		if currentContext != nil {
			currentProject = currentContext.Namespace
		}
		return currentProject
	}
	fmt.Fprintf(os.Stderr, "ERROR: could not get default project from kubeconfig: %v", err)
	// see if we are running in pod, or the NAMESPACE env var is set
	b, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	currentProject = string(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching token from k8s pod: %s", err.Error())
		currentProject = os.Getenv("NAMESPACE")
	}
	return currentProject
}

func GetCurrentToken(cfg *rest.Config) string {
	token := cfg.BearerToken
	if len(token) == 0 {
		b, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		token = string(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching token from k8s pod: %s", err.Error())
			token = os.Getenv("K8S_TOKEN")
		}
	}
	return token
}
