package yaml

import (
	"strings"

	"gopkg.in/yaml.v2"
)

type YAML struct {
	Documents []interface{}
}

func Parse(data []byte) (YAML, error) {
	result := YAML{
		Documents: []interface{}{},
	}
	str := string(data)
	docsRaw := strings.Split(str, "---")

	for _, docRaw := range docsRaw {
		var doc interface{}
		if err := yaml.Unmarshal([]byte(docRaw), &doc); err != nil {
			return YAML{}, err
		}
		if doc == nil {
			continue
		}
		result.Documents = append(result.Documents, doc)
	}

	return result, nil
}

func (y YAML) Marshal() ([]byte, error) {
	docs := make([]string, 0, len(y.Documents))
	for _, doc := range y.Documents {
		docRaw, err := yaml.Marshal(doc)
		if err != nil {
			return nil, err
		}
		docs = append(docs, string(docRaw))
	}
	str := strings.Join(docs, "---\n")
	return []byte(str), nil
}
