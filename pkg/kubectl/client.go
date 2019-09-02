package kubectl

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

type Client struct {
	kubeconfig string
}

const (
	kubectlCmd = "kubectl"
)

var execCommand = exec.Command

func New(kubeconfigPath string) *Client {
	return &Client{
		kubeconfig: kubeconfigPath,
	}
}

func (cli *Client) ExecStdInOut(stdin io.Reader, stdout io.Writer, arg ...string) error {
	var stderr bytes.Buffer
	cmd := execCommand(kubectlCmd, arg...)
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}
	return nil
}

func (cli *Client) ExecInData(data []byte, arg ...string) ([]byte, error) {
	var stdout bytes.Buffer
	if err := cli.ExecStdInOut(bytes.NewReader(data), &stdout, arg...); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}

func (cli *Client) Exec(arg ...string) ([]byte, error) {
	return cli.ExecInData(nil, arg...)
}
