package helm

import (
	"os"
	"os/exec"
	"strings"
	"time"

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

func New(kubeconfigPath string) *Client {
	return &Client{
		kubeconfig: kubeconfigPath,
	}
}

func (cli *Client) Init() error {
	cmd := exec.Command(helmCmd, "init")
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	k := kubectl.New(cli.kubeconfig)
	return k.Exec("create", "clusterrolebinding", "add-on-cluster-admin", "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default")
}

func (cli *Client) Upgrade(name, path string, sets map[string]string) error {
	cmd := exec.Command(helmCmd, "upgrade", "--set", createSet(sets), name, path)
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
	cmd := exec.Command(helmCmd, "install", "--name", name, "--set", createSet(sets), path)
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createSet(sets map[string]string) string {
	vv := make([]string, 0, len(sets))
	for k, v := range sets {
		v := k + "=" + v
		vv = append(vv, v)
	}
	return strings.Join(vv, ",")
}
