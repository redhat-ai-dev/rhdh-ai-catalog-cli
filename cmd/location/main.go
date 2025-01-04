package main

import (
	goflag "flag"
	gin_gonic_http_srv "github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/gin-gonic-http-srv"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	"k8s.io/klog/v2"
	"os"
)

func main() {
	flagset := goflag.NewFlagSet("location", goflag.ContinueOnError)
	klog.InitFlags(flagset)

	//TODO content should be mounted in pod at well known location
	content, err := os.ReadFile("/data/catalog-info-yaml")
	if err != nil {
		klog.Errorf("%s", err.Error())
		os.Exit(-1)
	}
	server := gin_gonic_http_srv.NewImportLocationServer(content)
	stopCh := util.SetupSignalHandler()
	server.Run(stopCh)

}
