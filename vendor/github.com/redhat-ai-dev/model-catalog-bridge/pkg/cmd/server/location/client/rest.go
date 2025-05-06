package client

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/rest"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/util"
	"net/http"
)

type BridgeLocationRESTClient struct {
	RESTClient *resty.Client
	HostURL    string
	UpsertURL  string
	RemoveURL  string
	Token      string
}

func SetupBridgeLocationRESTClient(hostURL, token string) *BridgeLocationRESTClient {
	b := &BridgeLocationRESTClient{
		RESTClient: resty.New(),
		HostURL:    hostURL,
		UpsertURL:  hostURL + util.UpsertURI,
		RemoveURL:  hostURL + util.RemoveURI,
		Token:      token,
	}
	return b
}

func (b *BridgeLocationRESTClient) UpsertModel(importKey string, buf []byte) (int, string, *rest.PostBody, error) {
	var err error
	var locationResp *resty.Response
	body := rest.PostBody{
		Body: buf,
	}
	locationResp, err = b.RESTClient.R().SetBody(body).SetAuthToken(b.Token).SetQueryParam(util.KeyQueryParam, importKey).SetHeader("Accept", "application/json").Post(b.UpsertURL)
	msg := fmt.Sprintf("%#v", locationResp)
	if err != nil {
		return http.StatusInternalServerError, msg, &body, err
	}
	return locationResp.StatusCode(), msg, &body, nil
}

func (b *BridgeLocationRESTClient) RemoveModel(key string) (int, string, error) {
	resp, err := b.RESTClient.R().SetAuthToken(b.Token).SetQueryParam(util.KeyQueryParam, key).SetHeader("Accept", "application/json").Delete(b.RemoveURL)
	msg := fmt.Sprintf("%#v", resp)
	if err != nil {
		return http.StatusInternalServerError, msg, err
	}
	return resp.StatusCode(), msg, nil
}
