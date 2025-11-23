package config

import (
	"bytes"

	"github.com/genelet/horizon/dethcl"
)

type Bake struct {
	Groups  []*BakeGroup  `json:"group" yaml:"group" hcl:"group,block"`
	Targets []*BakeTarget `json:"target" yaml:"target" hcl:"target,block"`
}

func (c *Bake) HCL() ([]byte, error) {
	b, err := dethcl.MarshalLevel(c, 0)
	if err != nil {
		return nil, err
	}

	b = bytes.ReplaceAll(b, []byte("$${"), []byte("${"))

	lines := bytes.Split(b, []byte("\n"))
	for i, line := range lines {
		lines[i] = bytes.TrimPrefix(line, []byte("  "))
	}

	return bytes.Join(append(lines, []byte("")), []byte("\n")), nil
}
