{{- range .Versions }}
## {{ .Tag.Name }}

{{ range .Commits -}}
{{ if not (and (hasPrefix .Header "Prepare v") (hasSuffix .Header " release")) -}}
- {{ .Header }}
{{ end -}}
{{ end -}}
{{ end -}}
