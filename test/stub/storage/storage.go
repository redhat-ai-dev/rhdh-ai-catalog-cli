package storage

import (
	"encoding/json"
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/server/storage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/common"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func SetupBridgeStorageRESTClient(ts *httptest.Server) *storage.BridgeStorageRESTClient {
	storageTC := &storage.BridgeStorageRESTClient{}
	storageTC.RESTClient = common.DC()
	storageTC.UpsertURL = ts.URL
	return storageTC
}

func CreateBridgeStorageRESTClient(t *testing.T, called *sync.Map) *httptest.Server {
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
				data := rest.PostBody{}
				err = json.Unmarshal(bodyBuf, &data)
				if err != nil {
					t.Logf("error unmarshall into storage PostBody: %s", err.Error())
					_, _ = w.Write([]byte(fmt.Sprintf(common.TestPostJSONStringOneLinePlusBody, err.Error())))
					w.WriteHeader(500)
					return
				}
				t.Logf("GGM got buf of len %d", len(data.Body))
				called.Store(r.URL.Path, string(data.Body))
				_, _ = w.Write([]byte(fmt.Sprintf(common.TestPostJSONStringOneLinePlusBody, string(data.Body))))
				w.WriteHeader(201)

			}
		}
	})
	return ts
}
