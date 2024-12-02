package stub

import (
	"bytes"

	"github.com/spf13/cobra"
)

// ExecuteCommand executes the root command passing the args and returns
// the output as a string and error
func ExecuteCommand(root *cobra.Command, args ...string) (string, string, error) {
	_, stdout, stderr, err := ExecuteCommandC(root, args...)
	return stdout, stderr, err
}

// ExecuteCommandC executes the root command passing the args and returns
// the root command, output as a string and error if any
func ExecuteCommandC(c *cobra.Command, args ...string) (*cobra.Command, string, string, error) {
	obuf := new(bytes.Buffer)
	c.SetOut(obuf)
	ebuf := new(bytes.Buffer)
	c.SetErr(ebuf)
	c.SetArgs(args)
	c.SilenceUsage = true

	root, err := c.ExecuteC()

	return root, obuf.String(), ebuf.String(), err
}
