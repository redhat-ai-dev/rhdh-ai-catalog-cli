package client

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/server/gin-gonic-http/server"
	"net/http"
)

type BridgeLocationRESTClient struct {
	RESTClient *resty.Client
	UpsertURL  string
	Token      string
}

func SetupBridgeLocationRESTClient(hostURL, token string) *BridgeLocationRESTClient {
	b := &BridgeLocationRESTClient{
		RESTClient: resty.New(),
		UpsertURL:  hostURL + "/upsert",
		Token:      token,
	}
	return b
}

func (b *BridgeLocationRESTClient) UpsertModel(importKey string, buf []byte) (int, string, *server.PostBody, error) {
	var err error
	var locationResp *resty.Response
	body := server.PostBody{
		Body: buf,
	}
	locationResp, err = b.RESTClient.R().SetBody(body).SetAuthToken(b.Token).SetQueryParam("key", importKey).SetHeader("Accept", "application/json").Post(b.UpsertURL)
	msg := fmt.Sprintf("%#v", locationResp)
	if err != nil {
		return http.StatusInternalServerError, msg, &body, err
	}
	return locationResp.StatusCode(), msg, &body, nil
}
