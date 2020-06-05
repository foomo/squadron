{{- range $k, $v := .spec.selector }}
    {{- if eq $k "matchLabels" }}
        {{- range $k, $l := $v }}
        {{- printf "%s:%s\n" $k $l }}
        {{- end }}
    {{- else }}
        {{- printf "%s:%s\n" $k $v }}
    {{- end }}
{{- end }}
