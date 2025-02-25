package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub"
	"testing"
)

func TestListResources(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	// Get with no args calls List
	str, err := SetupBackstageTestRESTClient(ts).GetResource()
	stub.AssertError(t, err)
	stub.AssertLineCompare(t, str, stub.Resources, 0)
}

func TestGetResources(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	nsName := "default:phi-mini-instruct"
	str, err := SetupBackstageTestRESTClient(ts).GetResource(nsName)

	stub.AssertError(t, err)
	stub.AssertContains(t, str, []string{nsName})
}

func TestGetResourceError(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).GetResource(nsName)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetResourceWithTags(t *testing.T) {
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
			args: []string{"task-text-generation", "multilingual", "meta", "llm", "llama", "genai", "conversational", "1b"},
			str:  stub.ResourcesFromTagsNoSubset,
		},
		{
			args:   []string{"genai", "meta"},
			subset: true,
			str:    stub.ResourcesFromTags,
		},
	} {
		bs.Subset = tc.subset
		str, err := bs.GetResource(tc.args...)
		stub.AssertError(t, err)
		stub.AssertLineCompare(t, str, tc.str, 0)
	}
}
