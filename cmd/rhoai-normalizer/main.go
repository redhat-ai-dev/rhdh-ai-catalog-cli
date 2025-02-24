package main

import (
	"flag"
	rhoai_normalizer "github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/server/rhoai-normalizer"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	_ "net/http/pprof"
	"os"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

var (
	mainLog logr.Logger
)

func main() {
	var pprofAddr string
	flag.StringVar(&pprofAddr, "pprof-address", "6000", "The address the pprof endpoint binds to.")

	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	klog.InitFlags(flag.CommandLine)
	flag.Parse()

	/*
			FYI tracing set set with this zap argument on the deployment (see https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/)
		          args:
		            - -zap-log-level=6
	*/

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)
	mainLog = ctrl.Log.WithName("main")

	ctx := ctrl.SetupSignalHandler()
	restConfig := ctrl.GetConfigOrDie()
	restConfig.QPS = 50
	restConfig.Burst = 50
	var mgr ctrl.Manager
	var err error
	mopts := ctrl.Options{}

	mgr, err = rhoai_normalizer.NewControllerManager(ctx, restConfig, mopts, pprofAddr)
	if err != nil {
		mainLog.Error(err, "unable to start controller-runtime manager")
		os.Exit(1)
	}

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		mainLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		mainLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	mainLog.Info("Starting controller-runtime manager")

	if err = mgr.Start(ctx); err != nil {
		mainLog.Error(err, "problem running controller-runtime manager")
		os.Exit(1)
	}

}
