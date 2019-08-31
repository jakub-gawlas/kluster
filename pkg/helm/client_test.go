package helm

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/jakub-gawlas/kluster/pkg/tests"
	"github.com/stretchr/testify/assert"
)

func TestClient_Upgrade(t *testing.T) {
	cases := []struct {
		name        string
		givenName   string
		givenPath   string
		givenSets   map[string]string
		testExec    func(t *testing.T, args []string)
		expectedErr error
	}{
		{
			name:      "no sets",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: nil,
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"helm", "upgrade", "test", "helm/test"}
				assert.Equal(t, expectedArgs, args)
			},
			expectedErr: nil,
		},
		{
			name:      "single set",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: map[string]string{
				"foo": "test",
			},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"helm", "upgrade", "test", "helm/test", "--set", "foo=test"}
				assert.Equal(t, expectedArgs, args)
			},
			expectedErr: nil,
		},
		{
			name:      "many sets sorted ascending",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: map[string]string{
				"foo": "test",
				"bar": "123",
				"baz": "test test",
			},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"helm", "upgrade", "test", "helm/test", "--set", "bar=123,baz=test test,foo=test"}
				assert.Equal(t, expectedArgs, args)
			},
			expectedErr: nil,
		},
		{
			name:      "exec error",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: nil,
			testExec: func(t *testing.T, args []string) {
				_, err := os.Stderr.Write([]byte("some-error"))
				assert.NoError(t, err)
				os.Exit(1)
			},
			expectedErr: fmt.Errorf("some-error"),
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

			cli := New(nil, kubeconfigPath)
			err := cli.Upgrade(c.givenName, c.givenPath, c.givenSets)
			assert.Equal(t, c.expectedErr, err)
		})
	}
}

func TestClient_Install(t *testing.T) {
	cases := []struct {
		name        string
		givenName   string
		givenPath   string
		givenSets   map[string]string
		testExec    func(*testing.T, tests.ExecutionInfo)
		expectedErr error
	}{
		{
			name:      "no sets",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: nil,
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				expectedArgs := []string{"helm", "install", "--name", "test", "helm/test"}
				assert.Equal(t, expectedArgs, info.Args)
			},
			expectedErr: nil,
		},
		{
			name:      "single set",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: map[string]string{
				"foo": "test",
			},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				expectedArgs := []string{"helm", "install", "--name", "test", "helm/test", "--set", "foo=test"}
				assert.Equal(t, expectedArgs, info.Args)
			},
			expectedErr: nil,
		},
		{
			name:      "many sets sorted ascending",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: map[string]string{
				"foo": "test",
				"bar": "123",
				"baz": "test test",
			},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				expectedArgs := []string{"helm", "install", "--name", "test", "helm/test", "--set", "bar=123,baz=test test,foo=test"}
				assert.Equal(t, expectedArgs, info.Args)
			},
			expectedErr: nil,
		},
		{
			// FIX: case is not executed (probably os.Exit break)
			name:      "exec error",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: nil,
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				expectedArgs := []string{"helm", "install", "--name", "test", "helm/test"}
				assert.Equal(t, expectedArgs, info.Args)
				_, err := os.Stderr.Write([]byte("some-error"))
				assert.NoError(t, err)
				os.Exit(1)
			},
			expectedErr: fmt.Errorf("some-error"),
		},
		{
			name:      "retry if exec error",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: nil,
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				expectedArgs := []string{"helm", "install", "--name", "test", "helm/test"}
				assert.Equal(t, expectedArgs, info.Args)

				if info.Execution < 3 {
					os.Exit(1)
				}
			},
			expectedErr: nil,
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
					c.testExec(t, execInfo)
				}
				return
			}

			execCommand = fake.FakeExec(c.name)
			defer func() { execCommand = exec.Command }()

			cli := New(nil, kubeconfigPath)
			err := cli.Install(c.givenName, c.givenPath, c.givenSets, WithMaxRetries(3), WithInterval(time.Microsecond))
			assert.Equal(t, c.expectedErr, err)
		})
	}
}

func TestClient_Init(t *testing.T) {
	cases := []struct {
		name      string
		givenName string
		givenPath string
		givenSets map[string]string
		prepare   func(*mocked)
		testExec  func(t *testing.T, args []string)
		shouldErr bool
	}{
		{
			name:      "all pass",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: nil,
			prepare: func(m *mocked) {
				m.On("Exec", []string{"create", "clusterrolebinding", "add-on-cluster-admin", "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default"}).Return(nil)
			},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"helm", "init"}
				assert.Equal(t, expectedArgs, args)
			},
			shouldErr: false,
		},
		{
			name:      "kubectl exec error",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: nil,
			prepare: func(m *mocked) {
				m.On("Exec", []string{"create", "clusterrolebinding", "add-on-cluster-admin", "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default"}).Return(fmt.Errorf("test-err"))
			},
			testExec: func(t *testing.T, args []string) {
				expectedArgs := []string{"helm", "init"}
				assert.Equal(t, expectedArgs, args)
			},
			shouldErr: true,
		},
		{
			name:      "exec error",
			givenName: "test",
			givenPath: "helm/test",
			givenSets: nil,
			prepare:   func(m *mocked) {},
			testExec: func(t *testing.T, args []string) {
				os.Exit(1)
			},
			shouldErr: true,
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

			m := &mocked{}
			c.prepare(m)
			cli := New(m, kubeconfigPath)
			err := cli.Init()

			if c.shouldErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}

			m.AssertExpectations(t)
		})
	}
}

type mocked struct {
	mock.Mock
}

func (m *mocked) Exec(arg ...string) error {
	args := m.Called(arg)
	return args.Error(0)
}
