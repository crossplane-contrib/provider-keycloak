{{- $link := .Get "link" -}}
{{- with .Page.GetPage $link -}}
  {{- with .OutputFormats.Get "markdown" -}}
    {{- $link = .Permalink -}}
  {{- end -}}
{{- end }}
- [{{ .Get "title" }}]({{ $link }}){{ with .Get "subtitle" }}: {{ . }}{{ end }}
