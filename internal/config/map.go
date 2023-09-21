package config

import (
	"reflect"
	"slices"
	"sort"

	"github.com/pkg/errors"
)

type Map[T any] map[string]T

// Trim remove empty entries
func (m Map[T]) Trim() {
	for key, value := range m {
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if !val.IsValid() {
			delete(m, key)
			continue
		}
		if val.IsZero() {
			delete(m, key)
			continue
		}

		switch val.Kind() {
		case reflect.Map, reflect.Slice:
			if val.Len() == 0 {
				delete(m, key)
				continue
			}
		}
	}
}

// Keys returns the keys as a sorted list
func (m Map[T]) Keys() []string {
	if reflect.ValueOf(m).IsZero() {
		return nil
	}
	ret := make([]string, 0, len(m))
	for key := range m {
		ret = append(ret, key)
	}
	sort.Strings(ret)
	return ret
}

// Values returns all values sorted by keys
func (m Map[T]) Values() []T {
	if len(m) == 0 {
		return nil
	}
	keys := m.Keys()
	ret := make([]T, 0, len(keys))
	for i, key := range keys {
		ret[i] = m[key]
	}
	return ret
}

func (m Map[T]) Filter(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	validKeys := m.Keys()
	for _, key := range keys {
		if !slices.Contains(validKeys, key) {
			return errors.Errorf("key not found: `%s`", key)
		}
	}
	for key := range m {
		if !slices.Contains(keys, key) {
			delete(m, key)
		}
	}
	return nil
}

func (m Map[T]) Iterate(handler func(key string, value T) error) error {
	if len(m) == 0 {
		return nil
	}
	for _, key := range m.Keys() {
		if err := handler(key, m[key]); err != nil {
			return err
		}
	}
	return nil
}
