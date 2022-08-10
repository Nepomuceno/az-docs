# Azure compliance docs

Table of Contents
=================
* [Managment Groups](#managment-groups)
* [Assignments](#assignments)
{{- range $key, $value := .Assignments }}
  * [{{ $value.Properties.DisplayName }}](#{{ mdlink $value.Properties.DisplayName }})
{{- end }}

## Managment Groups

{{- range $key, $value := .Entities }}
* {{ $value.Name }}
{{- end }}

## Assignments
{{- range $key, $value := .Assignments }}

### {{ $value.Properties.DisplayName }}

Scope: **{{ $value.Properties.Scope }}**

{{ $value.Properties.Description }}

{{- end }}