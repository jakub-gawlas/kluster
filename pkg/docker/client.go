package docker

import (
	"os"
	"os/exec"
)

type Client struct{}

const (
	dockerCmd = "docker"
)

func New() *Client {
	return &Client{}
}

func (cli *Client) BuildImage(dockerfilePath string, imageName string) error {
	cmd := exec.Command(dockerCmd, "build", "--rm", "-f", dockerfilePath, "-t", imageName, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
