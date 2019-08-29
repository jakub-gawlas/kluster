package kluster

import (
	"fmt"

	"github.com/jakub-gawlas/kluster/pkg/kubectl"

	"github.com/jakub-gawlas/kluster/pkg/cluster"
	"github.com/jakub-gawlas/kluster/pkg/docker"
	"github.com/jakub-gawlas/kluster/pkg/helm"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

	sets, err := chart.prepareApps(cluster)
	if err != nil {
		return errors.Wrap(err, "prepare apps")
	}

	kube := kubectl.New(kubeconfig)
	h := helm.New(kube, kubeconfig)
	if installed {
		fmt.Printf("\nUpgrading helm chart: %s ⬆", chart.Name)
		if err := h.Upgrade(chart.Name, chart.Path, sets); err != nil {
			if err := h.Install(chart.Name, chart.Path, sets); err != nil {
				return err
			}
		}
	} else {
		fmt.Printf("\nInstalling helm chart: %s 👷", chart.Name)
		if err := h.Install(chart.Name, chart.Path, sets); err != nil {
			return err
		}
	}

	return nil
}

func (chart Chart) prepareApps(cluster *cluster.Cluster) (sets map[string]string, err error) {
	cli := docker.New()
	defer func() {
		if err := cli.Cleanup(); err != nil {
			log.Errorf("cleanup docker client: %v", err)
		}
	}()

	sets = map[string]string{}
	for _, app := range chart.Apps {
		if err := app.Prepare(); err != nil {
			return nil, err
		}

		fmt.Printf("\nBuilding image for app: %s 🧩", app.Name)
		image, err := cli.BuildImageWithChecksum(app.Dockerfile, app.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "build image for app: %s", app.Name)
		}

		fmt.Printf("\nLoading image for app: %s to cluster ⤵", app.Name)
		if err := cluster.LoadImage(image.FullName); err != nil {
			return nil, errors.Wrapf(err, "load image: %s to cluster", image)
		}

		sets = extendSets(sets, app.Name, image.Name, image.Tag)
	}
	return sets, nil
}

func extendSets(sets map[string]string, appName, imageName, imageTag string) map[string]string {
	nameKey := fmt.Sprintf("app.%s.image.name", appName)
	sets[nameKey] = imageName

	tagKey := fmt.Sprintf("app.%s.image.tag", appName)
	sets[tagKey] = imageTag

	pullPolicyKey := fmt.Sprintf("app.%s.image.pullPolicy", appName)
	sets[pullPolicyKey] = "IfNotPresent"

	return sets
}
