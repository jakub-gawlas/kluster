package kluster

import (
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/jakub-gawlas/kluster/pkg/kubectl"

	"github.com/jakub-gawlas/kluster/pkg/cluster"
	"github.com/jakub-gawlas/kluster/pkg/helm"
	"github.com/jakub-gawlas/kluster/pkg/yaml"
)

type Kluster struct {
	cluster *cluster.Cluster
	cfg     *Config
	cfgPath string
}

func New(cfgPath string) (*Kluster, error) {
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	return &Kluster{
		cluster: cluster.New(cfg.Name),
		cfg:     cfg,
		cfgPath: cfgPath,
	}, nil
}

func (k Kluster) Name() string {
	return k.cfg.Name
}

func (k Kluster) Cluster() *cluster.Cluster {
	return k.cluster
}

func (k Kluster) Deploy() error {
	exists, err := k.cluster.Exists()
	if err != nil {
		return err
	}

	if !exists {
		if err := k.cluster.Create(k.cfgPath); err != nil {
			return err
		}
		kubeconfig, err := k.cluster.KubeConfigPath()
		if err != nil {
			return err
		}
		h := helm.New(kubeconfig)
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
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		resolved, err := yaml.ResolveRefs(content)
		if err != nil {
			return err
		}

		kubeconfig, err := k.cluster.KubeConfigPath()
		if err != nil {
			return err
		}

		kube := kubectl.New(kubeconfig)
		if err := kube.ExecStdinData(resolved, "apply", "-f", "-"); err != nil {
			return err
		}
	}
	return nil
}
