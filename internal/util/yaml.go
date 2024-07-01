package util

import (
	"os"

	yamlv2 "gopkg.in/yaml.v2"
)

func GenerateYaml(path string, data interface{}) (err error) {
	out, marshalErr := yamlv2.Marshal(data)
	if marshalErr != nil {
		return marshalErr
	}
	file, crateErr := os.Create(path)
	if crateErr != nil {
		return crateErr
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	_, err = file.Write(out)
	return
}
