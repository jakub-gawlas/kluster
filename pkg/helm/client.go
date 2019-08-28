package helm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/jakub-gawlas/kluster/pkg/kubectl"
)

type Client struct {
	kubeconfig string
}

const (
	helmCmd              = "helm"
	installMaxRetries    = 10
	installRetryInterval = time.Second * 15
)

var execCommand = exec.Command

func New(kubeconfigPath string) *Client {
	return &Client{
		kubeconfig: kubeconfigPath,
	}
}

func (cli *Client) Init() error {
	var stderr bytes.Buffer
	cmd := execCommand(helmCmd, "init")
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrap(fmt.Errorf(stderr.String()), "helm init")
	}

	k := kubectl.New(cli.kubeconfig)
	if err := k.Exec("create", "clusterrolebinding", "add-on-cluster-admin", "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default"); err != nil {
		return errors.Wrap(err, "create helm role")
	}

	return nil
}

func (cli *Client) Upgrade(name, path string, sets map[string]string) error {
	args := []string{"upgrade", name, path}
	if len(sets) > 0 {
		args = append(args, "--set", createSet(sets))
	}

	var stderr bytes.Buffer
	cmd := execCommand(helmCmd, args...)
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}

	return nil
}

func (cli *Client) Install(name, path string, sets map[string]string) error {
	i := 0
	for {
		err := cli.install(name, path, sets)
		if err == nil {
			return nil
		}
		if i > installMaxRetries {
			return err
		}
		i++
		time.Sleep(installRetryInterval)
	}
}

func (cli *Client) install(name, path string, sets map[string]string) error {
	var stderr bytes.Buffer
	cmd := exec.Command(helmCmd, "install", "--name", name, "--set", createSet(sets), path)
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}

	return nil
}

func createSet(sets map[string]string) string {
	vv := make([]string, 0, len(sets))
	for k, v := range sets {
		v := k + "=" + v
		vv = append(vv, v)
	}
	return strings.Join(vv, ",")
}
