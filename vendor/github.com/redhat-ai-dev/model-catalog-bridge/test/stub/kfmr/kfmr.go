package kfmr

import (
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/config"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/rest"
	"github.com/redhat-ai-dev/model-catalog-bridge/test/stub/common"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func SetupKubeflowTestRESTClient(ts *httptest.Server, cfg *config.Config) {
	cfg.StoreURL = ts.URL
	cfg.KubeflowRESTClient = common.DC()
}

func CreateGetServer(t *testing.T) *httptest.Server {
	ts := common.CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case common.MethodGet:
			switch {
			case strings.HasSuffix(r.URL.Path, rest.LIST_REG_MODEL_URI):
				_, _ = w.Write([]byte(common.TestJSONStringRegisteredModelOneLine))
			case strings.HasSuffix(r.URL.Path, "/versions"):
				_, _ = w.Write([]byte(common.TestJSONStringModelVersionOneLine))
			case strings.HasSuffix(r.URL.Path, "/artifacts"):
				_, _ = w.Write([]byte(common.TestJSONStringModelArtifactOneLine))
			case strings.Contains(r.URL.Path, rest.LIST_REG_MODEL_URI):
				_, _ = w.Write([]byte(common.TestJSONStringRegisteredModelOneLineGet))
			}
		}
	})

	return ts
}

func CreateGetServerWithInference(t *testing.T) *httptest.Server {
	ts := common.CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case common.MethodGet:
			switch {
			case strings.HasSuffix(r.URL.Path, rest.LIST_REG_MODEL_URI):
				_, _ = w.Write([]byte(common.MnistRegisteredModels))
			case strings.HasSuffix(r.URL.Path, "versions"):
				_, _ = w.Write([]byte(common.MnistModelVersions))
			case strings.HasSuffix(r.URL.Path, "artifacts"):
				_, _ = w.Write([]byte(common.MnistModelArtifacts))
			case strings.HasSuffix(r.URL.Path, rest.LIST_INFERENCE_SERVICES_URI):
				_, _ = w.Write([]byte(common.MnistInferenceServices))
			case strings.Contains(r.URL.Path, "versions"):
				_, _ = w.Write([]byte(common.MnistModelVersionGet))
			case strings.Contains(r.URL.Path, "artifacts"):
				_, _ = w.Write([]byte(common.MnistModelArtifactsGet))
			case strings.Contains(r.URL.Path, "serving"):
				_, _ = w.Write([]byte(common.MnistServingEnvironmentsGet))
			}
		}
	})

	return ts

}
