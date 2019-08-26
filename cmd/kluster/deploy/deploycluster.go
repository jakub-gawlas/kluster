package deploy

import (
	"github.com/jakub-gawlas/kluster/pkg/kluster"
	"github.com/spf13/cobra"
)

type flagpole struct {
	ConfigPath string
}

func NewCommand() *cobra.Command {
	flags := &flagpole{}
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploys local cluster",
		Long:  "Deploys local kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags)
		},
	}
	cmd.Flags().StringVar(&flags.ConfigPath, "config", kluster.DefaultConfigPath, "path to a kluster config file")
	return cmd
}

func runE(flags *flagpole) error {
	k, err := kluster.New(flags.ConfigPath)
	if err != nil {
		return err
	}
	return k.Deploy()
}
