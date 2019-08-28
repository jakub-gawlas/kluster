package resolver

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestResolver_ResolveRefs(t *testing.T) {
	input, err := ioutil.ReadFile("test/input.yaml")
	assert.NoError(t, err)

	expectedOutputRaw, err := ioutil.ReadFile("test/output.yaml")
	assert.NoError(t, err)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	baseDir := path.Dir(path.Join(wd, "test/input.yaml"))
	r := New(baseDir)

	actualOutputRaw, err := r.ResolveRefs(input)
	assert.NoError(t, err)

	var expectedOutput, actualOutput interface{}
	err = yaml.Unmarshal(expectedOutputRaw, &expectedOutput)
	assert.NoError(t, err)
	err = yaml.Unmarshal(actualOutputRaw, &actualOutput)
	assert.NoError(t, err)

	assert.Equal(t, expectedOutput, actualOutput)
}

func TestResolver_ResolveValue(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)

	cases := []struct {
		name          string
		givenBasePath string
		givenValue    interface{}
		expectedValue interface{}
		expectedOk    bool
		shouldErr     bool
	}{
		{
			name:          "key secret exists file",
			givenBasePath: wd,
			givenValue: map[interface{}]interface{}{
				"$secret": "test/ref/secret",
			},
			expectedValue: "U09NRV9DT05URU5U",
			expectedOk:    true,
			shouldErr:     false,
		},
		{
			name:          "key secret non exists file",
			givenBasePath: wd,
			givenValue: map[interface{}]interface{}{
				"$secret": "test/non-exists-file",
			},
			expectedValue: nil,
			expectedOk:    false,
			shouldErr:     true,
		},
		{
			name:          "key secret non exists file bad base path",
			givenBasePath: "bad/base/path",
			givenValue: map[interface{}]interface{}{
				"$secret": "test/ref/secret",
			},
			expectedValue: nil,
			expectedOk:    false,
			shouldErr:     true,
		},
		{
			name:          "invalid key",
			givenBasePath: wd,
			givenValue: map[interface{}]interface{}{
				"$invalidKey": "test/ref/secret",
			},
			expectedValue: nil,
			expectedOk:    false,
			shouldErr:     false,
		},
		{
			name:          "non ref value",
			givenBasePath: wd,
			givenValue:    "some-value",
			expectedValue: nil,
			expectedOk:    false,
			shouldErr:     false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := New(c.givenBasePath)
			actualValue, actualOk, err := r.ResolveValue(c.givenValue)
			assert.Equal(t, c.expectedValue, actualValue)
			assert.Equal(t, c.expectedOk, actualOk)
			if c.shouldErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
