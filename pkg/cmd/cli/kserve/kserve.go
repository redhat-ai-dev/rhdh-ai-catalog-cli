package kserve

import (
	"context"
	"fmt"
	serverapiv1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/cmd/cli/kserve"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/config"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/util"
	"github.com/spf13/cobra"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"strings"
)

const (
	kserveExamples = `
# Both Owner and Lifecycle are required parameters.  Examine Backstage Catalog documentation for details.
# This will query all the InferenceService instances in the current namespace and build Catalog Component, Resource, and
# API Entities from the data.
$ %s new-model kserve <Owner> <Lifecycle> <args...>

# This example shows the flags that will set the URL, Token, and Skip TLS when accessing the cluster for InferenceService instances 
$ %s new-model kserve <Owner> <Lifecycle> --model-metadata-url=https://my-kubeflow.com --model-metadata-token=my-token --model-metadata-skip-tls=true

# The '--kubeconfig' flag can be used to set the location of the configuration file used for accessing the credentials
# used for interacting with the Kubernetes cluster.
$ %s new-model kserve <Owner> <Lifecycle> --kubeconfig=/home/myid/my-kube.json

# This form will pull in only the InferenceService instances with the names 'inferenceservice1' and 'inferenceservice2'
# in the 'my-datascience-project'namespace in order to build Catalog Component, Resource, and API Entities.
$ %s new-model kserve Owner Lifecycle inferenceservice1 inferenceservice2 --namespace my-datascience-project
`
	sklearn     = "sklearn"
	xgboost     = "xgboost"
	tensorflow  = "tensorflow"
	pytorch     = "pytorch"
	triton      = "triton"
	onnx        = "onnx"
	huggingface = "huggingface"
	pmml        = "pmml"
	lightgbm    = "lightgbm"
	paddle      = "paddle"
)

type CommonPopulator struct {
	Owner     string
	Lifecycle string
	InferSvc  *serverapiv1beta1.InferenceService
}

func NewCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kserve",
		Short:   "KServe related API",
		Long:    "Interact with KServe related instances on a K8s cluster to manage AI related catalog entities in a Backstage instance.",
		Example: strings.ReplaceAll(kserveExamples, "%s", util.ApplicationName),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := []string{}

			if len(args) < 2 {
				err := fmt.Errorf("need to specify an Owner and Lifecycle setting")
				klog.Errorf("%s", err.Error())
				klog.Flush()
				return err
			}
			owner := args[0]
			lifecycle := args[1]

			if len(args) > 2 {
				ids = args[2:]
			}

			kserve.SetupKServeClient(cfg)
			namespace := cfg.Namespace
			servingClient := cfg.ServingClient

			if len(ids) != 0 {
				for _, id := range ids {
					is, err := servingClient.InferenceServices(namespace).Get(context.Background(), id, metav1.GetOptions{})
					if err != nil {
						klog.Errorf("inference service retrieval error for %s:%s: %s", namespace, id, err.Error())
						klog.Flush()
						return err
					}

					err = CallBackstagePrinters(owner, lifecycle, is, cmd.OutOrStdout())
					if err != nil {
						return err
					}
				}
			} else {
				isl, err := servingClient.InferenceServices(namespace).List(context.Background(), metav1.ListOptions{})
				if err != nil {
					klog.Errorf("inference service retrieval error for %s: %s", namespace, err.Error())
					klog.Flush()
					return err
				}
				for _, is := range isl.Items {
					err = CallBackstagePrinters(owner, lifecycle, &is, cmd.OutOrStdout())
					if err != nil {
						klog.Errorf("%s", err.Error())
						klog.Flush()
						return err
					}
				}
			}
			return nil
		},
	}

	return cmd
}

func CallBackstagePrinters(owner, lifecycle string, is *serverapiv1beta1.InferenceService, writer io.Writer) error {
	compPop := kserve.ComponentPopulator{}
	compPop.Owner = owner
	compPop.Lifecycle = lifecycle
	compPop.InferSvc = is
	err := backstage.PrintComponent(&compPop, writer)
	if err != nil {
		return err
	}

	resPop := kserve.ResourcePopulator{}
	resPop.Owner = owner
	resPop.Lifecycle = lifecycle
	resPop.InferSvc = is
	err = backstage.PrintResource(&resPop, writer)
	if err != nil {
		return err
	}

	apiPop := kserve.ApiPopulator{}
	apiPop.Owner = owner
	apiPop.Lifecycle = lifecycle
	apiPop.InferSvc = is
	err = backstage.PrintAPI(&apiPop, writer)
	return err
}
