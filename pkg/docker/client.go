package docker

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/pkg/errors"
)

type Client struct {
	tempImage   string
	builtImages map[string]struct{}
}

const (
	dockerCmd       = "docker"
	tempImagePrefix = "kluster-temp:"
)

func New() *Client {
	return &Client{
		tempImage:   tempImagePrefix + uuid.New().String(),
		builtImages: map[string]struct{}{},
	}
}

type Image struct {
	Name     string
	Tag      string
	FullName string
}

var execCommand = exec.Command

func (cli Client) BuildImageWithChecksum(dockerfilePath string, imageName string) (Image, error) {
	if err := cli.buildImage(dockerfilePath, cli.tempImage); err != nil {
		return Image{}, errors.Wrap(err, "build temp image")
	}

	checksum, err := cli.imageChecksum(cli.tempImage)
	if err != nil {
		return Image{}, errors.Wrap(err, "calculate image checksum")
	}

	image := imageName + ":" + checksum
	if err := cli.tagImage(cli.tempImage, image); err != nil {
		return Image{}, errors.Wrap(err, "tag image")
	}

	cli.builtImages[cli.tempImage] = struct{}{}
	cli.builtImages[image] = struct{}{}

	return Image{
		Name:     imageName,
		Tag:      checksum,
		FullName: image,
	}, nil
}

func (cli Client) Cleanup() error {
	images := make([]string, 0, len(cli.builtImages))
	for image := range cli.builtImages {
		images = append(images, image)
	}
	sort.Slice(images, func(i, j int) bool {
		return images[i] < images[j]
	})
	for _, image := range images {
		if err := cli.removeImage(image); err != nil {
			return errors.Wrapf(err, `remove image "%s"`, image)
		}
	}
	return nil
}

func (cli *Client) buildImage(dockerfilePath string, imageName string) error {
	var stderr bytes.Buffer
	cmd := execCommand(dockerCmd, "build", "--rm", "-f", dockerfilePath, "-t", imageName, ".")
	cmd.Env = []string{"DOCKER_BUILDKIT=1"}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}
	return nil
}

func (cli *Client) tagImage(sourceImage, targetImage string) error {
	var stderr bytes.Buffer
	cmd := execCommand(dockerCmd, "tag", sourceImage, targetImage)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}
	return nil
}

func (cli *Client) removeImage(image string) error {
	var stderr bytes.Buffer
	cmd := execCommand(dockerCmd, "image", "rm", image)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}
	return nil
}

func (cli *Client) imageChecksum(image string) (string, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd := execCommand(dockerCmd, "inspect", "--format='{{.ID}}'", image)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(fmt.Errorf(stderr.String()), "inspect temp image")
	}

	imageID := stdout.String()
	checksum, err := checksumFromImageID(imageID)
	if err != nil {
		return "", errors.Wrap(err, "parse checksum from imageID")
	}
	return checksum, nil
}

// imageID format: sha256:f30bc46dc114438d72e6ac19a82bd83c0dee86252e622ebc96f874d555a0e836
func checksumFromImageID(imageID string) (string, error) {
	checksum := strings.Split(imageID, ":")
	if len(checksum) < 2 {
		return "", fmt.Errorf("invalid format (%s)", imageID)
	}

	// TODO: clean response from os.Exec, it looks: 'sha256:f30bc46dc114438d72e6ac19a82bd83c0dee86252e622ebc96f874d555a0e836'\n
	r := strings.Replace(strings.Trim(checksum[1], "\n'"), "'", "", -1)

	return r, nil
}
