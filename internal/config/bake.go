package config

import (
	"github.com/genelet/horizon/dethcl"
)

type Bake struct {
	Groups  []*BakeGroup  `json:"group" yaml:"group" hcl:"group,block"`
	Targets []*BakeTarget `json:"target" yaml:"target" hcl:"target,block"`
}

func (c *Bake) HCL() ([]byte, error) {
	return dethcl.MarshalLevel(c, 0)
}
