package kubectl

import (
	"os"
	"os/exec"
)

type Client struct {
	kubeconfig string
}

const (
	kubectlCmd = "kubectl"
)

func New(kubeconfigPath string) *Client {
	return &Client{
		kubeconfig: kubeconfigPath,
	}
}

func (cli *Client) Exec(arg ...string) error {
	cmd := exec.Command(kubectlCmd, arg...)
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
