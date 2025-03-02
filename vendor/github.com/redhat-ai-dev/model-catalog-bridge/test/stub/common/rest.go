package common

import (
	"github.com/go-resty/resty/v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodDelete = "DELETE"
)

var (
	hdrContentTypeKey = http.CanonicalHeaderKey("Content-Type")
)

func CreateTestServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func DC() *resty.Client {
	c := resty.New()
	c.SetLogger(&logger{})
	return c
}

type logger struct{}

func (l *logger) Errorf(format string, v ...interface{}) {
}

func (l *logger) Warnf(format string, v ...interface{}) {
}

func (l *logger) Debugf(format string, v ...interface{}) {
}

func LogResponse(t *testing.T, resp *resty.Response) {
	t.Logf("Response Status: %v", resp.Status())
	t.Logf("Response Time: %v", resp.Time())
	t.Logf("Response Headers: %v", resp.Header())
	t.Logf("Response Cookies: %v", resp.Cookies())
	t.Logf("Response Body: %v", resp)
}
