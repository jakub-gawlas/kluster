package kubectl

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/jakub-gawlas/kluster/pkg/tests"
	"github.com/stretchr/testify/assert"
)

func TestClient_Exec(t *testing.T) {
	cases := []struct {
		name        string
		givenArgs   []string
		testExec    func(t *testing.T, args []string)
		expectedErr error
	}{
		{
			name:      "no args",
			givenArgs: []string{},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl"}
				assert.Equal(t, expectedArgs, args)
			},
			expectedErr: nil,
		},
		{
			name:      "one arg",
			givenArgs: []string{"test"},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test"}
				assert.Equal(t, expectedArgs, args)
			},
			expectedErr: nil,
		},
		{
			name:      "many args",
			givenArgs: []string{"test", "foo", "-bar", "123"},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test", "foo", "-bar", "123"}
				assert.Equal(t, expectedArgs, args)
			},
			expectedErr: nil,
		},
		{
			name:      "exec error",
			givenArgs: []string{"test"},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test"}
				assert.Equal(t, expectedArgs, args)
				_, err := os.Stderr.Write([]byte("test-err"))
				assert.NoError(t, err)
				os.Exit(1)
			},
			expectedErr: fmt.Errorf("test-err"),
		},
	}
	const kubeconfigPath = "test-path"
	fake := tests.NewExecFaker()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			execInfo := fake.ExecInfo()
			if execInfo.IsFakeExecution {
				if execInfo.CaseName == c.name {
					assert.Equal(t, kubeconfigPath, os.Getenv("KUBECONFIG"))
					c.testExec(t, execInfo.Args)
				}
				return
			}

			execCommand = fake.FakeExec(c.name)
			defer func() { execCommand = exec.Command }()

			cli := New(kubeconfigPath)
			err := cli.Exec(c.givenArgs...)
			assert.Equal(t, c.expectedErr, err)
		})
	}
}

func TestClient_ExecStdinData(t *testing.T) {
	cases := []struct {
		name        string
		givenArgs   []string
		givenData   []byte
		testExec    func(t *testing.T, args []string)
		expectedErr error
	}{
		{
			name:      "many args no data",
			givenArgs: []string{"test", "foo", "-bar", "123"},
			givenData: nil,
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test", "foo", "-bar", "123"}
				assert.Equal(t, expectedArgs, args)

				expectedStdin := make([]byte, 0)
				actualStdin, err := ioutil.ReadAll(os.Stdin)
				assert.NoError(t, err)
				assert.Equal(t, expectedStdin, actualStdin)
			},
			expectedErr: nil,
		},
		{
			name:      "many args with data",
			givenArgs: []string{"test"},
			givenData: []byte("test-data"),
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test"}
				assert.Equal(t, expectedArgs, args)

				expectedStdin := []byte("test-data")
				actualStdin, err := ioutil.ReadAll(os.Stdin)
				assert.NoError(t, err)
				assert.Equal(t, expectedStdin, actualStdin)
			},
			expectedErr: nil,
		},
		{
			name:      "exec error with data",
			givenArgs: []string{"test"},
			givenData: []byte("test"),
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test"}
				assert.Equal(t, expectedArgs, args)

				expectedStdin := []byte("test")
				actualStdin, err := ioutil.ReadAll(os.Stdin)
				assert.NoError(t, err)
				assert.Equal(t, expectedStdin, actualStdin)

				_, err = os.Stderr.Write([]byte("test-err"))
				assert.NoError(t, err)
				os.Exit(1)
			},
			expectedErr: fmt.Errorf("test-err"),
		},
	}
	const kubeconfigPath = "test-path"
	fake := tests.NewExecFaker()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			execInfo := fake.ExecInfo()
			if execInfo.IsFakeExecution {
				if execInfo.CaseName == c.name {
					assert.Equal(t, kubeconfigPath, os.Getenv("KUBECONFIG"))
					c.testExec(t, execInfo.Args)
				}
				return
			}

			execCommand = fake.FakeExec(c.name)
			defer func() { execCommand = exec.Command }()

			cli := New(kubeconfigPath)
			err := cli.ExecStdinData(c.givenData, c.givenArgs...)
			assert.Equal(t, c.expectedErr, err)
		})
	}
}
