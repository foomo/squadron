package squadron

import (
	"github.com/miracl/conflate"
	yamlv2 "gopkg.in/yaml.v2"
)

func init() {
	yamlv2.FutureLineWrap()
	// define the unmarshallers for the given file extensions, blank extension is the global unmarshaller
	conflate.Unmarshallers = conflate.UnmarshallerMap{
		".yaml": conflate.UnmarshallerFuncs{conflate.YAMLUnmarshal},
		".yml":  conflate.UnmarshallerFuncs{conflate.YAMLUnmarshal},
	}
}
