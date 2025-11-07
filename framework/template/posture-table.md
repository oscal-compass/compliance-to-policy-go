# Compliance Posture Summary

## Catalog: {{.Catalog}}

{{- range $component := .Components}}

### Component: {{$component.ComponentTitle}}

| Control ID | Status | Failed Rules | Missing Rules | Passed Rules |
|------------|--------|--------------|---------------|--------------|
{{- range $finding := $component.Findings}}
{{- $statusEmoji := "🟡" }}
{{- $statusText := "Missing Results" }}
{{- $failedRulesList := "" }}
{{- $missingRulesList := "" }}
{{- $passedRulesList := "" }}
{{- $hasAnyResults := false }}
{{- if and $finding.Results (gt (len $finding.Results) 0) }}
{{- $firstFailed := true }}
{{- $firstMissing := true }}
{{- $firstPassed := true }}
{{- range $ruleResult := $finding.Results}}
{{- $ruleFailed := false }}
{{- $rulePassed := false }}
{{- $hasSubjects := false }}
{{- if and $ruleResult.Subjects (gt (len $ruleResult.Subjects) 0) }}
{{- $hasSubjects = true }}
{{- $hasAnyResults = true }}
{{- range $subj := $ruleResult.Subjects}}
{{- $subjFailed := false }}
{{- $subjPassed := false }}
{{- $subjIsWaived := false }}
{{- range $prop := $subj.Props}}
{{- if and (eq $prop.Name "result") (eq $prop.Value "pass") }}
{{- $subjPassed = true }}
{{- else if and (eq $prop.Name "result") (ne $prop.Value "pass") }}
{{- $subjFailed = true }}
{{- end}}
{{- if and (eq $prop.Name "waived") (eq $prop.Value "true") }}
{{- $subjIsWaived = true }}
{{- end}}
{{- end}}
{{- if and $subjFailed (not $subjIsWaived) }}
{{- $ruleFailed = true }}
{{- end}}
{{- if and $subjPassed (not $subjIsWaived) }}
{{- $rulePassed = true }}
{{- end}}
{{- end}}
{{- if $ruleFailed }}
{{- if $firstFailed }}
{{- $failedRulesList = $ruleResult.RuleId }}
{{- $firstFailed = false }}
{{- else }}
{{- $failedRulesList = printf "%s, %s" $failedRulesList $ruleResult.RuleId }}
{{- end}}
{{- else if and $hasSubjects $rulePassed }}
{{- if $firstPassed }}
{{- $passedRulesList = $ruleResult.RuleId }}
{{- $firstPassed = false }}
{{- else }}
{{- $passedRulesList = printf "%s, %s" $passedRulesList $ruleResult.RuleId }}
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
{{- else if ne $passedRulesList "" }}
{{- $statusEmoji = "🟢" }}
{{- $statusText = "Passed" }}
{{- end}}
{{- else }}
{{- $statusEmoji = "🟡" 
{{- $statusText = "Missing Results" }}
{{- $missingRulesList = "All rules" }}
{{- end}}
| {{$finding.ControlID}} | {{$statusEmoji}} {{$statusText}} | {{if ne $failedRulesList ""}}{{$failedRulesList}}{{else}}-{{end}} | {{if ne $missingRulesList ""}}{{$missingRulesList}}{{else}}-{{end}} | {{if ne $passedRulesList ""}}{{$passedRulesList}}{{else}}-{{end}} |
{{- end}}
{{- end}}
