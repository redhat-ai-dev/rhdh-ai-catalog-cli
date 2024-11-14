package kubeflowmodelregistry

import (
	"crypto/tls"
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
	"os"
)

const (
	BASE_URI                         = "/api/model_registry/v1alpha3"
	GET_REG_MODEL_URI                = "/registered_models/%s"
	LIST_VERSIONS_OFF_REG_MODELS_URI = "/registered_models/%s/versions"
	LIST_ARTFIACTS_OFF_VERSIONS_URI  = "/model_versions/%s/artifacts"
	LIST_REG_MODEL_URI               = "/registered_models"
)

type KubeFlowRESTClientWrapper struct {
	RESTClient *resty.Client
	RootURL    string
	Token      string
}

func SetupKubeflowRESTClient(cfg *config.Config) *KubeFlowRESTClientWrapper {
	if cfg == nil {
		klog.Error("Command config is nil")
		klog.Flush()
		os.Exit(1)
	}
	kubeFlowRESTClient := &KubeFlowRESTClientWrapper{
		Token:      cfg.StoreToken,
		RootURL:    cfg.StoreURL + BASE_URI,
		RESTClient: cfg.KubeflowRESTClient,
	}
	if cfg.KubeflowRESTClient != nil {
		return kubeFlowRESTClient
	}
	cfg.KubeflowRESTClient = resty.New()
	kubeFlowRESTClient.RESTClient = cfg.KubeflowRESTClient
	if cfg.KubeflowRESTClient == nil {
		klog.Errorf("Unable to get Kubeflow REST client wrapper")
		klog.Flush()
		os.Exit(1)
	}
	tlsCfg := &tls.Config{}
	if cfg.StoreSkipTLS {
		tlsCfg.InsecureSkipVerify = true
	}
	kubeFlowRESTClient.RESTClient.SetTLSClientConfig(tlsCfg)

	return kubeFlowRESTClient
}

func (k *KubeFlowRESTClientWrapper) getFromModelRegistry(url string) ([]byte, error) {
	resp, err := k.RESTClient.R().SetAuthToken(k.Token).Get(url)
	if err != nil {
		return nil, err
	}
	rc := resp.StatusCode()
	getResp := resp.String()
	if rc != 200 {
		return nil, fmt.Errorf("get for %s rc %d body %s\n", url, rc, getResp)
	} else {
		klog.V(4).Infof("get for %s returned ok\n", url)
	}
	return resp.Body(), err

}
