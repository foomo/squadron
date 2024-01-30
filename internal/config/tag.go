package config

type Tag string

func (t Tag) String() string {
	return string(t)
}
