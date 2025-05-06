package kfmr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/config"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/rest"
	"github.com/redhat-ai-dev/model-catalog-bridge/test/stub/common"
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
			case strings.HasSuffix(r.URL.Path, fmt.Sprintf("%s/%s", rest.LIST_REG_MODEL_URI, "1")):
				_, _ = w.Write([]byte(common.MnistRegisteredModelsGet))
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

func CreateGetServerWithMixInferenceMultiModel(t *testing.T) *httptest.Server {
	ts := common.CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case common.MethodGet:
			switch {
			case strings.HasSuffix(r.URL.Path, rest.LIST_REG_MODEL_URI):
				_, _ = w.Write([]byte(common.MultiModelRegisteredModels))

			case strings.HasSuffix(r.URL.Path, fmt.Sprintf(rest.LIST_VERSIONS_OFF_REG_MODELS_URI, "1")):
				_, _ = w.Write([]byte(common.Granite8bLabModelVersions))
			case strings.HasSuffix(r.URL.Path, fmt.Sprintf(rest.LIST_VERSIONS_OFF_REG_MODELS_URI, "3")):
				_, _ = w.Write([]byte(common.Granite8bCodeBaseModelVersions))
			case strings.HasSuffix(r.URL.Path, fmt.Sprintf(rest.LIST_VERSIONS_OFF_REG_MODELS_URI, "5")):
				_, _ = w.Write([]byte(common.MultiModelMnistModelVersions))

			case strings.HasSuffix(r.URL.Path, fmt.Sprintf(rest.LIST_ARTFIACTS_OFF_VERSIONS_URI, "2")):
				_, _ = w.Write([]byte(common.Granite8bLabModelArtifacts))
			case strings.HasSuffix(r.URL.Path, fmt.Sprintf(rest.LIST_ARTFIACTS_OFF_VERSIONS_URI, "4")):
				_, _ = w.Write([]byte(common.Granite8bCodeBaseModelArtifacts))
			case strings.HasSuffix(r.URL.Path, fmt.Sprintf(rest.LIST_ARTFIACTS_OFF_VERSIONS_URI, "6")):
				_, _ = w.Write([]byte(common.MultiModelMnistModelArtifacts))

			case strings.HasSuffix(r.URL.Path, rest.LIST_INFERENCE_SERVICES_URI):
				_, _ = w.Write([]byte(common.MultiModelInferenceServices))

			case strings.Contains(r.URL.Path, "serving"):
				_, _ = w.Write([]byte(common.MultiModelServingEnvironmentGet))
			}
		}
	})

	return ts

}

func CreateGetServerArchived(t *testing.T) *httptest.Server {
	ts := common.CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case common.MethodGet:
			switch {
			case strings.HasSuffix(r.URL.Path, rest.LIST_REG_MODEL_URI):
				_, _ = w.Write([]byte(common.MnistRegisteredModelsArchived))
			case strings.HasSuffix(r.URL.Path, "versions"):
				_, _ = w.Write([]byte(common.MnistModelVersionsArchived))
			case strings.HasSuffix(r.URL.Path, "artifacts"):
				_, _ = w.Write([]byte(common.MnistModelArtifactsArchived))
			case strings.Contains(r.URL.Path, "versions"):
				_, _ = w.Write([]byte(common.MnistModelVersionGetArchived))
			case strings.Contains(r.URL.Path, "artifacts"):
				_, _ = w.Write([]byte(common.MnistModelArtifactsGetArchived))
			}
		}
	})

	return ts

}
