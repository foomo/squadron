package template

type Vars map[string]any

func (tv *Vars) Add(name string, value any) {
	(*tv)[name] = value
}
