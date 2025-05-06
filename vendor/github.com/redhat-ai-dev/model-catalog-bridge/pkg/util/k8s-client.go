package util

import (
	"fmt"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog/v2"
	"net"
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

const (
	PodIPEnvVar              = "POD_IP"
	PodNSEnvVar              = "POD_NAMESPACE"
	BrdgNSEnvVar             = "NAMESPACE"
	KubeconfigEnvVar         = "KUBECONFIG"
	K8sTokenEnvVar           = "K8S_TOKEN"
	SvcHostEnvVar            = "KUBERNETES_SERVICE_HOST"
	SvcPortEnvVar            = "KUBERNETES_SERVICE_PORT"
	SidecarSecretTokenEnvVar = "token"
	SidecarSecretCertEnvVar  = "ca.crt"
)

func GetK8sConfig(cfg *config.Config) (*rest.Config, error) {
	if len(cfg.Kubeconfig) > 0 {
		klog.Infof("loading k8s cfg from cfg setting for kubeconfig %s", cfg.Kubeconfig)
		return clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	}
	// If an env variable is specified with the config locaiton, use that
	if len(os.Getenv(KubeconfigEnvVar)) > 0 {
		klog.Infof("loading k8s cfg from KUBECONFIG env var %s", cfg.Kubeconfig)
		return clientcmd.BuildConfigFromFlags("", os.Getenv(KubeconfigEnvVar))
	}
	klog.Info("using rest in cluster config")
	// If no explicit location, try the in-cluster config
	if c, err := rest.InClusterConfig(); err == nil {
		klog.Info("found in cluster config")
		return c, nil
	} else {
		klog.Errorf("in cluster config error: %s", err.Error())
	}
	klog.Info("trying .config user home location for k8s creds")
	// If no in-cluster config, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, e := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "config")); e == nil {
			klog.Info("using .config user home location for k8s creds")
			return c, nil
		}
	}

	if c, err := InClusterConfigHackForRHDHSidecars(); err == nil {
		klog.Info("using our for RHDH sidecars")
		return c, err
	}

	klog.Errorf("could not location any k8s config")
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
		if len(currentProject) > 0 {
			klog.Infof("raw kube config loader found ns %s", currentProject)
			return currentProject
		}
	}
	fmt.Fprintf(os.Stderr, "ERROR: could not get default project from kubeconfig: %v", err)
	// see if we are running in pod, or the NAMESPACE env var is set
	b, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	currentProject = string(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching token from k8s pod: %s", err.Error())
		currentProject = os.Getenv(BrdgNSEnvVar)
	}
	if len(currentProject) == 0 {
		currentProject = os.Getenv(PodNSEnvVar)
		klog.Infof("sidecar pod namespace is %s", currentProject)
	}
	return currentProject
}

func GetCurrentToken(cfg *rest.Config) string {
	if cfg == nil || len(cfg.BearerToken) == 0 {
		b, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		token := string(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching token from k8s pod: %s", err.Error())
			token = os.Getenv(K8sTokenEnvVar)
		}
		return token
	}
	return cfg.BearerToken
}

// InClusterConfigHackForRHDHSidecars - between the RHDH operator forcing the RHDH deployment to have the setting of
// automountServiceAccountToken: false
// and patching into the volumes array using the coarse grained deployment object proving extremely difficult
// we have had to get *VERY* hacky wrt supplying the REST config to controller runtime.  The method patterns
// k8s.io/client-go/rest/config.go#func InClusterConfig() (*rest.Config, error) but checks env vars we set
// on our various containers for our secret contents; fortunately, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT
// are still set in the container; now, the root CA setup still wants a file, but it was possible to pass the
// ephemeral "dynamic-plugins-root" volume to our sidecar, and then take the ca.crt env setting we get from
// the bridge secret and store in a file we can seed into the TLS client config for our K8S REST Config.
func InClusterConfigHackForRHDHSidecars() (*rest.Config, error) {
	token := os.Getenv(SidecarSecretTokenEnvVar)
	host, port := os.Getenv(SvcHostEnvVar), os.Getenv(SvcPortEnvVar)
	if len(host) == 0 || len(port) == 0 || len(token) == 0 {
		return nil, rest.ErrNotInCluster
	}

	tlsClientConfig := rest.TLSClientConfig{Insecure: true}
	cacrt := os.Getenv(SidecarSecretCertEnvVar)
	if len(cacrt) > 0 {
		fn := "/opt/app-root/src/dynamic-plugins-root/ca.crt"
		file, err := os.Create(fn)
		klog.Infof("create ca.crt file err: %#v", err)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		_, err = file.WriteString(cacrt)
		klog.Infof("write string ca.crt file err: %#v", err)
		if err != nil {
			return nil, err
		}
		err = file.Sync()
		klog.Infof("sync ca.crt file err: %#v", err)
		if err != nil {
			return nil, err
		}
		if _, err = certutil.NewPool(fn); err == nil {
			tlsClientConfig.CAFile = fn
			tlsClientConfig.Insecure = false
		} else {
			klog.Infof("new pool ca.crt file err: %#v", err.Error())
			return nil, err
		}

	}

	return &rest.Config{
		Host:            "https://" + net.JoinHostPort(host, port),
		TLSClientConfig: tlsClientConfig,
		BearerToken:     token,
	}, nil
}
