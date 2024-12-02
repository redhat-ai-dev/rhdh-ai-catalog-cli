package kubeflowmodelregistry

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	TestJSONStringRegisteredModelOneLine    = `{"items":[{"createTimeSinceEpoch":"1731103949567","customProperties":{"foo":{"metadataType":"MetadataStringValue","string_value":"bar"}},"description":"dummy model 1","id":"1","lastUpdateTimeSinceEpoch":"1731103975700","name":"model-1","owner":"kube:admin","state":"LIVE"}],"nextPageToken":"","pageSize":0,"size":1}`
	TestJSONStringRegisteredModelOneLineGet = `{"createTimeSinceEpoch":"1731103949567","customProperties":{"foo":{"metadataType":"MetadataStringValue","string_value":"bar"}},"description":"dummy model 1","id":"1","lastUpdateTimeSinceEpoch":"1731103975700","name":"model-1","owner":"kube:admin","state":"LIVE"}`
	TestJSONStringModelVersionOneLine       = `{"items":[{"author":"kube:admin","createTimeSinceEpoch":"1731103949724","customProperties":{},"description":"version 1","id":"2","lastUpdateTimeSinceEpoch":"1731103949724","name":"v1","registeredModelId":"1","state":"LIVE"}],"nextPageToken":"","pageSize":0,"size":1}`
	TestJSONStringModelArtifactOneLine      = `{"items":[{"artifactType":"model-artifact","createTimeSinceEpoch":"1731103949909","customProperties":{},"description":"version 1","id":"1","lastUpdateTimeSinceEpoch":"1731103949909","modelFormatName":"tensorflow","modelFormatVersion":"v1","name":"model-1-v1-artifact","state":"LIVE","uri":"https://foo.com"}],"nextPageToken":"","pageSize":0,"size":1}`
)

func SetupKubeflowTestRESTClient(ts *httptest.Server, cfg *config.Config) {
	cfg.StoreURL = ts.URL
	cfg.KubeflowRESTClient = stub.DC()
}

func CreateGetServer(t *testing.T) *httptest.Server {
	ts := stub.CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case stub.MethodGet:
			switch {
			case strings.HasSuffix(r.URL.Path, LIST_REG_MODEL_URI):
				_, _ = w.Write([]byte(TestJSONStringRegisteredModelOneLine))
			case strings.HasSuffix(r.URL.Path, "/versions"):
				_, _ = w.Write([]byte(TestJSONStringModelVersionOneLine))
			case strings.HasSuffix(r.URL.Path, "/artifacts"):
				_, _ = w.Write([]byte(TestJSONStringModelArtifactOneLine))
			case strings.Contains(r.URL.Path, LIST_REG_MODEL_URI):
				_, _ = w.Write([]byte(TestJSONStringRegisteredModelOneLineGet))
			}
		}
	})

	return ts
}
