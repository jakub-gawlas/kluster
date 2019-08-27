package kubectl

import (
	"bytes"
	"fmt"
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
	var stderr bytes.Buffer
	cmd := exec.Command(kubectlCmd, arg...)
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}
	return nil
}

func (cli *Client) ExecStdinData(data []byte, arg ...string) error {
	var stderr bytes.Buffer
	cmd := exec.Command(kubectlCmd, arg...)
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewReader(data)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}
	return nil
}
