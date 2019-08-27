package yaml

import (
	"encoding/base64"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	refSecretKey = "$secret"
)

func ResolveRefs(data []byte) ([]byte, error) {
	var file interface{}
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, errors.Wrap(err, "unmarshal yaml")
	}
	if err := resolveRefs(file); err != nil {
		return nil, errors.Wrap(err, "resolve reference")
	}
	return yaml.Marshal(file)
}

func resolveRefs(value interface{}) error {
	m, ok := value.(map[interface{}]interface{})
	if ok {
		for k, v := range m {
			ref, ok := toReference(v)
			if ok {
				content, err := ioutil.ReadFile(ref.Value)
				if err != nil {
					return err
				}
				m[k] = base64.StdEncoding.EncodeToString(content)
			} else {
				if err := resolveRefs(v); err != nil {
					return err
				}
			}
		}
		return nil
	}

	slice, ok := value.([]interface{})
	if ok {
		for _, v := range slice {
			if err := resolveRefs(v); err != nil {
				return err
			}
		}
	}

	return nil
}

type reference struct {
	Type  string
	Value string
}

func toReference(value interface{}) (reference, bool) {
	m, ok := value.(map[interface{}]interface{})
	if !ok {
		return reference{}, false
	}

	secret, ok := m[refSecretKey]
	if !ok {
		return reference{}, false
	}

	v, ok := secret.(string)
	if !ok {
		return reference{}, false
	}

	return reference{
		Type:  "SECRET",
		Value: v,
	}, true
}
