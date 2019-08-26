package kluster

import (
	"fmt"

	"github.com/jakub-gawlas/kluster/pkg/cluster"
	"github.com/jakub-gawlas/kluster/pkg/helm"
)

type Chart struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	Apps []App  `yaml:"apps"`
}

func (chart Chart) Deploy(cluster *cluster.Cluster, installed bool) error {
	kubeconfig, err := cluster.KubeConfigPath()
	if err != nil {
		return err
	}

	if err := chart.prepareApps(cluster); err != nil {
		return err
	}

	sets, err := chart.setValues()
	if err != nil {
		return err
	}

	h := helm.New(kubeconfig)
	if installed {
		if err := h.Upgrade(chart.Name, chart.Path, sets); err != nil {
			return err
		}
	} else {
		if err := h.Install(chart.Name, chart.Path, sets); err != nil {
			return err
		}
	}

	return nil
}

func (chart Chart) prepareApps(cluster *cluster.Cluster) error {
	for _, app := range chart.Apps {
		if err := app.BuildBinary(); err != nil {
			return err
		}

		if err := app.BuildImage(); err != nil {
			return err
		}

		imageName, err := app.ImageName()
		if err != nil {
			return err
		}
		if err := cluster.LoadImage(imageName); err != nil {
			return err
		}
	}
	return nil
}

func (chart Chart) setValues() (map[string]string, error) {
	sets := map[string]string{}
	for _, app := range chart.Apps {
		tagKey := fmt.Sprintf("app.%s.image.tag", app.Name)
		tag, err := app.ImageTag()
		if err != nil {
			return nil, err
		}
		sets[tagKey] = tag

		pullPolicyKey := fmt.Sprintf("app.%s.image.pullPolicy", app.Name)
		sets[pullPolicyKey] = "IfNotPresent"
	}
	return sets, nil
}
