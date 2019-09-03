package kluster

import (
	"os"

	"github.com/jakub-gawlas/kluster/cmd/kluster/kubeconfig"
	"github.com/jakub-gawlas/kluster/cmd/kluster/provision"

	"github.com/jakub-gawlas/kluster/cmd/kluster/deploy"
	"github.com/jakub-gawlas/kluster/cmd/kluster/destroy"
	"github.com/jakub-gawlas/kluster/cmd/kluster/kubectl"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kluster",
		Short: "Kluster is tool for provision local Kubernetes cluster",
		Long:  "Kluster using KinD (Kubernetes in Docker) for provision local cluster and helm charts for deploy applications",
	}
	cmd.AddCommand(provision.NewCommand())
	cmd.AddCommand(deploy.NewCommand())
	cmd.AddCommand(destroy.NewCommand())
	cmd.AddCommand(kubectl.NewCommand())
	cmd.AddCommand(kubeconfig.NewCommand())
	return cmd
}

func Run() error {
	return NewCommand().Execute()
}

func Main() {
	if err := Run(); err != nil {
		os.Exit(1)
	}
}
