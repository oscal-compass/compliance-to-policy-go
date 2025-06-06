# Assessment Results Details

# Catalog
{{.Catalog}}
{{- range $component := .Components}}

## Component: {{$component.ComponentTitle}}
{{- if $component.Findings }}
{{- range $finding := $component.Findings}}

-------------------------------------------------------

#### Result of control: {{$finding.ControlID}}
{{ if $finding.Results }}
{{- range $ruleResult := $finding.Results}}
Rule ID: {{$ruleResult.RuleId}}
<details><summary>Details</summary>
{{- range $subj := $ruleResult.Subjects}}


  - Subject UUID: {{$subj.SubjectUuid}}
  - Title: {{$subj.Title}}
{{- range $prop := $subj.Props}}
{{- if eq $prop.Name "result"}}

    - Result: {{$prop.Value}}
{{- end}}
{{- if eq $prop.Name "reason"}}

    - Reason:
      ```
      {{ newline_with_indent $prop.Value 6}}
      ```
{{- end}}
{{- end}}
{{- end}}
</details>
{{- end}}
{{- end}}
{{- end}}
{{- else}}

No Findings.
{{- end}}
{{- end}}
