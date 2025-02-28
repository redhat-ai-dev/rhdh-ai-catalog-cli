package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/common"
	"testing"
)

func TestListComponents(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	// Get with no args calls List
	str, err := SetupBackstageTestRESTClient(ts).GetComponent()
	common.AssertError(t, err)
	common.AssertLineCompare(t, str, common.Components, 0)
}

func TestGetComponents(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	nsName := "default:ollama-service-component"
	str, err := SetupBackstageTestRESTClient(ts).GetComponent(nsName)

	common.AssertError(t, err)
	common.AssertContains(t, str, []string{nsName})
}

func TestGetComponentsError(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).GetComponent(nsName)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetComponentsWithTags(t *testing.T) {
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
			args: []string{"gateway", "authenticated", "developer-model-service", "llm", "vllm", "ibm-granite", "genai"},
			str:  common.ComponentsFromTagsNoSubset,
		},
		{
			args:   []string{"genai"},
			subset: true,
			str:    common.ComponentsFromTags,
		},
	} {
		bs.Subset = tc.subset
		str, err := bs.GetComponent(tc.args...)
		common.AssertError(t, err)
		common.AssertLineCompare(t, str, tc.str, 0)
	}
}
