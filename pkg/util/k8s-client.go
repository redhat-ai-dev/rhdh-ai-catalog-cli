package util

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	servingset "github.com/kserve/kserve/pkg/client/clientset/versioned"
	servingv1beta1 "github.com/kserve/kserve/pkg/client/clientset/versioned/typed/serving/v1beta1"

	"k8s.io/cli-runtime/pkg/genericclioptions"
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

func GetCurrentProject() string {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	matchVersionKubeConfigFlags := kcmdutil.NewMatchVersionFlags(kubeConfigFlags)
	f := kcmdutil.NewFactory(matchVersionKubeConfigFlags)
	cfg, err := f.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not get default project: %v", err)
		return ""
	}
	currentProject := ""
	currentContext := cfg.Contexts[cfg.CurrentContext]
	if currentContext != nil {
		currentProject = currentContext.Namespace
	}
	return currentProject
}
