package storage

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
	"net/http"
)

type BridgeStorageRESTClient struct {
	RESTClient *resty.Client
	UpsertURL  string
	Token      string
}

func SetupBridgeStorageRESTClient(hostURL, token string) *BridgeStorageRESTClient {
	b := &BridgeStorageRESTClient{
		RESTClient: resty.New(),
		UpsertURL:  hostURL + "/upsert",
		Token:      token,
	}
	return b
}

func (b *BridgeStorageRESTClient) UpsertModel(importKey string, buf []byte) (int, string, *rest.PostBody, error) {
	var err error
	var storageResp *resty.Response
	body := rest.PostBody{
		Body: buf,
	}
	storageResp, err = b.RESTClient.R().SetBody(body).SetAuthToken(b.Token).SetQueryParam("key", importKey).SetHeader("Accept", "application/json").Post(b.UpsertURL)
	msg := fmt.Sprintf("%#v", storageResp)
	if err != nil {
		return http.StatusInternalServerError, msg, &body, err
	}
	return storageResp.StatusCode(), msg, &body, nil
}
