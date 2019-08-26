package kluster

import (
	"github.com/jakub-gawlas/kluster/pkg/cluster"
	"github.com/jakub-gawlas/kluster/pkg/helm"
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
