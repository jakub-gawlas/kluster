package cluster

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/create"
	clusternodes "sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/container/docker"
	"sigs.k8s.io/kind/pkg/fs"
	"sigs.k8s.io/kind/pkg/util"
	"sigs.k8s.io/kind/pkg/util/concurrent"
)

type Cluster struct {
	name string
}

func New(name string) *Cluster {
	return &Cluster{
		name: name,
	}
}

func (c *Cluster) Exists() (bool, error) {
	return cluster.IsKnown(c.name)
}

func (c *Cluster) KubeConfigPath() (string, error) {
	exists, err := c.Exists()
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("cluster: %s not exists", c.name)
	}

	ctx := cluster.NewContext(c.name)
	return ctx.KubeConfigPath(), nil
}

func (c *Cluster) Create(configPath string) error {
	exists, err := c.Exists()
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("cluster: %s already exists", c.name)
	}

	ctx := cluster.NewContext(c.name)
	if err := ctx.Create(
		create.WithConfigFile(configPath),
	); err != nil {
		if utilErrors, ok := err.(util.Errors); ok {
			for _, problem := range utilErrors.Errors() {
				log.Error(problem)
			}
			return errors.New("aborting due to invalid configuration")
		}
		return errors.Wrap(err, "failed to create cluster")
	}
	return nil
}

func (c *Cluster) Destroy() error {
	exists, err := c.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("cluster: %s not exists", c.name)
	}

	ctx := cluster.NewContext(c.name)
	return ctx.Delete()
}

func (c *Cluster) LoadImage(imageName string) error {
	exists, err := c.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("cluster: %s not exists", c.name)
	}

	imageID, err := docker.ImageID(imageName)
	if err != nil {
		return fmt.Errorf("image: %qs not present locally", imageName)
	}

	ctx := cluster.NewContext(c.name)
	nodes, err := ctx.ListInternalNodes()
	if err != nil {
		return err
	}

	// pick only the nodes that don't have the image
	selectedNodes := []clusternodes.Node{}
	for _, node := range nodes {
		id, err := node.ImageID(imageName)
		if err != nil || id != imageID {
			selectedNodes = append(selectedNodes, node)
		}
	}

	if len(selectedNodes) == 0 {
		return nil
	}

	// Save the image into a tar
	dir, err := fs.TempDir("", "image-tar")
	if err != nil {
		return errors.Wrap(err, "failed to create tempdir")
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			log.Errorf("remove image temp dir: %s: %v", dir, err)
		}
	}()

	imageTarPath := filepath.Join(dir, "image.tar")
	err = docker.Save(imageName, imageTarPath)
	if err != nil {
		return err
	}

	// Load the image on the selected nodes
	fns := make([]func() error, 0)
	for _, selectedNode := range selectedNodes {
		selectedNode := selectedNode // capture loop variable
		fns = append(fns, func() error {
			return loadImage(imageTarPath, &selectedNode)
		})
	}

	return concurrent.UntilError(fns)
}

// loads an image tarball onto a node
func loadImage(imageTarName string, node *clusternodes.Node) error {
	f, err := os.Open(imageTarName)
	if err != nil {
		return errors.Wrap(err, "failed to open image")
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("close image tar: %s: %v", imageTarName, err)
		}
	}()
	return node.LoadImageArchive(f)
}
