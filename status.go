package squadron

type Status struct {
	User     string `json:"user,omitempty"`
	Branch   string `json:"branch,omitempty"`
	Commit   string `json:"commit,omitempty"`
	Squadron string `json:"squadron,omitempty"`
}
