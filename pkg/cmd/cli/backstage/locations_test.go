package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/common"
	"testing"
)

func TestListLocations(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	str, err := SetupBackstageTestRESTClient(ts).ListLocations()
	common.AssertError(t, err)
	common.AssertEqual(t, common.TestJSONStringIndented, str)
}

func TestGetLocations(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	key := "key1"
	str, err := SetupBackstageTestRESTClient(ts).GetLocation(key)
	common.AssertError(t, err)
	common.AssertContains(t, str, []string{key})
}

func TestGetLocationsError(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).GetLocation(nsName)
	if err == nil {
		t.Error("expected error")
	}
}

func TestImportLocation(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	arg := "https://my-repo/my.yaml"
	retJSON, err := SetupBackstageTestRESTClient(ts).ImportLocation(arg)
	common.AssertError(t, err)
	str, err := SetupBackstageTestRESTClient(ts).PrintImportLocation(retJSON)
	common.AssertError(t, err)
	common.AssertContains(t, str, []string{arg})
}

func TestImportLocationError(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	arg := ":"
	_, err := SetupBackstageTestRESTClient(ts).ImportLocation(arg)
	if err == nil {
		t.Error("expected error")
	}
}

func TestDeleteLocation(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	arg := "my-location-id"
	str, err := SetupBackstageTestRESTClient(ts).DeleteLocation(arg)
	common.AssertError(t, err)
	common.AssertContains(t, str, []string{arg})
}

func TestDeleteLocationsError(t *testing.T) {
	ts := backstage.CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).DeleteLocation(nsName)
	if err == nil {
		t.Error("expected error")
	}
}
