package tests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecFaker(t *testing.T) {
	cases := []struct {
		name           string
		givenName      string
		givenArgs      []string
		givenStdIn     []byte
		testExec       func(*testing.T, ExecutionInfo)
		expectedStdOut string
		expectedStdErr string
		shouldErr      bool
	}{
		{
			name:      "stdout",
			givenName: "cmd",
			givenArgs: []string{"foo", "bar"},
			testExec: func(t *testing.T, info ExecutionInfo) {
				expectedArgs := []string{"cmd", "foo", "bar"}
				assert.Equal(t, expectedArgs, info.Args)

				_, err := os.Stdout.Write([]byte("test"))
				assert.NoError(t, err)
			},
			expectedStdOut: "test",
			expectedStdErr: "",
			shouldErr:      false,
		},
		{
			name:      "stderr",
			givenName: "cmd",
			givenArgs: []string{"foo", "bar", "baz"},
			testExec: func(t *testing.T, info ExecutionInfo) {
				expectedArgs := []string{"cmd", "foo", "bar", "baz"}
				assert.Equal(t, expectedArgs, info.Args)

				_, err := os.Stderr.Write([]byte("test"))
				assert.NoError(t, err)
			},
			expectedStdOut: "",
			expectedStdErr: "test",
			shouldErr:      false,
		},
		{
			name:       "stdin",
			givenName:  "cmd",
			givenArgs:  []string{"foo", "bar"},
			givenStdIn: []byte("test"),
			testExec: func(t *testing.T, info ExecutionInfo) {
				expectedArgs := []string{"cmd", "foo", "bar"}
				assert.Equal(t, expectedArgs, info.Args)

				_, err := os.Stderr.Write([]byte("test"))
				assert.NoError(t, err)

				actualStdIn, err := ioutil.ReadAll(os.Stdin)
				assert.Equal(t, "test", string(actualStdIn))
			},
			expectedStdOut: "",
			expectedStdErr: "test",
			shouldErr:      false,
		},
		{
			name:      "exec error",
			givenName: "cmd",
			givenArgs: []string{"foo", "bar"},
			testExec: func(t *testing.T, info ExecutionInfo) {
				expectedArgs := []string{"cmd", "foo", "bar"}
				assert.Equal(t, expectedArgs, info.Args)

				os.Exit(1)
			},
			expectedStdOut: "",
			expectedStdErr: "",
			shouldErr:      true,
		},
	}
	faker := NewExecFaker()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			execInfo := faker.ExecInfo()
			if execInfo.IsFakeExecution {
				if execInfo.CaseName == c.name {
					c.testExec(t, execInfo)
				}
				return
			}

			exec := faker.FakeExec(c.name)

			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)
			cmd := exec(c.givenName, c.givenArgs...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			cmd.Stdin = bytes.NewReader(c.givenStdIn)
			err := cmd.Run()

			assert.Equal(t, c.expectedStdOut, string(FakeStdout(stdout.Bytes())))
			assert.Equal(t, c.expectedStdErr, stderr.String())

			if c.shouldErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecutionCount(t *testing.T) {
	faker := NewExecFaker()

	execInfo := faker.ExecInfo()
	fmt.Println(execInfo)
	if execInfo.IsFakeExecution {
		// TODO: write to stdout after get rid of result test info from the one
		_, err := os.Stderr.Write([]byte(strconv.Itoa(execInfo.Execution)))
		assert.NoError(t, err)
		return
	}

	exec := faker.FakeExec("test")
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			var stderr bytes.Buffer
			cmd := exec("test")
			cmd.Stderr = &stderr
			err := cmd.Run()

			assert.NoError(t, err)
			executionCount, err := strconv.Atoi(stderr.String())
			assert.NoError(t, err)
			assert.Equal(t, i, executionCount)
		})
	}
}

func TestFakeStdout(t *testing.T) {
	cases := []struct {
		name     string
		given    []byte
		expected []byte
	}{
		{
			name:     "nil data",
			given:    nil,
			expected: nil,
		},
		{
			name:     "empty data",
			given:    []byte{},
			expected: []byte{},
		},
		{
			name:     "no from fake exec",
			given:    []byte("test data"),
			expected: []byte("test data"),
		},
		{
			name:     "data from fake exec",
			given:    []byte("test dataPASS\n"),
			expected: []byte("test data"),
		},
		{
			name:     "empty from fake exec",
			given:    []byte("PASS\n"),
			expected: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := FakeStdout(c.given)
			assert.Equal(t, c.expected, actual)
		})
	}
}
