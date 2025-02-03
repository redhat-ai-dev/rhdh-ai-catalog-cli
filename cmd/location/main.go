package main

import (
	goflag "flag"
	gin_gonic_http_srv "github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/gin-gonic-http/server"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	"io/fs"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	flagset := goflag.NewFlagSet("location", goflag.ContinueOnError)
	klog.InitFlags(flagset)

	content := map[string][]byte{}
	err := filepath.Walk("/data", func(path string, info fs.FileInfo, err error) error {
		if strings.Contains(info.Name(), "_") {
			c := []byte{}
			klog.Infoln("processing configmap file " + info.Name())
			if info.IsDir() {
				return nil
			}
			if strings.HasPrefix(info.Name(), "..") {
				klog.Infof("skipping file starting with ..: %s", info.Name())
			}
			fullName := "/data/" + info.Name()
			klog.Infof("reading file %s", fullName)
			c, err = os.ReadFile(fullName)
			if err != nil {
				klog.Errorf("%s", err.Error())
				klog.Flush()
				return err
			}
			klog.Infof("adding file %s with content len %d to list", info.Name(), len(c))
			content[info.Name()] = c
		}
		return nil
	})
	if err != nil {
		klog.Errorf("%s", err.Error())
		os.Exit(-1)
	}
	server := gin_gonic_http_srv.NewImportLocationServer(content)
	stopCh := util.SetupSignalHandler()
	server.Run(stopCh)

}
