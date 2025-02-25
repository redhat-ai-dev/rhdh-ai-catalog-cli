package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub"
	"testing"
)

func TestListComponents(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	// Get with no args calls List
	str, err := SetupBackstageTestRESTClient(ts).GetComponent()
	stub.AssertError(t, err)
	stub.AssertLineCompare(t, str, stub.Components, 0)
}

func TestGetComponents(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	nsName := "default:ollama-service-component"
	str, err := SetupBackstageTestRESTClient(ts).GetComponent(nsName)

	stub.AssertError(t, err)
	stub.AssertContains(t, str, []string{nsName})
}

func TestGetComponentsError(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).GetComponent(nsName)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetComponentsWithTags(t *testing.T) {
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
			args: []string{"gateway", "authenticated", "developer-model-service", "llm", "vllm", "ibm-granite", "genai"},
			str:  stub.ComponentsFromTagsNoSubset,
		},
		{
			args:   []string{"genai"},
			subset: true,
			str:    stub.ComponentsFromTags,
		},
	} {
		bs.Subset = tc.subset
		str, err := bs.GetComponent(tc.args...)
		stub.AssertError(t, err)
		stub.AssertLineCompare(t, str, tc.str, 0)
	}
}
