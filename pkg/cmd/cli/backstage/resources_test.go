package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/common"
	"testing"
)

func TestListResources(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	// Get with no args calls List
	str, err := SetupBackstageTestRESTClient(ts).GetResource()
	common.AssertError(t, err)
	common.AssertLineCompare(t, str, common.Resources, 0)
}

func TestGetResources(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	nsName := "default:phi-mini-instruct"
	str, err := SetupBackstageTestRESTClient(ts).GetResource(nsName)

	common.AssertError(t, err)
	common.AssertContains(t, str, []string{nsName})
}

func TestGetResourceError(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).GetResource(nsName)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetResourceWithTags(t *testing.T) {
	ts := backstage.CreateServer(t)
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
			str:  common.ResourcesFromTagsNoSubset,
		},
		{
			args:   []string{"genai", "meta"},
			subset: true,
			str:    common.ResourcesFromTags,
		},
	} {
		bs.Subset = tc.subset
		str, err := bs.GetResource(tc.args...)
		common.AssertError(t, err)
		common.AssertLineCompare(t, str, tc.str, 0)
	}
}
