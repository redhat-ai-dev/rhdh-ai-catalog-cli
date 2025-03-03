package kubeflowmodelregistry

import (
	"fmt"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/cmd/cli/kubeflowmodelregistry"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"strings"
)

const (
	kubeflowExample = `
# Both Owner and Lifecycle are required parameters.  Examine Backstage Catalog documentation for details.
# This will query all the RegisteredModel, ModelVersion, ModelArtifact, and InferenceService instances in the Kubeflow Model Registry and build Catalog Component, Resource, and
# API Entities from the data.
$ %s new-model kubeflow <Owner> <Lifecycle> <args...>

# This will set the URL, Token, and Skip TLS when accessing Kubeflow
$ %s new-model kubeflow <Owner> <Lifecycle> --model-metadata-url=https://my-kubeflow.com --model-metadata-token=my-token --model-metadata-skip-tls=true

# This form will pull in only the RegisteredModels with the specified IDs '1' and '2' and the ModelVersion, ModelArtifact, and InferenceService
# artifacts that are linked to those RegisteredModels in order to build Catalog Component, Resource, and API Entities.
$ %s new-model kubeflow <Owner> <Lifecycle> 1 2 
`

	// pulled from makeValidator.ts in the catalog-model package in core backstage
	tagRegexp = "^[a-z0-9:+#]+(\\-[a-z0-9:+#]+)*$"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kubeflow",
		Aliases: []string{"kf"},
		Short:   "Kubeflow Model Registry related API",
		Long:    "Interact with the Kubeflow Model Registry REST API as part of managing AI related catalog entities in a Backstage instance.",
		Example: strings.ReplaceAll(kubeflowExample, "%s", util.ApplicationName),
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

			kfmr := kubeflowmodelregistry.SetupKubeflowRESTClient(cfg)

			_, _, err := kubeflowmodelregistry.LoopOverKFMR(owner, lifecycle, ids, cmd.OutOrStdout(), kfmr, nil)
			return err

		},
	}

	return cmd
}
