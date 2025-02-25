package backstage

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub"
	"k8s.io/klog/v2"
	"net/http/httptest"
	nurl "net/url"
	"os"
)

type BackstageRESTClientWrapper struct {
	RESTClient *resty.Client
	RootURL    string
	Token      string
	Tags       bool
	Subset     bool
}

var backstageRESTClient = &BackstageRESTClientWrapper{}

func init() {
	backstageRESTClient.RESTClient = resty.New()
	if backstageRESTClient == nil {
		klog.Errorf("Unable to get Backstage REST client wrapper")
		os.Exit(1)
	}
}

func SetupBackstageTestRESTClient(ts *httptest.Server) *BackstageRESTClientWrapper {
	backstageTestRESTClient := &BackstageRESTClientWrapper{}
	backstageTestRESTClient.RESTClient = stub.DC()
	backstageTestRESTClient.RootURL = ts.URL
	return backstageTestRESTClient
}

func SetupBackstageRESTClient(cfg *config.Config) *BackstageRESTClientWrapper {
	if cfg == nil {
		klog.Error("Command config is nil")
		os.Exit(1)
	}
	tlsCfg := &tls.Config{}
	if cfg.BackstageSkipTLS {
		tlsCfg.InsecureSkipVerify = true
	}
	backstageRESTClient.RESTClient.SetTLSClientConfig(tlsCfg)
	backstageRESTClient.Token = cfg.BackstageToken
	backstageRESTClient.RootURL = cfg.BackstageURL + rest.BASE_URI

	backstageRESTClient.Tags = cfg.ParamsAsTags
	backstageRESTClient.Subset = cfg.AnySubsetWorks

	return backstageRESTClient
}

func (k *BackstageRESTClientWrapper) processUpdate(resp *resty.Response, action, url, body string) (map[string]any, error) {
	postResp := resp.String()
	rc := resp.StatusCode()
	if rc != 200 && rc != 201 {
		return nil, fmt.Errorf("%s %s with body %s status code %d resp: %s\n", url, action, body, rc, postResp)
	} else {
		klog.V(4).Infof("%s %s with body %s status code %d resp: %s\n", url, action, body, rc, postResp)
	}
	return k.processBody(resp)
}

func (k *BackstageRESTClientWrapper) processBody(resp *resty.Response) (map[string]any, error) {
	retJSON := make(map[string]any)
	err := json.Unmarshal(resp.Body(), &retJSON)
	if err != nil {
		return nil, fmt.Errorf("json unmarshall error for %s: %s\n", resp.Body(), err.Error())
	}
	return retJSON, err
}

func (k *BackstageRESTClientWrapper) postToBackstage(url string, body interface{}) (map[string]any, error) {
	resp, err := backstageRESTClient.RESTClient.R().SetAuthToken(k.Token).SetBody(body).SetHeader("Accept", "application/json").Post(url)
	if err != nil {
		return nil, err
	}
	rc := resp.StatusCode()
	if rc != 200 && rc != 201 {
		return nil, fmt.Errorf("post for %s rc %d body %s\n", url, rc, resp.String())
	} else {
		klog.V(4).Infof("post for %s returned ok\n", url)

	}

	return k.processUpdate(resp, "post", url, fmt.Sprintf("%#v", body))
}

func (k *BackstageRESTClientWrapper) processFetch(resp *resty.Response, url, action string) (string, error) {
	rc := resp.StatusCode()
	getResp := resp.String()
	if rc != 200 {
		return "", fmt.Errorf("%s for %s rc %d body %s\n", action, url, rc, getResp)
	} else {
		klog.V(4).Infof("%s for %s returned ok\n", action, url)
	}
	return getResp, nil
}

func (k *BackstageRESTClientWrapper) processDelete(resp *resty.Response, url, action string) (string, error) {
	rc := resp.StatusCode()
	getResp := resp.String()
	if rc != 204 && rc != 200 {
		return "", fmt.Errorf("%s for %s rc %d body %s\n", action, url, rc, getResp)
	} else {
		klog.V(4).Infof("%s for %s returned ok\n", action, url)
	}
	return getResp, nil
}

func (k *BackstageRESTClientWrapper) getFromBackstage(url string) (string, error) {
	resp, err := backstageRESTClient.RESTClient.R().SetAuthToken(k.Token).SetHeader("Accept", "application/json").Get(url)
	if err != nil {
		return "", err
	}
	return k.processFetch(resp, url, "get")

}

func (k *BackstageRESTClientWrapper) getWithKindParamFromBackstage(url string, qparams *nurl.Values) (string, error) {
	req := backstageRESTClient.RESTClient.R().SetAuthToken(k.Token).SetHeader("Accept", "application/json")
	if qparams.Has("filter") {
		req.SetQueryParamsFromValues(*qparams)
	}
	resp, err := req.Get(url)
	if err != nil {
		return "", err
	}
	return k.processFetch(resp, url, "get")

}

func (k *BackstageRESTClientWrapper) deleteFromBackstage(url string) (string, error) {
	resp, err := backstageRESTClient.RESTClient.R().SetAuthToken(k.Token).Delete(url)
	if err != nil {
		return "", err
	}
	return k.processDelete(resp, url, "delete")
}
