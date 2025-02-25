package stub

import (
	"encoding/json"
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const (
	TestJSONStringOneLine  = `{"TestGet": "JSON response"}`
	TestJSONStringIndented = `{
    "TestGet": "JSON response"
}`
	TestJSONStringOneLinePlusPathParam  = `{"TestGet": "JSON response path %s"}`
	TestJSONStringOneLinePlusQueryParam = `{"TestGet": "JSON response query %s"}`

	TestPostJSONStringOneLinePlusBody = `{"TestPost": "JSON response body %s"}`

	TestDeleteJSONStringOneLinePlusPathParam = `{"TestDelete": "JSON response path %s"}`
)

func CreateServer(t *testing.T) *httptest.Server {
	ts := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		switch r.Method {
		case MethodGet:
			switch r.URL.Path {
			case "/":
				_, _ = w.Write([]byte("TestGet: text response"))
				return
			case rest.ENTITIES_URI:
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(TestJSONStringOneLine))
				return
			case rest.LOCATION_URI:
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(TestJSONStringOneLine))
				return
			}

			switch {
			case strings.HasPrefix(r.URL.Path, rest.QUERY_URI):
				w.Header().Set("Content-Type", "application/json")
				values := r.URL.Query()
				filter := values.Get("filter")
				switch {
				case strings.Contains(filter, "api"):
					_, _ = w.Write([]byte(ApisJson))
					//if strings.Contains(filter, "metadata") {
					//	_, _ = w.Write([]byte(apisJsonFromTags))
					//} else {
					//	_, _ = w.Write([]byte(apisJson))
					//}
				case strings.Contains(filter, "component"):
					_, _ = w.Write([]byte(ComponentsJson))
				case strings.Contains(filter, "resource"):
					_, _ = w.Write([]byte(ResourcesJson))
					//if strings.Contains(filter, "metadata") {
					//	_, _ = w.Write([]byte(resourcesFromTagsJson))
					//} else {
					//	_, _ = w.Write([]byte(resourcesJson))
					//}
				default:
					_, _ = w.Write([]byte(fmt.Sprintf(TestJSONStringOneLinePlusQueryParam, r.URL.Query().Encode())))
				}

			case strings.HasPrefix(r.URL.Path, rest.LOCATION_URI):
				path := strings.TrimPrefix(r.URL.Path, rest.LOCATION_URI)
				w.Header().Set("Content-Type", "application/json")
				if strings.Contains(path, "404") {
					w.WriteHeader(404)
					return
				}
				_, _ = w.Write([]byte(fmt.Sprintf(TestJSONStringOneLinePlusPathParam, path)))
			case strings.HasPrefix(r.URL.Path, rest.API_URI):
				path := strings.TrimPrefix(r.URL.Path, rest.LOCATION_URI)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(fmt.Sprintf(TestJSONStringOneLinePlusPathParam, path)))
			case strings.HasPrefix(r.URL.Path, rest.ENTITIES_URI):
				w.Header().Set("Content-Type", "application/json")
				segs := strings.Split(r.URL.Path, "/")
				ns := segs[len(segs)-2]
				if ns == "404" {
					w.WriteHeader(404)
					return
				}
				_, _ = w.Write([]byte(fmt.Sprintf(TestJSONStringOneLinePlusPathParam, fmt.Sprintf("%s:%s", ns, segs[len(segs)-1]))))
			}
		case MethodPost:
			switch r.URL.Path {
			case rest.LOCATION_URI:
				w.Header().Set("Content-Type", "application/json")
				bodyBuf, err := io.ReadAll(r.Body)
				if err != nil {
					_, _ = w.Write([]byte(fmt.Sprintf(TestPostJSONStringOneLinePlusBody, err.Error())))
					w.WriteHeader(500)
					return
				}
				if len(bodyBuf) == 0 {
					w.WriteHeader(500)
					return
				}
				data := Post{}
				err = json.Unmarshal(bodyBuf, &data)
				if err != nil {
					_, _ = w.Write([]byte(fmt.Sprintf(TestPostJSONStringOneLinePlusBody, err.Error())))
					w.WriteHeader(500)
					return
				}
				_, err = url.Parse(data.Target)
				if err != nil {
					w.WriteHeader(500)
					return
				}
				_, _ = w.Write([]byte(fmt.Sprintf(TestPostJSONStringOneLinePlusBody, data.Target)))
			}
		case MethodDelete:
			switch {
			case strings.HasPrefix(r.URL.Path, rest.LOCATION_URI):
				path := strings.TrimPrefix(r.URL.Path, rest.LOCATION_URI)
				if strings.Contains(path, "404") {
					w.WriteHeader(404)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(fmt.Sprintf(TestDeleteJSONStringOneLinePlusPathParam, path)))
			}
		}
	})

	return ts
}

type Post struct {
	Target string `json:"target"`
	Type   string `json:"type"`
}
