# Azure compliance docs

Table of Contents
=================
* [Managment Groups](#managment-groups)
* [Assignments](#assignments)
{{- range $key, $value := .Assignments }}
  * [{{ $value.Properties.DisplayName }}](#{{ mdlink $value.Properties.DisplayName }})
{{- end }}
* [Initiatives](#Initiatives)
{{- range $key, $value := .UsedDefinitionSets }}
  * [{{ $value.Properties.DisplayName }}](#{{ mdlink $value.Properties.DisplayName }})
{{- end }}
* [Definitions](#Definitions)
{{- range $key, $value := .UsedDefinitions }}
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

> {{ $value.Properties.Description }}
{{- end }}

## Initiatives

{{- range $key, $value := .UsedDefinitionSets }}
### {{ $value.Properties.DisplayName }}

> {{ $value.Properties.Description }}

**Policies:**

{{- range $policyKey, $policy := $value.Definitions }}

- **{{ $policy.Properties.DisplayName }}:**

  _{{ $policy.Properties.Description }}_

{{- end }}
{{- end }}

## Definitions

{{- range $key, $value := .UsedDefinitions }}
### {{ $value.Properties.DisplayName }}

> {{ $value.Properties.Description }}

{{- end }}

