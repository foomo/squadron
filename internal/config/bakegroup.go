package config

type BakeGroup struct {
	Name        string   `json:"-" yaml:"-" hcl:"name,label"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty" hcl:"description,optional"`
	Targets     []string `json:"targets" yaml:"targets" hcl:"targets"`
}
