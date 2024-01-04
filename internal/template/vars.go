package template

type Vars map[string]interface{}

func (tv *Vars) Add(name string, value interface{}) {
	(*tv)[name] = value
}
