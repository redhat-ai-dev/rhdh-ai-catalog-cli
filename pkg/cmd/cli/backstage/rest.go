package backstage

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
	nurl "net/url"
	"os"
)

const (
	BASE_URI      = "/api/catalog"
	LOCATION_URI  = "/locations"
	ENTITIES_URI  = "/entities"
	COMPONENT_URI = "/entities/by-name/component/%s/%s"
	RESOURCE_URI  = "/entities/by-name/resource/%s/%s"
	API_URI       = "/entities/by-name/api/%s/%s"
	QUERY_URI     = "/entities/by-query"
	DEFAULT_NS    = "default"
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
	backstageRESTClient.RootURL = cfg.BackstageURL + BASE_URI

	backstageRESTClient.Tags = cfg.ParamsAsTags
	backstageRESTClient.Subset = cfg.AnySubsetWorks

	return backstageRESTClient
}

func (k *BackstageRESTClientWrapper) processUpdate(resp *resty.Response, action, url, body string) (string, error) {
	postResp := resp.String()
	rc := resp.StatusCode()
	if rc != 200 && rc != 201 {
		return "", fmt.Errorf("%s %s with body %s status code %d resp: %s\n", url, action, body, rc, postResp)
	} else {
		klog.V(4).Infof("%s %s with body %s status code %d resp: %s\n", url, action, body, rc, postResp)
	}
	return k.processBody(resp)
}

func (k *BackstageRESTClientWrapper) processBody(resp *resty.Response) (string, error) {
	retJSON := make(map[string]any)
	err := json.Unmarshal(resp.Body(), &retJSON)
	if err != nil {
		return "", fmt.Errorf("json unmarshall error for %s: %s\n", resp.Body(), err.Error())
	}
	var location interface{}
	var id interface{}
	var target interface{}
	var ok bool
	location, ok = retJSON["location"]
	if ok {
		locationMap, o1 := location.(map[string]interface{})
		if o1 {
			id = locationMap["id"]
			target = locationMap["target"]
		}
		return fmt.Sprintf("Backstage location %s from %s created", id, target), nil
	}
	id, ok = retJSON["id"]
	if ok {
		target, ok = retJSON["target"]
		if ok {
			return fmt.Sprintf("Backstage location %s from %s created", id, target), nil
		}
		return fmt.Sprintf("Backstage location %s created", id), nil
	}
	return fmt.Sprintf("%#v", retJSON), nil
}

func (k *BackstageRESTClientWrapper) postToBackstage(url string, body interface{}) (string, error) {
	resp, err := backstageRESTClient.RESTClient.R().SetAuthToken(k.Token).SetBody(body).SetHeader("Accept", "application/json").Post(url)
	if err != nil {
		return "", err
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
