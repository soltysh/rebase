package options

import (
	"os"

	"github.com/spf13/pflag"
)

// Common provides the standard flags and options used in all commands.
type Common struct {
	IOStreams

	// kubernetes repository directory, as specified by the user or current working dir
	RepositoryDir string
}

func NewCommon(streams IOStreams) Common {
	return Common{
		IOStreams: streams,
	}
}

func (o *Common) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.RepositoryDir, "repository", o.RepositoryDir, "Kubernetes repository directory, or current if none specified")
}

func (o *Common) Complete() error {
	if len(o.RepositoryDir) == 0 {
		var err error
		o.RepositoryDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	return nil
}
