# Assessment Results Details

## Catalog

{{.Catalog}}
{{- range $component := .Components}}

### Component: {{$component.ComponentTitle}}

{{- if $component.Findings }}
{{- range $finding := $component.Findings}}

-------------------------------------------------------

#### Result of control: {{$finding.ControlID}} ({{$component.ComponentTitle}})
{{- if $finding.Results }}
{{- $hasFailedRules := false }}
{{- $hasPassedRules := false }}
{{- $hasWaivedRules := false }}

{{- range $ruleResult := $finding.Results}}
{{- $hasFailure := false }}
{{- $isWaived := false }}
{{- range $subj := $ruleResult.Subjects}}
{{- range $prop := $subj.Props}}
{{- if and (eq $prop.Name "result") (eq $prop.Value "fail") }}
{{- $hasFailure = true }}
{{- end}}
{{- if and (eq $prop.Name "waived") (eq $prop.Value "true") }}
{{- $isWaived = true }}
{{- end}}
{{- end}}
{{- end}}
{{- if $isWaived}}
{{- $hasWaivedRules = true }}
{{- else if $hasFailure}}
{{- $hasFailedRules = true }}
{{- else}}
{{- $hasPassedRules = true }}
{{- end}}
{{- end}}

{{- if or $hasFailedRules $hasWaivedRules}}
<details open>
<summary> Failed Rules</summary>

{{- if $hasWaivedRules}}
<details open>
<summary> Waived Rules</summary>

{{- range $ruleResult := $finding.Results}}
{{- $hasFailure := false }}
{{- $isWaived := false }}
{{- range $subj := $ruleResult.Subjects}}
{{- range $prop := $subj.Props}}
{{- if and (eq $prop.Name "result") (eq $prop.Value "fail") }}
{{- $hasFailure = true }}
{{- end}}
{{- if and (eq $prop.Name "waived") (eq $prop.Value "true") }}
{{- $isWaived = true }}
{{- end}}
{{- end}}
{{- end}}
{{- if $isWaived}}

**Rule ID:** {{$ruleResult.RuleId}}
{{- if not $hasFailure}} **(Unexpectedly Passed)**{{- end}}

<details open>
<summary>Waived Rule Details</summary>

{{- range $subj := $ruleResult.Subjects}}

- **Subject UUID:** {{$subj.SubjectUuid}}
- **Title:** {{$subj.Title}}
{{- range $prop := $subj.Props}}
{{- if eq $prop.Name "result"}}

  - **Result: {{$prop.Value}}**{{if not $hasFailure}}  **(Expected to fail but passed)**{{end}}
{{- end}}

{{- if eq $prop.Name "waived"}}

- **Waived: {{$prop.Value}}**
{{- end}}

{{- if eq $prop.Name "reason"}}
    <details open>
    <summary>{{if $hasFailure}}Waiver Details{{else}}Reason for Unexpected Pass{{end}}</summary>

    ```text
    {{ newline_with_indent $prop.Value 4}}
    ```

    </details>
{{- end}}
{{- end}}
{{- end}}
</details>
{{- end}}
{{- end}}
</details>
{{- end}}

{{- if $hasFailedRules}}
<details open>
<summary> Failed Rules</summary>

{{- range $ruleResult := $finding.Results}}
{{- $hasFailure := false }}
{{- $isWaived := false }}
{{- range $subj := $ruleResult.Subjects}}
{{- range $prop := $subj.Props}}
{{- if and (eq $prop.Name "result") (eq $prop.Value "fail") }}
{{- $hasFailure = true }}
{{- end}}
{{- if and (eq $prop.Name "waived") (eq $prop.Value "true") }}
{{- $isWaived = true }}
{{- end}}
{{- end}}
{{- end}}
{{- if and $hasFailure (not $isWaived)}}

**Rule ID:** {{$ruleResult.RuleId}}

<details open>
<summary>Failed Rule Details</summary>

{{- range $subj := $ruleResult.Subjects}}

- **Subject UUID:** {{$subj.SubjectUuid}}
- **Title:** {{$subj.Title}}
{{- range $prop := $subj.Props}}
{{- if eq $prop.Name "result"}}

  - **Result: {{$prop.Value}}**
{{- end}}

{{- if eq $prop.Name "reason"}}
    <details open>
    <summary>Failure Reason</summary>

    ```text
    {{ newline_with_indent $prop.Value 4}}
    ```

    </details>
{{- end}}
{{- end}}
{{- end}}
</details>
{{- end}}
{{- end}}
</details>
{{- end}}

</details>
{{- end}}

{{- if $hasPassedRules}}
<details>
<summary> Passed Rules</summary>

{{- range $ruleResult := $finding.Results}}
{{- $hasFailure := false }}
{{- $isWaived := false }}
{{- range $subj := $ruleResult.Subjects}}
{{- range $prop := $subj.Props}}
{{- if and (eq $prop.Name "result") (eq $prop.Value "fail") }}
{{- $hasFailure = true }}
{{- end}}
{{- if and (eq $prop.Name "waived") (eq $prop.Value "true") }}
{{- $isWaived = true }}
{{- end}}
{{- end}}
{{- end}}
{{- if and (not $hasFailure) (not $isWaived)}}

**Rule ID:** {{$ruleResult.RuleId}}

<details>
<summary>Passed Rule Details</summary>

{{- range $subj := $ruleResult.Subjects}}

- **Subject UUID:** {{$subj.SubjectUuid}}
- **Title:** {{$subj.Title}}
{{- range $prop := $subj.Props}}
{{- if eq $prop.Name "result"}}

  - **Result: {{$prop.Value}}**
{{- end}}

{{- if eq $prop.Name "reason"}}
    <details>
    <summary>Details</summary>

    ```text
    {{ newline_with_indent $prop.Value 4}}
    ```

    </details>
{{- end}}
{{- end}}
{{- end}}
</details>
{{- end}}
{{- end}}
</details>
{{- end}}
{{- end}}
{{- else}}

No Findings.
{{- end}}
{{- end}}
{{- end}}
