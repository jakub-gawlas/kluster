package resolver

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/jakub-gawlas/kluster/pkg/yaml"

	"github.com/pkg/errors"
)

type Resolver struct {
	basePath string
}

const (
	ReferenceSecret = "$secret"
)

func New(basePath string) *Resolver {
	return &Resolver{
		basePath: basePath,
	}
}

func (r *Resolver) ResolveRefs(data []byte) ([]byte, error) {
	file, err := yaml.Parse(data)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal yaml")
	}
	for _, doc := range file.Documents {
		if err := r.resolveRefs(doc); err != nil {
			return nil, errors.Wrap(err, "resolve reference")
		}
	}
	return file.Marshal()
}

func (r *Resolver) ResolveValue(value interface{}) (interface{}, bool, error) {
	ref, ok := toReference(value)
	if !ok {
		return nil, false, nil
	}

	switch ref.Type {
	case ReferenceSecret:
		filePath, ok := ref.Value.(string)
		if !ok {
			return nil, true, fmt.Errorf("expected string value")
		}
		data, err := r.readFile(filePath)
		if err != nil {
			return nil, true, err
		}
		v := base64.StdEncoding.EncodeToString(data)
		return v, true, nil
	}
	return nil, false, nil
}

func (r *Resolver) resolveRefs(value interface{}) error {
	m, ok := value.(map[interface{}]interface{})
	if ok {
		for k, v := range m {
			resolved, ok, err := r.ResolveValue(v)
			if err != nil {
				return err
			}
			if ok {
				m[k] = resolved
			} else {
				if err := r.resolveRefs(v); err != nil {
					return err
				}
			}
		}
		return nil
	}

	slice, ok := value.([]interface{})
	if ok {
		for _, v := range slice {
			if err := r.resolveRefs(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Resolver) readFile(filePath string) ([]byte, error) {
	fullPath := path.Join(r.basePath, filePath)
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, errors.Wrapf(err, "read file: %s", fullPath)
	}
	return data, nil
}

type reference struct {
	Type  string
	Value interface{}
}

func toReference(value interface{}) (reference, bool) {
	m, ok := value.(map[interface{}]interface{})
	if !ok {
		return reference{}, false
	}

	secret, ok := m[ReferenceSecret]
	if !ok {
		return reference{}, false
	}

	return reference{
		Type:  ReferenceSecret,
		Value: secret,
	}, true
}
