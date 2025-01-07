package client

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
)

func (a *Artifacts) Import() {
	//TODO - there seems to be a delay between our import here and the catalog entries becoming visible from the UI or CLI.  Also, the location ID does not show up in the entity dump, but the uid or name does not work with 'bac delete-model ...'.  You have to save the location ID returned from the original import and use that to delete once the entries show up.
	util.ProcessOutput(backstage.SetupBackstageRESTClient(a.cfg).ImportLocation(a.routeURL))
}
