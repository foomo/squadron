package squadron

import (
	"sort"

	"github.com/pkg/errors"
)

type Units map[string]*Unit

func (u Units) Keys() []string {
	if len(u) == 0 {
		return nil
	}
	ret := make([]string, 0, len(u))
	for name := range u {
		ret = append(ret, name)
	}
	sort.Strings(ret)
	return ret
}

func (u Units) Values() []*Unit {
	if len(u) == 0 {
		return nil
	}
	ret := make([]*Unit, len(u))
	for _, name := range u.Keys() {
		ret = append(ret, u[name])
	}
	return ret
}

func (u Units) Filter(names []string) (Units, error) {
	if len(u) == 0 {
		return nil, nil
	}
	if len(names) == 0 {
		return u, nil
	}
	ret := make(Units, len(u))
	for _, name := range names {
		if unit, ok := u[name]; !ok {
			return nil, errors.Errorf("unknown unit: %s", name)
		} else {
			ret[name] = unit
		}
	}
	return ret, nil
}

func (u Units) Iterate(i func(name string, unit *Unit) error) error {
	if len(u) == 0 {
		return nil
	}
	for _, name := range u.Keys() {
		if err := i(name, u[name]); err != nil {
			return err
		}
	}
	return nil
}
