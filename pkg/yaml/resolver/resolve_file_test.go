package resolver

import (
	"testing"

	"github.com/jakub-gawlas/kluster/pkg/yaml"
	"github.com/stretchr/testify/assert"
)

func TestResolveFile(t *testing.T) {
	cases := []struct {
		name      string
		given     string
		expected  []byte
		shouldErr bool
	}{
		{
			name:  "exists file",
			given: "test/input.yaml",
			expected: []byte(`
foo: bar
nested:
  ref: U09NRV9DT05URU5U
  foo: 123
---
baz: VE9QX1NFQ1JFVA==
`),
			shouldErr: false,
		},
		{
			name:      "non exists file",
			given:     "test/input_non_exists.yaml",
			expected:  nil,
			shouldErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, actualErr := ResolveFile(c.given)

			expectedYaml, err := yaml.Parse(c.expected)
			assert.NoError(t, err)
			actualYaml, err := yaml.Parse(actual)
			assert.NoError(t, err)

			assert.Equal(t, expectedYaml, actualYaml)
			if c.shouldErr {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}
