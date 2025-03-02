package util

import (
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty/v2"
)

func FetchURL(url string) ([]byte, error) {
	rest := resty.New()
	tlsCfg := &tls.Config{InsecureSkipVerify: true}
	rest.SetTLSClientConfig(tlsCfg)
	resp, err := rest.R().Get(url)
	if err != nil {
		return nil, err
	}
	rc := resp.StatusCode()
	if rc != 200 {
		return nil, fmt.Errorf("url %s got rc: %d", url, rc)
	}
	return resp.Body(), nil
}
