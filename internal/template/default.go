package template

func defaultIndexValue(v map[string]any, index string, def any) any {
	var ok bool
	if _, ok = v[index]; ok {
		return v[index]
	}
	return def
}
