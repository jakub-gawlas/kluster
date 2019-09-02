package kluster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResource_Deploy(t *testing.T) {
	cmd := []string{"apply", "-f", "-"}
	cases := []struct {
		name        string
		given       Resource
		prepareMock func(*mocked)
		shouldErr   bool
	}{
		{
			name: "resolve dir recurrent",
			given: Resource{
				Name: "test",
				Paths: []string{
					"test/resources",
				},
			},
			prepareMock: func(m *mocked) {
				m.On("ExecInData", []byte("kind: Test1\n"), cmd).Return([]byte("test-1"), nil).Once()
				m.On("ExecInData", []byte("kind: Test2\n"), cmd).Return([]byte("test-2"), nil).Once()
				m.On("ExecInData", []byte("kind: Test3\n"), cmd).Return([]byte("test-3"), nil).Once()
			},
			shouldErr: false,
		},
		{
			name: "resolve files",
			given: Resource{
				Name: "test",
				Paths: []string{
					"test/resources/test_1.yaml",
					"test/resources/dir/test_2.yml",
				},
			},
			prepareMock: func(m *mocked) {
				m.On("ExecInData", []byte("kind: Test1\n"), cmd).Return([]byte("test-1"), nil).Once()
				m.On("ExecInData", []byte("kind: Test2\n"), cmd).Return([]byte("test-2"), nil).Once()
			},
			shouldErr: false,
		},
		{
			name: "resolve dir and files omit duplicates",
			given: Resource{
				Name: "test1",
				Paths: []string{
					"test/resources/dir",
					"test/resources/dir/test_2.yml",
				},
			},
			prepareMock: func(m *mocked) {
				m.On("ExecInData", []byte("kind: Test2\n"), cmd).Return([]byte("test-2"), nil).Once()
				m.On("ExecInData", []byte("kind: Test3\n"), cmd).Return([]byte("test-3"), nil).Once()
			},
			shouldErr: false,
		},
		{
			name: "non exists reference",
			given: Resource{
				Name: "test",
				Paths: []string{
					"test/resources/test_1.yaml",
					"test/resources/dir/test_2.yml",
					"test/resources/non_exists",
				},
			},
			prepareMock: func(m *mocked) {},
			shouldErr:   true,
		},
		{
			name: "exec error",
			given: Resource{
				Name: "test",
				Paths: []string{
					"test/resources/test_1.yaml",
				},
			},
			prepareMock: func(m *mocked) {
				m.On("ExecInData", []byte("kind: Test1\n"), cmd).Return(nil, fmt.Errorf("test-err")).Once()
			},
			shouldErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := &mocked{}
			c.prepareMock(m)
			err := c.given.Deploy(m)
			if c.shouldErr {
				assert.Error(t, err)
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

func (m *mocked) ExecInData(data []byte, arg ...string) ([]byte, error) {
	args := m.Called(data, arg)
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.([]byte), args.Error(1)
}
