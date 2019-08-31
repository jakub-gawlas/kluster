package kluster

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jakub-gawlas/kluster/pkg/yaml/resolver"
	"github.com/pkg/errors"
)

type Resource struct {
	Name  string   `yaml:"name"`
	Paths []string `yaml:"paths"`
}

type KubectlStdinExecutor interface {
	ExecStdinData([]byte, ...string) ([]byte, error)
}

func (r Resource) Deploy(kube KubectlStdinExecutor) error {
	paths := make([]string, 0)
	for _, path := range r.Paths {
		filePaths, err := resolveFilesFromPath(path)
		if err != nil {
			return err
		}
		paths = append(paths, filePaths...)
	}

	uniquePaths := unique(paths)

	for _, path := range uniquePaths {
		resolved, err := resolver.ResolveFile(path)
		if err != nil {
			return errors.Wrapf(err, "resolve references in resource: %s", path)
		}

		result, err := kube.ExecStdinData(resolved, "apply", "-f", "-")
		if err != nil {
			return errors.Wrapf(err, "kubectl apply on file: %s", path)
		}
		fmt.Printf("Deployed resource: %s\n", path)
		fmt.Println(string(result))
	}
	return nil
}

func resolveFilesFromPath(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return []string{path}, nil
	}

	patterns := []string{
		filepath.Join(path, "*.yaml"),
		filepath.Join(path, "*.yml"),
		filepath.Join(path, "**/*.yaml"),
		filepath.Join(path, "**/*.yml"),
	}

	paths := make([]string, 0)
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		paths = append(paths, matches...)
	}

	return paths, nil
}

func unique(values []string) []string {
	m := map[string]struct{}{}
	for _, v := range values {
		if _, ok := m[v]; !ok {
			m[v] = struct{}{}
		}
	}
	vv := make([]string, 0, len(m))
	for v := range m {
		vv = append(vv, v)
	}
	return vv
}
