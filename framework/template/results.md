# Assessment Results Details

# Catalog
{{.Catalog}}

## Component: {{.Component}}
{{- range $index, $result := .AssessmentResults.Results}}
{{- if $result.Findings}}
{{- range $findingIndex, $finding := $result.Findings}}

-------------------------------------------------------

#### Result of control: {{extractControlId $finding.Target.TargetId}}
{{- range $reobsIndex, $reobs := $finding.RelatedObservations}}
{{- range $obsIndex, $obs := index $result.Observations}} 
{{- if extractRuleId $obs $reobs.ObservationUuid}}

Rule ID: {{extractRuleId $obs $reobs.ObservationUuid}}
<details><summary>Details</summary>
{{- range $subjsIndex, $subj := $obs.Subjects}}


  - Subject UUID: {{$subj.SubjectUuid}}
  - Title: {{$subj.Title}}
{{- range $propIndex, $prop := $subj.Props }}
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
{{- end}}
{{- else}}

No Findings.
{{- end}}
{{- end}}
