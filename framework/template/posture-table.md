# Compliance Posture Summary

## Catalog: {{.Catalog}}

{{- range $component := .Components}}

### Component: {{$component.ComponentTitle}}

| Control ID | Status | Failed Rules | Missing Rules |
|------------|--------|--------------|---------------|
{{- range $finding := $component.Findings}}
{{- $statusEmoji := "🟡" }}
{{- $statusText := "Missing Results" }}
{{- $failedRulesList := "" }}
{{- $missingRulesList := "" }}
{{- $hasAnyResults := false }}
{{- if and $finding.Results (gt (len $finding.Results) 0) }}
{{- $firstFailed := true }}
{{- $firstMissing := true }}
{{- range $ruleResult := $finding.Results}}
{{- $ruleFailed := false }}
{{- $hasSubjects := false }}
{{- if and $ruleResult.Subjects (gt (len $ruleResult.Subjects) 0) }}
{{- $hasSubjects = true }}
{{- $hasAnyResults = true }}
{{- range $subj := $ruleResult.Subjects}}
{{- $subjFailed := false }}
{{- $subjIsWaived := false }}
{{- range $prop := $subj.Props}}
{{- if and (eq $prop.Name "result") (eq $prop.Value "fail") }}
{{- $subjFailed = true }}
{{- end}}
{{- if and (eq $prop.Name "waived") (eq $prop.Value "true") }}
{{- $subjIsWaived = true }}
{{- end}}
{{- end}}
{{- if and $subjFailed (not $subjIsWaived) }}
{{- $ruleFailed = true }}
{{- end}}
{{- end}}
{{- if $ruleFailed }}
{{- if $firstFailed }}
{{- $failedRulesList = $ruleResult.RuleId }}
{{- $firstFailed = false }}
{{- else }}
{{- $failedRulesList = printf "%s, %s" $failedRulesList $ruleResult.RuleId }}
{{- end}}
{{- end}}
{{- end}}
{{- if not $hasSubjects }}
{{- if $firstMissing }}
{{- $missingRulesList = $ruleResult.RuleId }}
{{- $firstMissing = false }}
{{- else }}
{{- $missingRulesList = printf "%s, %s" $missingRulesList $ruleResult.RuleId }}
{{- end}}
{{- end}}
{{- end}}
{{- if ne $failedRulesList "" }}
{{- $statusEmoji = "🔴" }}
{{- $statusText = "Failed" }}
{{- else if ne $missingRulesList "" }}
{{- $statusEmoji = "🟡" }}
{{- $statusText = "Missing Results" }}
{{- else if $hasAnyResults }}
{{- $statusEmoji = "🟢" }}
{{- $statusText = "Passed" }}
{{- end}}
{{- else }}
{{- $statusEmoji = "🟡" }}
{{- $statusText = "Missing Results" }}
{{- $missingRulesList = "All rules" }}
{{- end}}
| {{$finding.ControlID}} | {{$statusEmoji}} {{$statusText}} | {{if ne $failedRulesList ""}}{{$failedRulesList}}{{else}}-{{end}} | {{if ne $missingRulesList ""}}{{$missingRulesList}}{{else}}-{{end}} |
{{- end}}
{{- end}}
