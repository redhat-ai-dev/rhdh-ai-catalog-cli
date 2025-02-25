package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub"
	"testing"
)

func TestListAPIs(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	// Get with no args calls List
	str, err := SetupBackstageTestRESTClient(ts).GetAPI()

	stub.AssertError(t, err)
	stub.AssertLineCompare(t, str, stub.Apis, 0)
}

func TestGetAPIs(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	nsName := "default:ollama-service-api"
	str, err := SetupBackstageTestRESTClient(ts).GetAPI(nsName)

	stub.AssertError(t, err)
	stub.AssertContains(t, str, []string{nsName})
}

func TestGetAPIsError(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).GetAPI(nsName)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetAPIsWithTags(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	bs := SetupBackstageTestRESTClient(ts)
	bs.Tags = true

	for _, tc := range []struct {
		args   []string
		str    string
		subset bool
	}{
		{
			args: []string{"genai", "meta"},
		},
		{
			args: []string{"vllm", "api", "openai"},
			str:  stub.ApisFromTagsNoSubset,
		},
		{
			args:   []string{"vllm"},
			subset: true,
			str:    stub.ApisFromTags,
		},
	} {
		bs.Subset = tc.subset
		str, err := bs.GetAPI(tc.args...)
		stub.AssertError(t, err)
		stub.AssertLineCompare(t, str, tc.str, 0)
	}
}
