package kubeflowmodelregistry

import (
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/kserve"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"k8s.io/klog/v2"
	"os"
)

const (
	BASE_URI                         = "/api/model_registry/v1alpha3"
	GET_REG_MODEL_URI                = "/registered_models/%s"
	LIST_VERSIONS_OFF_REG_MODELS_URI = "/registered_models/%s/versions"
	LIST_ARTFIACTS_OFF_VERSIONS_URI  = "/model_versions/%s/artifacts"
	LIST_INFERENCE_SERVICES_URI      = "/inference_services"
	LIST_REG_MODEL_URI               = "/registered_models"
	GET_SERVING_ENV_URI              = "/serving_environments/%s"
)

type KubeFlowRESTClientWrapper struct {
	RESTClient *resty.Client
	Config     *config.Config
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

	//TODO unless https://issues.redhat.com/browse/RHOAIENG-16898 gets processed such that KFMR
	// starts adding the KServer inferenceservice.status.url in the inference_service / serving_environment
	// object, we need to take the names from those two items as the name and namespace for a KServe client lookup.
	// Hence we have to set up the KServe serving client
	kserve.SetupKServeClient(cfg)
	kubeFlowRESTClient.Config = cfg

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
