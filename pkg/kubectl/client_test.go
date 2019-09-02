package kubectl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/jakub-gawlas/kluster/pkg/tests"
	"github.com/stretchr/testify/assert"
)

func TestClient_ExecStdInOut(t *testing.T) {
	cases := []struct {
		name        string
		givenArgs   []string
		givenIn     []byte
		testExec    func(t *testing.T, args []string)
		expectedOut string
		expectedErr error
	}{
		{
			name:      "many args no std",
			givenArgs: []string{"test", "foo", "-bar", "123"},
			givenIn:   nil,
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test", "foo", "-bar", "123"}
				assert.Equal(t, expectedArgs, args)

				expectedStdin := make([]byte, 0)
				actualStdin, err := ioutil.ReadAll(os.Stdin)
				assert.NoError(t, err)
				assert.Equal(t, expectedStdin, actualStdin)
			},
			expectedOut: "",
			expectedErr: nil,
		},
		{
			name:      "only std in",
			givenArgs: []string{"test", "bar"},
			givenIn:   []byte("test-data"),
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test", "bar"}
				assert.Equal(t, expectedArgs, args)

				expectedStdin := []byte("test-data")
				actualStdin, err := ioutil.ReadAll(os.Stdin)
				assert.NoError(t, err)
				assert.Equal(t, expectedStdin, actualStdin)
			},
			expectedOut: "",
			expectedErr: nil,
		},
		{
			name:      "only std out",
			givenArgs: []string{"test", "baz"},
			givenIn:   nil,
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test", "baz"}
				assert.Equal(t, expectedArgs, args)

				_, err := os.Stdout.Write([]byte("test-data"))
				assert.NoError(t, err)
			},
			expectedOut: "test-data",
			expectedErr: nil,
		},
		{
			name:      "std in out",
			givenArgs: []string{"test", "bar", "baz"},
			givenIn:   []byte("test-data-in"),
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test", "bar", "baz"}
				assert.Equal(t, expectedArgs, args)

				expectedStdin := []byte("test-data-in")
				actualStdin, err := ioutil.ReadAll(os.Stdin)
				assert.NoError(t, err)
				assert.Equal(t, expectedStdin, actualStdin)

				_, err = os.Stdout.Write([]byte("test-data-out"))
				assert.NoError(t, err)
			},
			expectedOut: "test-data-out",
			expectedErr: nil,
		},
		{
			name:      "exec error",
			givenArgs: []string{"test"},
			givenIn:   nil,
			testExec: func(t *testing.T, args []string) {
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
			var stdout bytes.Buffer
			err := cli.ExecStdInOut(bytes.NewReader(c.givenIn), &stdout, c.givenArgs...)
			assert.Equal(t, c.expectedErr, err)
			assert.Equal(t, c.expectedOut, tests.FakeStdoutString(stdout.String()))
		})
	}
}

func TestClient_ExecInData(t *testing.T) {
	cases := []struct {
		name        string
		givenArgs   []string
		givenData   []byte
		testExec    func(t *testing.T, args []string)
		expectedRes []byte
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
			name:      "many args with data and stdout response",
			givenArgs: []string{"test"},
			givenData: []byte("test-data"),
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test"}
				assert.Equal(t, expectedArgs, args)

				expectedStdin := []byte("test-data")
				actualStdin, err := ioutil.ReadAll(os.Stdin)
				assert.NoError(t, err)
				assert.Equal(t, expectedStdin, actualStdin)

				_, err = os.Stdout.Write([]byte("test-data"))
				assert.NoError(t, err)
			},
			expectedRes: []byte("test-data"),
			expectedErr: nil,
		},
		{
			name:      "exec error with data",
			givenArgs: []string{"test"},
			givenData: []byte("test"),
			testExec: func(t *testing.T, args []string) {
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
			actualRes, err := cli.ExecInData(c.givenData, c.givenArgs...)
			assert.Equal(t, c.expectedRes, tests.FakeStdout(actualRes))
			assert.Equal(t, c.expectedErr, err)
		})
	}
}

func TestClient_Exec(t *testing.T) {
	cases := []struct {
		name        string
		givenArgs   []string
		testExec    func(t *testing.T, args []string)
		expectedRes []byte
		expectedErr error
	}{
		{
			name:      "no args",
			givenArgs: []string{},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl"}
				assert.Equal(t, expectedArgs, args)
				_, err := os.Stdout.Write([]byte("test-data"))
				assert.NoError(t, err)
			},
			expectedRes: []byte("test-data"),
			expectedErr: nil,
		},
		{
			name:      "one arg",
			givenArgs: []string{"test"},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test"}
				assert.Equal(t, expectedArgs, args)
				_, err := os.Stdout.Write([]byte("test-data-one"))
				assert.NoError(t, err)
			},
			expectedRes: []byte("test-data-one"),
			expectedErr: nil,
		},
		{
			name:      "many args",
			givenArgs: []string{"test", "foo", "-bar", "123"},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"kubectl", "test", "foo", "-bar", "123"}
				assert.Equal(t, expectedArgs, args)
				_, err := os.Stdout.Write([]byte("test-data-many"))
				assert.NoError(t, err)
			},
			expectedRes: []byte("test-data-many"),
			expectedErr: nil,
		},
		{
			name:      "exec error",
			givenArgs: []string{"test"},
			testExec: func(t *testing.T, args []string) {
				_, err := os.Stderr.Write([]byte("test-err"))
				assert.NoError(t, err)
				os.Exit(1)
			},
			expectedRes: nil,
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
			actualRes, err := cli.Exec(c.givenArgs...)
			assert.Equal(t, c.expectedRes, tests.FakeStdout(actualRes))
			assert.Equal(t, c.expectedErr, err)
		})
	}
}
