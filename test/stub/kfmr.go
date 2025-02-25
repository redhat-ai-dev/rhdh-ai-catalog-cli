package stub

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	TestJSONStringRegisteredModelOneLine    = `{"items":[{"createTimeSinceEpoch":"1731103949567","customProperties":{"foo":{"metadataType":"MetadataStringValue","string_value":"bar"}},"description":"dummy model 1","id":"1","lastUpdateTimeSinceEpoch":"1731103975700","name":"model-1","Owner":"kube:admin","state":"LIVE"}],"nextPageToken":"","pageSize":0,"size":1}`
	TestJSONStringRegisteredModelOneLineGet = `{"createTimeSinceEpoch":"1731103949567","customProperties":{"foo":{"metadataType":"MetadataStringValue","string_value":"bar"}},"description":"dummy model 1","id":"1","lastUpdateTimeSinceEpoch":"1731103975700","name":"model-1","Owner":"kube:admin","state":"LIVE"}`
	TestJSONStringModelVersionOneLine       = `{"items":[{"author":"kube:admin","createTimeSinceEpoch":"1731103949724","customProperties":{},"description":"version 1","id":"2","lastUpdateTimeSinceEpoch":"1731103949724","name":"v1","registeredModelId":"1","state":"LIVE"}],"nextPageToken":"","pageSize":0,"size":1}`
	TestJSONStringModelArtifactOneLine      = `{"items":[{"artifactType":"model-artifact","createTimeSinceEpoch":"1731103949909","customProperties":{},"description":"version 1","id":"1","lastUpdateTimeSinceEpoch":"1731103949909","modelFormatName":"tensorflow","modelFormatVersion":"v1","name":"model-1-v1-artifact","state":"LIVE","uri":"https://foo.com"}],"nextPageToken":"","pageSize":0,"size":1}`

	MnistRegisteredModels    = `{"items":[{"createTimeSinceEpoch":"1740498236442","customProperties":{"_lastModified":{"metadataType":"MetadataStringValue","string_value":"2025-02-25T15:43:56.876Z"}},"id":"1","lastUpdateTimeSinceEpoch":"1740498237116","name":"mnist","owner":"kube:admin","state":"LIVE"}],"nextPageToken":"","pageSize":0,"size":1}`
	MnistModelVersions       = `{"items":[{"author":"kube:admin","createTimeSinceEpoch":"1740498236719","customProperties":{"_lastModified":{"metadataType":"MetadataStringValue","string_value":"2025-02-25T19:45:29.959Z"}},"id":"2","lastUpdateTimeSinceEpoch":"1740512730384","name":"v1","registeredModelId":"1","state":"LIVE"}],"nextPageToken":"","pageSize":0,"size":1}`
	MnistModelArtifacts      = `{"items":[{"artifactType":"model-artifact","createTimeSinceEpoch":"1740498237469","customProperties":{},"id":"1","lastUpdateTimeSinceEpoch":"1740498237469","modelFormatName":"onnx","modelFormatVersion":"1","name":"v1","state":"LIVE","uri":"https://huggingface.co/tarilabs/mnist/resolve/v20231206163028/mnist.onnx"}],"nextPageToken":"","pageSize":0,"size":1}`
	MnistInferenceServices   = `{"items":[{"createTimeSinceEpoch":"1740512730723","customProperties":{},"desiredState":"DEPLOYED","id":"4","lastUpdateTimeSinceEpoch":"1740512730723","modelVersionId":"2","name":"mnist-v1/8c2c357f-bf82-4d2d-a254-43eca96fd31d","registeredModelId":"1","runtime":"mnist-v1","servingEnvironmentId":"3"}],"nextPageToken":"","pageSize":0,"size":1}`
	MnistServingEnvironments = `{"items":[{"createTimeSinceEpoch":"1740512730477","customProperties":{},"id":"3","lastUpdateTimeSinceEpoch":"1740512730477","name":"ggmtest"}],"nextPageToken":"","pageSize":0,"size":1}`

	MnistServingEnvironmentsGet = `{"createTimeSinceEpoch":"1740512730477","customProperties":{},"id":"3","lastUpdateTimeSinceEpoch":"1740512730477","name":"ggmtest"}`
	MnistModelArtifactsGet      = `{"artifactType":"model-artifact","createTimeSinceEpoch":"1740498237469","customProperties":{},"id":"1","lastUpdateTimeSinceEpoch":"1740498237469","modelFormatName":"onnx","modelFormatVersion":"1","name":"v1","state":"LIVE","uri":"https://huggingface.co/tarilabs/mnist/resolve/v20231206163028/mnist.onnx"}`
	MnistModelVersionGet        = `{"author":"kube:admin","createTimeSinceEpoch":"1740498236719","customProperties":{"_lastModified":{"metadataType":"MetadataStringValue","string_value":"2025-02-25T19:45:29.959Z"}},"id":"2","lastUpdateTimeSinceEpoch":"1740512730384","name":"v1","registeredModelId":"1","state":"LIVE"}`
)

func SetupKubeflowTestRESTClient(ts *httptest.Server, cfg *config.Config) {
	cfg.StoreURL = ts.URL
	cfg.KubeflowRESTClient = DC()
}

func CreateGetServer(t *testing.T) *httptest.Server {
	ts := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case MethodGet:
			switch {
			case strings.HasSuffix(r.URL.Path, rest.LIST_REG_MODEL_URI):
				_, _ = w.Write([]byte(TestJSONStringRegisteredModelOneLine))
			case strings.HasSuffix(r.URL.Path, "/versions"):
				_, _ = w.Write([]byte(TestJSONStringModelVersionOneLine))
			case strings.HasSuffix(r.URL.Path, "/artifacts"):
				_, _ = w.Write([]byte(TestJSONStringModelArtifactOneLine))
			case strings.Contains(r.URL.Path, rest.LIST_REG_MODEL_URI):
				_, _ = w.Write([]byte(TestJSONStringRegisteredModelOneLineGet))
			}
		}
	})

	return ts
}

func CreateGetServerWithInference(t *testing.T) *httptest.Server {
	ts := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case MethodGet:
			switch {
			case strings.HasSuffix(r.URL.Path, rest.LIST_REG_MODEL_URI):
				_, _ = w.Write([]byte(MnistRegisteredModels))
			case strings.HasSuffix(r.URL.Path, "versions"):
				_, _ = w.Write([]byte(MnistModelVersions))
			case strings.HasSuffix(r.URL.Path, "artifacts"):
				_, _ = w.Write([]byte(MnistModelArtifacts))
			case strings.HasSuffix(r.URL.Path, rest.LIST_INFERENCE_SERVICES_URI):
				_, _ = w.Write([]byte(MnistInferenceServices))
			case strings.Contains(r.URL.Path, "versions"):
				_, _ = w.Write([]byte(MnistModelVersionGet))
			case strings.Contains(r.URL.Path, "artifacts"):
				_, _ = w.Write([]byte(MnistModelArtifactsGet))
			case strings.Contains(r.URL.Path, "serving"):
				_, _ = w.Write([]byte(MnistServingEnvironmentsGet))
			}
		}
	})

	return ts

}
