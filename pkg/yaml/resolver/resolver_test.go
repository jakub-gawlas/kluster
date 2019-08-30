package resolver

import (
	"os"
	"testing"

	"github.com/jakub-gawlas/kluster/pkg/yaml"

	"github.com/stretchr/testify/assert"
)

func TestResolver_ResolveRefs(t *testing.T) {
	cases := []struct {
		name      string
		given     []byte
		expected  []byte
		shouldErr bool
	}{
		{
			name:      "nil",
			given:     nil,
			expected:  []byte{},
			shouldErr: false,
		},
		{
			name:      "empty",
			given:     []byte{},
			expected:  []byte{},
			shouldErr: false,
		},
		{
			name: "single doc single ref",
			given: []byte(`foo: { $secret: ref/secret }
`),
			expected: []byte(`foo: U09NRV9DT05URU5U
`),
			shouldErr: false,
		},
		{
			name: "single doc many refs",
			given: []byte(`foo: { $secret: ref/secret }
bar: { $secret: ref/top_secret }
`),
			expected: []byte(`foo: U09NRV9DT05URU5U
bar: VE9QX1NFQ1JFVA==
`),
			shouldErr: false,
		},
		{
			name: "many docs many refs",
			given: []byte(`foo: { $secret: ref/secret }
bar: { $secret: ref/top_secret }
---
baz: { $secret: ref/secret }
`),
			expected: []byte(`foo: U09NRV9DT05URU5U
bar: VE9QX1NFQ1JFVA==
---
baz: U09NRV9DT05URU5U
`),
			shouldErr: false,
		},
		{
			name: "invalid ref",
			given: []byte(`foo: { $secret: ref/secret_non_exists }
`),
			expected:  nil,
			shouldErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := New("./test")
			actual, actualErr := r.ResolveRefs(c.given)

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
			expectedOk:    true,
			shouldErr:     true,
		},
		{
			name:          "key secret non exists file bad base path",
			givenBasePath: "bad/base/path",
			givenValue: map[interface{}]interface{}{
				"$secret": "test/ref/secret",
			},
			expectedValue: nil,
			expectedOk:    true,
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
