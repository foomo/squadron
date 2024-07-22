package template

func defaultValue(value string, def any) any {
	if value == "" {
		return def
	}
	return value
}

func defaultIndexValue(v map[string]any, index string, def any) any {
	var ok bool
	if _, ok = v[index]; ok {
		return v[index]
	}
	return def
}
