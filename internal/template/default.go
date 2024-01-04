package template

func defaultValue(value string, def interface{}) interface{} {
	if value == "" {
		return def
	}
	return value
}

func defaultIndexValue(v map[string]interface{}, index string, def interface{}) interface{} {
	var ok bool
	if _, ok = v[index]; ok {
		return v[index]
	}
	return def
}
