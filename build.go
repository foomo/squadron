package configurd

type Service struct {
	Name  string `yaml:"-"`
	Image string `yaml:"image"`
	Tag   string `yaml:"tag"`
	Build string `yaml:"build"`
	Chart string `yaml:"chart"`
}
