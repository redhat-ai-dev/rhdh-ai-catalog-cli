package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/common"
	"testing"
)

func TestListEntities(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	str, err := SetupBackstageTestRESTClient(ts).ListEntities()
	common.AssertError(t, err)
	common.AssertEqual(t, common.TestJSONStringIndented, str)
}
