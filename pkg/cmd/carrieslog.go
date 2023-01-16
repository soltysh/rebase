package cmd

import (
	"github.com/spf13/cobra"

	"github.com/openshift/rebase/pkg/carry"
	"github.com/openshift/rebase/pkg/options"
)

type CarriesLogOptions struct {
	options.Common

	From string
}

func NewCarriesLogCommand(streams options.IOStreams) *cobra.Command {
	o := &CarriesLogOptions{Common: options.NewCommon(streams)}

	cmd := &cobra.Command{
		Use:          "carries-log --repository=/go/src/k8s.io/kubernetes --from=v1.26.0",
		Short:        "Generates log of carry patches from a given version of kubernetes",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Common.Complete(); err != nil {
				return err
			}
			carrieslog := carry.NewLog(o.From, o.Common.RepositoryDir)
			return carrieslog.Run()
		},
	}
	o.Common.AddFlags(cmd.Flags())
	cmd.Flags().StringVar(&o.From, "from", o.From, "Kubernetes version from which to generate carries log")
	cmd.MarkFlagRequired("from")

	return cmd
}
