package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub"
	"testing"
)

func TestListLocations(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	str, err := SetupBackstageTestRESTClient(ts).ListLocations()
	stub.AssertError(t, err)
	stub.AssertEqual(t, stub.TestJSONStringIndented, str)
}

func TestGetLocations(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	key := "key1"
	str, err := SetupBackstageTestRESTClient(ts).GetLocation(key)
	stub.AssertError(t, err)
	stub.AssertContains(t, str, []string{key})
}

func TestGetLocationsError(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).GetLocation(nsName)
	if err == nil {
		t.Error("expected error")
	}
}

func TestImportLocation(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	arg := "https://my-repo/my.yaml"
	retJSON, err := SetupBackstageTestRESTClient(ts).ImportLocation(arg)
	stub.AssertError(t, err)
	str, err := SetupBackstageTestRESTClient(ts).PrintImportLocation(retJSON)
	stub.AssertError(t, err)
	stub.AssertContains(t, str, []string{arg})
}

func TestImportLocationError(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	arg := ":"
	_, err := SetupBackstageTestRESTClient(ts).ImportLocation(arg)
	if err == nil {
		t.Error("expected error")
	}
}

func TestDeleteLocation(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	arg := "my-location-id"
	str, err := SetupBackstageTestRESTClient(ts).DeleteLocation(arg)
	stub.AssertError(t, err)
	stub.AssertContains(t, str, []string{arg})
}

func TestDeleteLocationsError(t *testing.T) {
	ts := stub.CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).DeleteLocation(nsName)
	if err == nil {
		t.Error("expected error")
	}
}
