package util

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/yaml"
)

const ApplicationName = "bac"

func PrintYaml(obj interface{}, addDivider bool, cmd *cobra.Command) error {
	writer := printers.GetNewTabWriter(cmd.OutOrStdout())
	output, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = writer.Write(output)
	if addDivider {
		fmt.Fprintln(cmd.OutOrStdout(), "---")
	}
	return err
}
