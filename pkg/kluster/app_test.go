package kluster

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jakub-gawlas/kluster/pkg/tests"
)

func TestApp_Prepare(t *testing.T) {
	cases := []struct {
		name        string
		given       App
		testExec    func(*testing.T, tests.ExecutionInfo)
		expectedErr error
	}{
		{
			name: "nil before build",
			given: App{
				BeforeBuild: nil,
			},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				panic("should not execute")
			},
			expectedErr: nil,
		},
		{
			name: "empty before build",
			given: App{
				BeforeBuild: []string{},
			},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				panic("should not execute")
			},
			expectedErr: nil,
		},
		{
			name: "empty before build command",
			given: App{
				BeforeBuild: []string{""},
			},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				panic("should not execute")
			},
			expectedErr: fmt.Errorf("invalid format"),
		},
		{
			name: "one before build command",
			given: App{
				BeforeBuild: []string{
					"test --arg foo",
				},
			},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				expectedArgs := []string{"test", "--arg", "foo"}
				assert.Equal(t, expectedArgs, info.Args)
			},
			expectedErr: nil,
		},
		{
			name: "many before build commands",
			given: App{
				BeforeBuild: []string{
					"test --arg foo",
					"test --arg bar",
					"test --arg baz",
				},
			},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				var expectedArgs []string
				switch info.Execution {
				case 0:
					expectedArgs = []string{"test", "--arg", "foo"}
				case 1:
					expectedArgs = []string{"test", "--arg", "bar"}
				case 2:
					expectedArgs = []string{"test", "--arg", "baz"}
				default:
					panic("too many executions")
				}
				assert.Equal(t, expectedArgs, info.Args)
			},
			expectedErr: nil,
		},
		{
			name: "exec error",
			given: App{
				BeforeBuild: []string{
					"test --arg foo",
				},
			},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				_, err := os.Stderr.Write([]byte("test-err"))
				assert.NoError(t, err)
				os.Exit(1)
			},
			expectedErr: fmt.Errorf("test-err"),
		},
	}
	fake := tests.NewExecFaker()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			execInfo := fake.ExecInfo()
			if execInfo.IsFakeExecution {
				if execInfo.CaseName == c.name {
					c.testExec(t, execInfo)
				}
				return
			}

			execCommand = fake.FakeExec(c.name)
			defer func() { execCommand = exec.Command }()

			err := c.given.Prepare()
			assert.Equal(t, c.expectedErr, err)
		})
	}
}
