package util

import (
	"os"

	yamlv2 "gopkg.in/yaml.v2"
)

func WriteYAMLFile(path string, data any) error {
	out, err := yamlv2.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, 0600)
}
