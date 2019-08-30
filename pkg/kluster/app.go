package kluster

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type App struct {
	Name        string   `yaml:"name"`
	Dockerfile  string   `yaml:"dockerfile"`
	BeforeBuild []string `yaml:"before_build"`
}

var execCommand = exec.Command

func (app App) Prepare() error {
	if len(app.BeforeBuild) == 0 {
		return nil
	}

	fmt.Printf("\nExecute before build scripts for app: %s ⚙", app.Name)
	for _, script := range app.BeforeBuild {
		fmt.Printf("\n↳ %s 🚀", script)
		out, err := run(script)
		if err != nil {
			return err
		}
		fmt.Printf("\n%s", out)
	}
	return nil
}

func run(script string) ([]byte, error) {
	cmds := strings.Split(script, " ")
	if len(cmds) == 0 || cmds[0] == "" {
		return nil, fmt.Errorf("invalid format")
	}

	var args []string
	if len(cmds) > 1 {
		args = cmds[1:]
	}

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd := execCommand(cmds[0], args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(stderr.String())
	}

	return stdout.Bytes(), nil
}
