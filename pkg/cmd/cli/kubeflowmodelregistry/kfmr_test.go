package kubeflowmodelregistry

import (
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/config"
	cobra2 "github.com/redhat-ai-dev/model-catalog-bridge/test/cobra"
	"github.com/redhat-ai-dev/model-catalog-bridge/test/stub/common"
	"github.com/redhat-ai-dev/model-catalog-bridge/test/stub/kfmr"
	"github.com/spf13/cobra"
	"strings"
	"testing"
)

func TestNewCmd(t *testing.T) {
	ts := kfmr.CreateGetServer(t)
	defer ts.Close()
	for _, tc := range []struct {
		args           []string
		generatesError bool
		generatesHelp  bool
		errorStr       string
		// we do output compare in chunks as ranges over the components status map are non-deterministic wrt order
		outStr []string
	}{
		{
			args:          []string{"--help"},
			generatesHelp: true,
		},
		{
			args:           []string{},
			generatesError: true,
			errorStr:       "need to specify an Owner and Lifecycle setting",
		},
		{
			args:           []string{"Owner"},
			generatesError: true,
			errorStr:       "need to specify an Owner and Lifecycle setting",
		},
		{
			args:   []string{"Owner", "Lifecycle"},
			outStr: []string{listOutput},
		},
		{
			args:   []string{"Owner", "Lifecycle", "1"},
			outStr: []string{listOutput},
		},
	} {
		cfg := &config.Config{}
		kfmr.SetupKubeflowTestRESTClient(ts, cfg)
		cmd := NewCmd(cfg)
		subCmd, stdout, stderr, err := cobra2.ExecuteCommandC(cmd, tc.args...)
		switch {
		case err == nil && tc.generatesError:
			t.Errorf("error should have been generated for '%s'", strings.Join(tc.args, " "))
		case err != nil && !tc.generatesError:
			t.Errorf("error generated unexpectedly for '%s': %s", strings.Join(tc.args, " "), err.Error())
		case err != nil && tc.generatesError && !strings.Contains(stderr, tc.errorStr):
			t.Errorf("unexpected error output for '%s'- got '%s' but expected '%s'", strings.Join(tc.args, " "), stderr, tc.errorStr)
		case tc.generatesHelp && !testHelpOK(stdout, subCmd):
			t.Errorf("unexpected help output for '%s' - got '%s' but expected '%s'", strings.Join(tc.args, " "), stdout, subCmd.Long)
		case err == nil && !tc.generatesError:
			for _, str := range tc.outStr {
				common.AssertLineCompare(t, stdout, str, 0)
			}
		}

	}
}

func testHelpOK(stdout string, cmd *cobra.Command) bool {
	if strings.Contains(stdout, cmd.Long) {
		return true
	}
	return false
}

const (
	listOutput = `apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  annotations:
    backstage.io/techdocs-ref: ./
  description: dummy model 1
  links:
  - icon: WebAsset
    title: version 1
    type: website
    url: https://foo.com
  name: model-1
  tags:
  - foo-bar
spec:
  dependsOn:
  - resource:v1
  - api:model-1-v1-artifact
  lifecycle: Lifecycle
  owner: user:Owner
  profile:
    displayName: model-1
  type: model-server
---
apiVersion: backstage.io/v1alpha1
kind: Resource
metadata:
  annotations:
    backstage.io/techdocs-ref: resource/
  description: dummy model 1
  links:
  - icon: WebAsset
    title: version 1
    type: website
    url: https://foo.com
  name: v1
spec:
  dependencyOf:
  - component:model-1
  lifecycle: Lifecycle
  owner: user:Owner
  profile:
    displayName: v1
  type: ai-model
---
apiVersion: backstage.io/v1alpha1
kind: API
metadata:
  annotations:
    backstage.io/techdocs-ref: api/
  description: dummy model 1
  name: model-1
spec:
  definition: no-definition-yet
  dependencyOf:
  - component:model-1
  lifecycle: Lifecycle
  owner: user:Owner
  profile:
    displayName: model-1
  type: unknown
`
)
