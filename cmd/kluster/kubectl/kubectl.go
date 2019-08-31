package kubectl

import (
	"os"

	"github.com/jakub-gawlas/kluster/pkg/cluster/kind"
	"github.com/jakub-gawlas/kluster/pkg/kluster"
	"github.com/jakub-gawlas/kluster/pkg/kubectl"
	"github.com/spf13/cobra"
)

type flagpole struct {
	ConfigPath string
}

func NewCommand() *cobra.Command {
	flags := &flagpole{}
	cmd := &cobra.Command{
		Use:   "kubectl",
		Short: "Executes kubectl command",
		Long:  "Executes kubectl command on cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, args)
		},
		DisableFlagParsing: true,
	}
	cmd.Flags().StringVar(&flags.ConfigPath, "config", kluster.DefaultConfigPath, "path to a kluster config file")
	return cmd
}

func runE(flags *flagpole, args []string) error {
	cfg, err := kluster.LoadConfig(flags.ConfigPath)
	if err != nil {
		return err
	}

	cluster := kind.New(cfg.Name, flags.ConfigPath)
	k, err := kluster.New(cluster, cfg)
	if err != nil {
		return err
	}

	kubeconfig, err := k.Cluster().KubeConfigPath()
	if err != nil {
		return err
	}

	kube := kubectl.New(kubeconfig)
	res, err := kube.Exec(args...)
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(res); err != nil {
		return err
	}

	return nil
}
