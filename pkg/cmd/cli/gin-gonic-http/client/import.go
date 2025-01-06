package client

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
)

func (a *Artifacts) Import() {
	util.ProcessOutput(backstage.SetupBackstageRESTClient(a.cfg).ImportLocation(a.routeURL))
}
