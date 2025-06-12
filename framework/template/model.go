/*
Copyright 2023 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package template

import oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"

type RuleResult struct {
	// Rule ID
	RuleId string `json:"ruleId,omitempty" yaml:"ruleId,omitempty"`
	// Subjects
	Subjects []oscalTypes.SubjectReference `json:"subjects,omitempty" yaml:"subjects,omitempty"`
}

type Findings struct {
	ControlID string       `json:"controlId,omitempty" yaml:"controlId,omitempty"`
	Results   []RuleResult `json:"results,omitempty" yaml:"results,omitempty"`
}

type Component struct {
	// Component title in component-definition
	ComponentTitle string `json:"componentTitle,omitempty" yaml:"componentTitle,omitempty"`
	// Results per control
	Findings []Findings `json:"findings,omitempty" yaml:"findings,omitempty"`
}
