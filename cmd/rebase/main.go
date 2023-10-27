package main

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/openshift/rebase/pkg/cmd"
	"github.com/openshift/rebase/pkg/options"
)

func main() {
	flags := pflag.NewFlagSet("rebase", pflag.ExitOnError)
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
		Short:        "OpenShift helper tool for performing automatic kubernetes updates",
		SilenceUsage: true,
	}
	streams := options.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	command.AddCommand(cmd.NewCarriesCommand(streams))
	command.AddCommand(cmd.NewApplyCommand(streams))

	logging := flag.NewFlagSet("logging", flag.ContinueOnError)
	klog.InitFlags(logging)
	if vFlag := logging.Lookup("v"); vFlag != nil {
		pf := pflag.PFlagFromGoFlag(vFlag)
		command.PersistentFlags().AddFlag(pf)
	}

	return command
}
