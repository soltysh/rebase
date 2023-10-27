package cmd

import (
	"github.com/spf13/cobra"

	"github.com/openshift/rebase/pkg/apply"
	"github.com/openshift/rebase/pkg/options"
)

type ApplyOptions struct {
	options.Common
}

func NewApplyCommand(streams options.IOStreams) *cobra.Command {
	o := &ApplyOptions{Common: options.NewCommon(streams)}

	cmd := &cobra.Command{
		Use:          "apply --repository=/go/src/k8s.io/kubernetes --from=v1.26.0",
		Short:        "Generates log of carry patches from a given version of kubernetes",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Common.Complete(); err != nil {
				return err
			}
			carrieslog := apply.NewApply(o.Common.From, o.Common.RepositoryDir)
			return carrieslog.Run()
		},
	}
	o.Common.AddFlags(cmd.Flags())

	return cmd
}
