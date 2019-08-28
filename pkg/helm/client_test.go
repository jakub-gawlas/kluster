package helm

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Upgrade(t *testing.T) {
	execCommand = createFakeExecCommand("TestMockHelmUpgrade")
	defer func() { execCommand = exec.Command }()

	cli := New("test")
	err := cli.Upgrade("test", "helm/test", nil)
	assert.NoError(t, err)
}

func TestMockHelmUpgrade(t *testing.T) {
	args, ok := fakeExec()
	if !ok {
		return
	}

	expectedArgs := []string{"helm", "upgrade", "test", "helm/test"}
	assert.Equal(t, expectedArgs, args)
}

func createFakeExecCommand(testName string) func(name string, arg ...string) *exec.Cmd {
	return func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=" + testName, "--", name}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)
		return cmd
	}
}

func fakeExec() ([]string, bool) {
	if len(os.Args) <= 3 || os.Args[2] != "--" {
		return nil, false
	}
	return os.Args[3:], true
}
