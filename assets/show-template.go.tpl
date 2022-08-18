Entities:
{{- range $key, $value := .Entities }}
{{ $value.ID }}
{{- end }}

Assignments:
{{- range $key, $value := .Assignments }}
{{ $value.Properties.DisplayName }} | {{ $value.Properties.Scope }}
{{- end }}

Used Definitions:
{{- range $key, $value := .UsedDefinitions }}
{{ $value.Properties.DisplayName }}
Description:
{{ $value.Properties.Description }}
{{- end }}

Used Definition Sets:
{{- range $key, $value := .UsedDefinitionSets }}
{{ $value.Properties.DisplayName }}
Description:
{{ $value.Properties.Description }}
Policies:
{{- end }}