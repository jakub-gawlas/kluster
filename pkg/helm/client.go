package helm

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	kubectl    KubectlExecuter
	kubeconfig string
}

type KubectlExecuter interface {
	Exec(...string) ([]byte, error)
}

const (
	helmCmd              = "helm"
	installMaxRetries    = 10
	installRetryInterval = time.Second * 15
)

func New(kubectl KubectlExecuter, kubeconfigPath string) *Client {
	return &Client{
		kubectl:    kubectl,
		kubeconfig: kubeconfigPath,
	}
}

var execCommand = exec.Command

func (cli *Client) Init() error {
	var stderr bytes.Buffer
	cmd := execCommand(helmCmd, "init")
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrap(fmt.Errorf(stderr.String()), "helm init")
	}

	if _, err := cli.kubectl.Exec("create", "clusterrolebinding", "add-on-cluster-admin", "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default"); err != nil {
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
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}

	return nil
}

func (cli *Client) Install(name, path string, sets map[string]string, retryOpts ...RetryOption) error {
	retry := &RetryOptions{
		MaxRetries: installMaxRetries,
		Interval:   installRetryInterval,
	}
	for _, opt := range retryOpts {
		opt(retry)
	}

	i := 0
	for {
		err := cli.install(name, path, sets)
		if err == nil {
			return nil
		}
		if i >= retry.MaxRetries {
			return err
		}
		i++
		time.Sleep(retry.Interval)
	}
}

func (cli *Client) install(name, path string, sets map[string]string) error {
	args := []string{"install", "--name", name, path}
	if len(sets) > 0 {
		args = append(args, "--set", createSet(sets))
	}

	var stderr bytes.Buffer
	cmd := execCommand(helmCmd, args...)
	cmd.Env = []string{"KUBECONFIG=" + cli.kubeconfig}
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
	sort.Slice(vv, func(i, j int) bool {
		return vv[i] < vv[j]
	})

	return strings.Join(vv, ",")
}

type RetryOptions struct {
	MaxRetries int
	Interval   time.Duration
}

type RetryOption func(*RetryOptions)

func WithMaxRetries(maxRetries int) RetryOption {
	return func(opts *RetryOptions) {
		opts.MaxRetries = maxRetries
	}
}

func WithInterval(interval time.Duration) RetryOption {
	return func(opts *RetryOptions) {
		opts.Interval = interval
	}
}
