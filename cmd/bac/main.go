package main

import (
	goflag "flag"
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	"github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
	"os"
)

var hiddenLogFlags = []string{
	"add_dir_header",
	"alsologtostderr",
	"log_backtrace_at",
	"log_dir",
	"log_file",
	"log_file_max_size",
	"logtostderr",
	"one_output",
	"skip_headers",
	"skip_log_headers",
	"stderrthreshold",
	"v",
	"vmodule",
}

func main() {
	if err := initGoFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	initPFlags()

	rootCmd := cli.NewCmd()
	if err := rootCmd.Execute(); err != nil {
		klog.Errorf("ERROR: %v\n", err)
		os.Exit(1)
	}
}

// initGoFlags initializes the flag sets for klog.
// Any flags for "-h" or "--help" are ignored because pflag will show the usage later with all subcommands.
func initGoFlags() error {
	flagset := goflag.NewFlagSet(util.ApplicationName, goflag.ContinueOnError)
	goflag.CommandLine = flagset
	klog.InitFlags(flagset)

	args := []string{}
	for _, arg := range os.Args[1:] {
		if arg != "-h" && arg != "--help" {
			args = append(args, arg)
		}
	}
	return flagset.Parse(args)
}

// initPFlags initializes the pflags used by Cobra subcommands.
func initPFlags() {
	flags := pflag.NewFlagSet(util.ApplicationName, pflag.ExitOnError)
	flags.AddGoFlagSet(goflag.CommandLine)
	pflag.CommandLine = flags

	for _, flag := range hiddenLogFlags {
		if err := flags.MarkHidden(flag); err != nil {
			panic(err)
		}
	}
}
