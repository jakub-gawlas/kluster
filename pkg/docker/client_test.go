package docker

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/jakub-gawlas/kluster/pkg/tests"
	"github.com/stretchr/testify/assert"
)

func TestClient_BuildImageWithChecksum(t *testing.T) {
	cases := []struct {
		name            string
		givenCli        Client
		givenDockerfile string
		givenImageName  string
		testExec        func(*testing.T, tests.ExecutionInfo)
		expectedImage   Image
		expectedErr     error
	}{
		{
			name: "no errors",
			givenCli: Client{
				tempImage:   "temp-image",
				builtImages: map[string]struct{}{},
			},
			givenDockerfile: "test.Dockerfile",
			givenImageName:  "test-svc",
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				switch info.Execution {
				case 0:
					assert.Equal(t, []string{"docker", "build", "--rm", "-f", "test.Dockerfile", "-t", "temp-image", "."}, info.Args)
				case 1:
					assert.Equal(t, []string{"docker", "inspect", "--format='{{.ID}}'", "temp-image"}, info.Args)
					_, err := os.Stdout.Write([]byte("sha256:some-hash"))
					assert.NoError(t, err)
				case 2:
					assert.Equal(t, []string{"docker", "tag", "temp-image"}, info.Args[:3])
					assert.Equal(t, "test-svc:some-hash", tests.FakeStdoutString(info.Args[3]))
				default:
					panic("too many executions")
				}
			},
			expectedImage: Image{
				Name:     "test-svc",
				Tag:      "some-hash",
				FullName: "test-svc:some-hash",
			},
			expectedErr: nil,
		},
		{
			name: "build exec error",
			givenCli: Client{
				tempImage:   "temp-image",
				builtImages: map[string]struct{}{},
			},
			givenDockerfile: "test.Dockerfile",
			givenImageName:  "test-svc",
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				switch info.Execution {
				case 0:
					_, err := fmt.Fprint(os.Stderr, "test-err")
					assert.NoError(t, err)
					os.Exit(1)
				default:
					panic("too many executions")
				}
			},
			expectedImage: Image{},
			expectedErr:   fmt.Errorf("build temp image: test-err"),
		},
		{
			name: "inspect exec error",
			givenCli: Client{
				tempImage:   "temp-image",
				builtImages: map[string]struct{}{},
			},
			givenDockerfile: "test.Dockerfile",
			givenImageName:  "test-svc",
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				switch info.Execution {
				case 0:
					assert.Equal(t, []string{"docker", "build", "--rm", "-f", "test.Dockerfile", "-t", "temp-image", "."}, info.Args)
				case 1:
					_, err := fmt.Fprint(os.Stderr, "test-err")
					assert.NoError(t, err)
					os.Exit(1)
				default:
					panic("too many executions")
				}
			},
			expectedImage: Image{},
			expectedErr:   fmt.Errorf("calculate image checksum: inspect temp image: test-err"),
		},
		{
			name: "tag exec error",
			givenCli: Client{
				tempImage:   "temp-image",
				builtImages: map[string]struct{}{},
			},
			givenDockerfile: "test.Dockerfile",
			givenImageName:  "test-svc",
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				switch info.Execution {
				case 0:
					assert.Equal(t, []string{"docker", "build", "--rm", "-f", "test.Dockerfile", "-t", "temp-image", "."}, info.Args)
				case 1:
					assert.Equal(t, []string{"docker", "inspect", "--format='{{.ID}}'", "temp-image"}, info.Args)
					_, err := os.Stdout.Write([]byte("sha256:some-hash"))
					assert.NoError(t, err)
				case 2:
					_, err := fmt.Fprint(os.Stderr, "test-err")
					assert.NoError(t, err)
					os.Exit(1)
				default:
					panic("too many executions")
				}
			},
			expectedImage: Image{},
			expectedErr:   fmt.Errorf("tag image: test-err"),
		},
		{
			name: "invalid checksum",
			givenCli: Client{
				tempImage:   "temp-image",
				builtImages: map[string]struct{}{},
			},
			givenDockerfile: "test.Dockerfile",
			givenImageName:  "test-svc",
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				switch info.Execution {
				case 0:
					assert.Equal(t, []string{"docker", "build", "--rm", "-f", "test.Dockerfile", "-t", "temp-image", "."}, info.Args)
				case 1:
					_, err := fmt.Fprint(os.Stderr, "invalid-checksum")
					assert.NoError(t, err)
					os.Exit(1)
				default:
					panic("too many executions")
				}
			},
			expectedImage: Image{},
			expectedErr:   fmt.Errorf("calculate image checksum: inspect temp image: invalid-checksum"),
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

			actualImage, err := c.givenCli.BuildImageWithChecksum(c.givenDockerfile, c.givenImageName)
			actualImage.Tag = tests.FakeStdoutString(actualImage.Tag)
			actualImage.FullName = tests.FakeStdoutString(actualImage.FullName)

			assert.Equal(t, c.expectedImage, actualImage)

			if c.expectedErr != nil {
				assert.EqualError(t, err, c.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_Cleanup(t *testing.T) {
	cases := []struct {
		name        string
		givenCli    Client
		testExec    func(*testing.T, tests.ExecutionInfo)
		expectedErr error
	}{
		{
			name: "one image",
			givenCli: Client{
				builtImages: map[string]struct{}{
					"image:1": struct{}{},
				},
			},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				switch info.Execution {
				case 0:
					assert.Equal(t, []string{"docker", "image", "rm", "image:1"}, info.Args)
				default:
					panic("too many executions")
				}
			},
			expectedErr: nil,
		},
		{
			name: "many images",
			givenCli: Client{
				builtImages: map[string]struct{}{
					"image:1": struct{}{},
					"image:2": struct{}{},
					"image:3": struct{}{},
				},
			},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				switch info.Execution {
				case 0:
					assert.Equal(t, []string{"docker", "image", "rm", "image:1"}, info.Args)
				case 1:
					assert.Equal(t, []string{"docker", "image", "rm", "image:2"}, info.Args)
				case 2:
					assert.Equal(t, []string{"docker", "image", "rm", "image:3"}, info.Args)
				default:
					panic("too many executions")
				}
			},
			expectedErr: nil,
		},
		{
			name: "exec error",
			givenCli: Client{
				builtImages: map[string]struct{}{
					"image:1": struct{}{},
				}},
			testExec: func(t *testing.T, info tests.ExecutionInfo) {
				_, err := os.Stderr.Write([]byte("test-err"))
				assert.NoError(t, err)
				os.Exit(1)
			},
			expectedErr: fmt.Errorf(`remove image "image:1": test-err`),
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

			err := c.givenCli.Cleanup()
			if c.expectedErr != nil {
				assert.EqualError(t, err, c.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
