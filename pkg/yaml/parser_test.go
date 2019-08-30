package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name      string
		given     []byte
		expected  YAML
		shouldErr bool
	}{
		{
			name:  "nil",
			given: nil,
			expected: YAML{
				Documents: []interface{}{},
			},
			shouldErr: false,
		},
		{
			name:  "empty",
			given: []byte{},
			expected: YAML{
				Documents: []interface{}{},
			},
			shouldErr: false,
		},
		{
			name: "single document",
			given: []byte(`
foo: bar
baz: 123
`),
			expected: YAML{
				Documents: []interface{}{
					map[interface{}]interface{}{
						"foo": "bar",
						"baz": 123,
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "many document",
			given: []byte(`
foo: bar
baz: 123
---
test: data
`),
			expected: YAML{
				Documents: []interface{}{
					map[interface{}]interface{}{
						"foo": "bar",
						"baz": 123,
					},
					map[interface{}]interface{}{
						"test": "data",
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "many document begins with separator",
			given: []byte(`
---
foo: bar
baz: 123
---
test: data
`),
			expected: YAML{
				Documents: []interface{}{
					map[interface{}]interface{}{
						"foo": "bar",
						"baz": 123,
					},
					map[interface{}]interface{}{
						"test": "data",
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "invalid format",
			given: []byte(`
	---
invalid: format
`),
			expected:  YAML{},
			shouldErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := Parse(c.given)
			assert.Equal(t, c.expected, actual)
			if c.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	cases := []struct {
		name      string
		given     YAML
		expected  []byte
		shouldErr bool
	}{
		{
			name:      "nil documents",
			given:     YAML{},
			expected:  []byte{},
			shouldErr: false,
		},
		{
			name: "empty documents",
			given: YAML{
				Documents: []interface{}{},
			},
			expected:  []byte{},
			shouldErr: false,
		},
		{
			name: "single document",
			given: YAML{
				Documents: []interface{}{
					map[interface{}]interface{}{
						"foo": "bar",
						"baz": 123,
					},
				},
			},
			expected: []byte(`baz: 123
foo: bar
`),
			shouldErr: false,
		},
		{
			name: "many documents",
			given: YAML{
				Documents: []interface{}{
					map[interface{}]interface{}{
						"foo": "bar",
						"baz": 123,
					},
					map[interface{}]interface{}{
						"test": "data",
					},
				},
			},
			expected: []byte(`baz: 123
foo: bar
---
test: data
`),
			shouldErr: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := c.given.Marshal()
			assert.Equal(t, string(c.expected), string(actual))
			if c.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
