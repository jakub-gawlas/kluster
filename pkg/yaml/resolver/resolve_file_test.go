package resolver

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestResolveFile(t *testing.T) {
	expectedOutputRaw, err := ioutil.ReadFile("test/output.yaml")
	assert.NoError(t, err)

	actualOutputRaw, err := ResolveFile("test/input.yaml")
	assert.NoError(t, err)

	var expectedOutput, actualOutput interface{}
	err = yaml.Unmarshal(expectedOutputRaw, &expectedOutput)
	assert.NoError(t, err)
	err = yaml.Unmarshal(actualOutputRaw, &actualOutput)
	assert.NoError(t, err)

	assert.Equal(t, expectedOutput, actualOutput)
}
