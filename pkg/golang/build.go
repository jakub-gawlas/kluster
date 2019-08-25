package golang

import (
	"os"
	"os/exec"
)

const (
	goCmd = "go"
)

func Build(srcPath string, outPath string) error {
	cmd := exec.Command(goCmd, "build", "-a", "-installsuffix", "cgo", "-o", outPath, srcPath)
	cmd.Env = []string{"CGO_ENABLED=0", "GOOS=linux", "HOME=" + os.Getenv("HOME")}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
