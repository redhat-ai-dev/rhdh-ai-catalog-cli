package cli

import (
	cobra2 "github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/cobra"
	"github.com/spf13/cobra"
	"strings"
	"testing"
)

func TestNewCmd(t *testing.T) {
	cmd := NewCmd()

	for _, tc := range []struct {
		args           []string
		generatesError bool
		generatesHelp  bool
		errorStr       string
		outStr         string
	}{
		{
			args:          []string{"new-model"},
			generatesHelp: true,
		},
		{
			args:           []string{"new-model", "kserve"},
			generatesError: true,
			errorStr:       "need to specify an Owner and Lifecycle setting",
		},
		{
			args:           []string{"new-model", "kubeflow"},
			generatesError: true,
			errorStr:       "need to specify an Owner and Lifecycle setting",
		},
		{
			args:          []string{"new-model", "help", "kserve"},
			generatesHelp: true,
		},
		{
			args:          []string{"new-model", "help", "kubeflow"},
			generatesHelp: true,
		},
		{
			args:          []string{"get"},
			generatesHelp: true,
		},
		{
			args:          []string{"get", "help"},
			generatesHelp: true,
		},
		{
			args:          []string{"get", "help", "location"},
			generatesHelp: true,
		},
		{
			args:           []string{"get", "location", "help"},
			generatesError: true,
			errorStr:       "unsupported protocol scheme",
		},
		{
			args:          []string{"get", "help"},
			generatesHelp: true,
		},
		{
			args:          []string{"get", "help", "locations"},
			generatesHelp: true,
		},
		{
			args:           []string{"get", "locations", "foo"},
			generatesError: true,
			errorStr:       "unsupported protocol scheme",
		},
		{
			args:          []string{"get", "help", "components"},
			generatesHelp: true,
		},
		{
			args:           []string{"get", "components", "foo"},
			generatesError: true,
			errorStr:       "unsupported protocol scheme",
		},
		{
			args:          []string{"get", "help", "resources"},
			generatesHelp: true,
		},
		{
			args:           []string{"get", "resources", "help"},
			generatesError: true,
			errorStr:       "unsupported protocol scheme",
		},
		{
			args:          []string{"get", "help", "apis"},
			generatesHelp: true,
		},
		{
			args:           []string{"get", "apis", "foo"},
			generatesError: true,
			errorStr:       "unsupported protocol scheme",
		},
		{
			args:          []string{"get", "help", "entities"},
			generatesHelp: true,
		},
	} {
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
		}
	}
}

func testHelpOK(stdout string, cmd *cobra.Command) bool {
	if strings.Contains(stdout, cmd.Long) {
		return true
	}
	return false
}
