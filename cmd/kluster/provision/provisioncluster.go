package provision

import (
	"github.com/jakub-gawlas/kluster/pkg/cluster/kind"
	"github.com/jakub-gawlas/kluster/pkg/kluster"
	"github.com/spf13/cobra"
)

type flagpole struct {
	ConfigPath string
}

func NewCommand() *cobra.Command {
	flags := &flagpole{}
	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provisions local cluster",
		Long:  "Provision local kubernetes cluster without deploy charts and resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags)
		},
	}
	cmd.Flags().StringVar(&flags.ConfigPath, "config", kluster.DefaultConfigPath, "path to a kluster config file")
	return cmd
}

func runE(flags *flagpole) error {
	cfg, err := kluster.LoadConfig(flags.ConfigPath)
	if err != nil {
		return err
	}

	cluster := kind.New(cfg.Name, flags.ConfigPath)
	k, err := kluster.New(cluster, cfg)
	if err != nil {
		return err
	}

	if err := k.Provision(); err != nil {
		return err
	}

	return nil
}
