package kluster

import (
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/jakub-gawlas/kluster/pkg/docker"
	"github.com/jakub-gawlas/kluster/pkg/golang"
	log "github.com/sirupsen/logrus"
)

type App struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

func (app App) BuildImage() error {
	cli := docker.New()
	imageName, err := app.ImageName()
	if err != nil {
		return err
	}
	return cli.BuildImage(app.Name+".Dockerfile", imageName)
}

func (app App) BuildBinary() error {
	return golang.Build(app.Path, app.BinaryPath())
}

func (app App) ImageName() (string, error) {
	tag, err := app.ImageTag()
	if err != nil {
		return "", err
	}
	return app.Name + ":" + tag, nil
}

func (app App) ImageTag() (string, error) {
	version, err := app.Version()
	if err != nil {
		return "", err
	}
	checksum, err := app.Checksum()
	if err != nil {
		return "", err
	}
	return version + "-" + checksum, nil
}

func (app App) Version() (string, error) {
	data, err := ioutil.ReadFile(app.Name + ".VERSION")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (app App) BinaryPath() string {
	return path.Join(path.Dir(app.Path), app.Name)
}

func (app App) Checksum() (string, error) {
	f, err := os.Open(app.BinaryPath())
	if err != nil {
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("close app binary: %s: %v", app.BinaryPath(), err)
		}
	}()

	h := fnv.New32a()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", h.Sum32()), nil
}
