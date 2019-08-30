package kluster

import (
	"fmt"

	"github.com/jakub-gawlas/kluster/pkg/cluster"

	"github.com/jakub-gawlas/kluster/pkg/yaml/resolver"
	"github.com/pkg/errors"

	"github.com/jakub-gawlas/kluster/pkg/kubectl"

	"github.com/jakub-gawlas/kluster/pkg/helm"
)

type Kluster struct {
	cluster        cluster.Cluster
	cfg            *Config
	kubeconfigPath string
}

func New(cluster cluster.Cluster, cfg *Config) (*Kluster, error) {
	return &Kluster{
		cluster: cluster,
		cfg:     cfg,
	}, nil
}

func (k Kluster) KubeconfigPath() string {
	if k.kubeconfigPath != "" {
		return k.kubeconfigPath
	}
	path, err := k.cluster.KubeConfigPath()
	if err != nil {
		panic("cannot get kubeconfig path to local cluster:" + err.Error())
	}
	k.kubeconfigPath = path
	return path
}

func (k Kluster) Cluster() cluster.Cluster {
	return k.cluster
}

func (k Kluster) Deploy() error {
	exists, err := k.cluster.Exists()
	if err != nil {
		return err
	}

	if !exists {
		if err := k.cluster.Create(); err != nil {
			return err
		}
		kube := kubectl.New(k.KubeconfigPath())
		h := helm.New(kube, k.KubeconfigPath())
		if err := h.Init(); err != nil {
			return err
		}
	}

	if err := k.deployResources(); err != nil {
		return errors.Wrap(err, "deploy resources")
	}

	for _, chart := range k.cfg.Charts {
		if err := chart.Deploy(k.cluster, exists); err != nil {
			return err
		}
	}

	return nil
}

func (k Kluster) Destroy() error {
	return k.cluster.Destroy()
}

func (k Kluster) deployResources() error {
	for _, path := range k.cfg.Resources {
		resolved, err := resolver.ResolveFile(path)
		if err != nil {
			return errors.Wrapf(err, "resolve references in resource: %s", path)
		}

		kube := kubectl.New(k.KubeconfigPath())
		result, err := kube.ExecStdinData(resolved, "apply", "-f", "-")
		if err != nil {
			return errors.Wrapf(err, "execute kubectl for resource: %s", path)
		}
		fmt.Printf("Deployed resource: %s\n", path)
		fmt.Println(string(result))
	}

	return nil
}
