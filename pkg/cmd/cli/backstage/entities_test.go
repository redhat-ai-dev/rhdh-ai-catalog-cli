package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub"
	"testing"
)

func TestListEntities(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	str, err := SetupBackstageTestRESTClient(ts).ListEntities()
	stub.AssertError(t, err)
	stub.AssertEqual(t, stub.TestJSONStringIndented, str)
}
