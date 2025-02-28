package location

import (
	"encoding/json"
	"fmt"
	bridgeclient "github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/server/location/client"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/common"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func SetupBridgeLocationRESTClient(ts *httptest.Server) *bridgeclient.BridgeLocationRESTClient {
	bkstgTC := &bridgeclient.BridgeLocationRESTClient{}
	bkstgTC.RESTClient = common.DC()
	bkstgTC.UpsertURL = ts.URL
	return bkstgTC
}

func CreateBridgeLocationServer(t *testing.T) *httptest.Server {
	ts := common.CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		switch r.Method {
		case common.MethodPost:
			switch r.URL.Path {
			default:
				w.Header().Set("Content-Type", "application/json")
				bodyBuf, err := io.ReadAll(r.Body)
				if err != nil {
					_, _ = w.Write([]byte(fmt.Sprintf(common.TestPostJSONStringOneLinePlusBody, err.Error())))
					w.WriteHeader(500)
					return
				}
				if len(bodyBuf) == 0 {
					w.WriteHeader(500)
					return
				}
				data := backstage.Post{}
				err = json.Unmarshal(bodyBuf, &data)
				if err != nil {
					_, _ = w.Write([]byte(fmt.Sprintf(common.TestPostJSONStringOneLinePlusBody, err.Error())))
					w.WriteHeader(500)
					return
				}
				_, err = url.Parse(data.Target)
				if err != nil {
					w.WriteHeader(500)
					return
				}
				_, _ = w.Write([]byte(fmt.Sprintf(common.TestPostJSONStringOneLinePlusBody, data.Target)))
				w.WriteHeader(201)

			}
		}
	})
	return ts
}
