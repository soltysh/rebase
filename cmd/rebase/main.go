package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/openshift/rebase/pkg/cmd"
)

func main() {
	flags := pflag.NewFlagSet("rebase-helper", pflag.ExitOnError)
	pflag.CommandLine = flags

	klog.InitFlags(nil)

	root := NewRootCommand()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:          "rebase",
		SilenceUsage: true,
	}
	command.AddCommand(cmd.NewApplyCommand())
	command.AddCommand(cmd.NewVerifyCommand())
	command.AddCommand(cmd.NewCopyCommand())

	return command
}
