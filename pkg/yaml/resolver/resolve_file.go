package resolver

import (
	"io/ioutil"
	"path"
)

func ResolveFile(filePath string) ([]byte, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	basePath := path.Dir(filePath)
	resolver := New(basePath)
	return resolver.ResolveRefs(data)
}
