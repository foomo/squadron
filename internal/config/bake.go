package config

type Bake struct {
	Groups  []*BakeGroup  `json:"group" yaml:"group" hcl:"group,block"`
	Targets []*BakeTarget `json:"target" yaml:"target" hcl:"target,block"`
}

type BakeGroup struct {
	Name        string   `json:"-" yaml:"-" hcl:"name,label"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty" hcl:"description,optional"`
	Targets     []string `json:"targets" yaml:"targets" hcl:"targets"`
}

type BakeTarget struct {
	Name        string `json:"-" yaml:"-" hcl:"name,label"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" hcl:"description,optional"`

	// Inherits is the only field that cannot be overridden with --set
	Inherits []string `json:"inherits,omitempty" yaml:"inherits,omitempty" hcl:"inherits,optional"`

	Annotations      []string          `json:"annotations,omitempty" yaml:"annotations,omitempty" hcl:"annotations,optional"`
	Attest           []map[string]any  `json:"attest,omitempty" yaml:"attest,omitempty" hcl:"attest,optional"`
	Context          string            `json:"context,omitempty" yaml:"context,omitempty" hcl:"context,optional"`
	Contexts         map[string]string `json:"contexts,omitempty" yaml:"contexts,omitempty" hcl:"contexts,optional"`
	Dockerfile       string            `json:"dockerfile,omitempty" yaml:"dockerfile,omitempty" hcl:"dockerfile,optional"`
	DockerfileInline string            `json:"dockerfile-inline,omitempty" yaml:"dockerfile-inline,omitempty" hcl:"dockerfile-inline,optional"`
	Args             map[string]string `json:"args,omitempty" yaml:"args,omitempty" hcl:"args,optional"`
	Labels           map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" hcl:"labels,optional"`
	Tags             []string          `json:"tags,omitempty" yaml:"tags,omitempty" hcl:"tags,optional"`
	CacheFrom        map[string]string `json:"cache-from,omitempty" yaml:"cache-from,omitempty" hcl:"cache-from,optional"`
	CacheTo          map[string]string `json:"cache-to,omitempty" yaml:"cache-to,omitempty" hcl:"cache-to,optional"`
	Target           string            `json:"target,omitempty" yaml:"target,omitempty" hcl:"target,optional"`
	Secrets          []map[string]any  `json:"secret,omitempty" yaml:"secret,omitempty" hcl:"secret,label,optional"`
	SSH              []map[string]any  `json:"ssh,omitempty" yaml:"ssh,omitempty" hcl:"ssh,optional"`
	Platforms        []string          `json:"platforms,omitempty" yaml:"platforms,omitempty" hcl:"platforms,optional"`
	Outputs          []map[string]any  `json:"output,omitempty" yaml:"output,omitempty" hcl:"output,optional"`
	Pull             bool              `json:"pull,omitempty" yaml:"pull,omitempty" hcl:"pull,optional"`
	NoCache          bool              `json:"no-cache,omitempty" yaml:"no-cache,omitempty" hcl:"no-cache,optional"`
	NetworkMode      string            `json:"network,omitempty" yaml:"network,omitempty" hcl:"network,optional"`
	NoCacheFilter    []string          `json:"no-cache-filter,omitempty" yaml:"no-cache-filter,omitempty" hcl:"no-cache-filter,optional"`
	ShmSize          string            `json:"shm-size,omitempty" yaml:"shm-size,omitempty" hcl:"shm-size,optional"`
	Ulimits          []string          `json:"ulimits,omitempty" yaml:"ulimits,omitempty" hcl:"ulimits,optional"`
	Call             string            `json:"call,omitempty" yaml:"call,omitempty" hcl:"call,optional"`
	Entitlements     []string          `json:"entitlements,omitempty" yaml:"entitlements,omitempty" hcl:"entitlements,optional"`
	ExtraHosts       map[string]string `json:"extra-hosts,omitempty" yaml:"extra-hosts,omitempty" hcl:"extra-hosts,optional"`
	// NOTE: use typed once it can be rendered as slice
	// Attest           buildflags.Attests      `json:"attest,omitempty" yaml:"attest,omitempty" hcl:"attest,optional"`
	// Secrets          buildflags.Secrets      `json:"secret,omitempty" yaml:"secret,omitempty" hcl:"secret,label,optional"`
	// SSH              buildflags.SSHKeys      `json:"ssh,omitempty" yaml:"ssh,omitempty" hcl:"ssh,optional"`
	// Outputs          buildflags.Exports      `json:"output,omitempty" yaml:"output,omitempty" hcl:"output,optional"`
	// CacheFrom        buildflags.CacheOptions `json:"cache-from,omitempty" yaml:"cache-from,omitempty" hcl:"cache-from,optional"`
	// CacheTo          buildflags.CacheOptions `json:"cache-to,omitempty" yaml:"cache-to,omitempty" hcl:"cache-to,optional"`
}
