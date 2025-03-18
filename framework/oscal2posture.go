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

package framework

import (
	"bytes"
	"embed"
	"html/template"
	"os"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/extensions"

	tp "github.com/oscal-compass/compliance-to-policy-go/v2/framework/template"
)

//go:embed template/*.md
var embeddedResources embed.FS

type Oscal2Posture struct {
	logger            hclog.Logger
	assessmentResults *oscalTypes.AssessmentResults
	catalog           *oscalTypes.Catalog
	compDef           *oscalTypes.ComponentDefinition
	templateFile      *string
}

type TemplateValues struct {
	CatalogTitle     string
	Components       oscalTypes.ComponentDefinition
	AssessmentResult oscalTypes.AssessmentResults
}

func NewOscal2Posture(assessmentResults *oscalTypes.AssessmentResults, catalog *oscalTypes.Catalog, compDef *oscalTypes.ComponentDefinition, logger hclog.Logger) *Oscal2Posture {
	return &Oscal2Posture{
		assessmentResults: assessmentResults,
		catalog:           catalog,
		compDef:           compDef,
		logger:            logger,
	}
}

func (r *Oscal2Posture) findSubjects(ruleId string) []oscalTypes.SubjectReference {
	var subjects []oscalTypes.SubjectReference
	for _, ar := range r.assessmentResults.Results {
		for _, ob := range *ar.Observations {
			prop, found := extensions.GetTrestleProp("assessment-rule-id", *ob.Props)
			if found && prop.Value == ruleId {
				subjects = append(subjects, *ob.Subjects...)
			}
		}
	}
	return subjects
}

func (r *Oscal2Posture) toTemplateValue() tp.TemplateValue {
	templateValue := tp.TemplateValue{
		CatalogTitle: r.catalog.Metadata.Title,
		Components:   []tp.Component{},
	}
	for _, componentObject := range *r.compDef.Components {
		if componentObject.Type == "validation" {
			continue
		}
		component := tp.Component{
			ComponentTitle: componentObject.Title,
			ControlResults: []tp.ControlResult{},
		}
		for _, cio := range *componentObject.ControlImplementations {
			for _, co := range cio.ImplementedRequirements {
				controlResult := tp.ControlResult{
					ControlId:   co.ControlId,
					RuleResults: []tp.RuleResult{},
				}

				if co.Props != nil {
					ruleIdsProps := extensions.FindAllProps(*co.Props, extensions.WithName(extensions.RuleIdProp))
					for _, ruleId := range ruleIdsProps {
						subjects := []tp.Subject{}
						rawSubjects := r.findSubjects(ruleId.Value)
						for _, rawSubject := range rawSubjects {
							var result, reason string
							resultProp, resultFound := extensions.GetTrestleProp("result", *rawSubject.Props)
							reasonProp, reasonFound := extensions.GetTrestleProp("reason", *rawSubject.Props)

							if resultFound {
								result = resultProp.Value
								if reasonFound {
									reason = reasonProp.Value
								}
							} else {
								result = "Error"
								reason = "No results found."
							}
							subject := tp.Subject{
								Title:  rawSubject.Title,
								UUID:   rawSubject.SubjectUuid,
								Result: result,
								Reason: reason,
							}
							subjects = append(subjects, subject)
						}
						controlResult.RuleResults = append(controlResult.RuleResults, tp.RuleResult{
							RuleId:   ruleId.Value,
							Subjects: subjects,
						})
					}
				}
				component.ControlResults = append(component.ControlResults, controlResult)
			}
		}
		templateValue.Components = append(templateValue.Components, component)
	}
	return templateValue
}

func (r *Oscal2Posture) SetTemplateFile(templateFile string) {
	r.templateFile = &templateFile
}

func (r *Oscal2Posture) Generate() ([]byte, error) {
	var templateData []byte
	var err error
	if r.templateFile == nil {
		templateData, err = embeddedResources.ReadFile("template/template.md")
	} else {
		templateData, err = os.ReadFile(*r.templateFile)
	}
	if err != nil {
		return nil, err
	}

	funcmap := template.FuncMap{
		"newline_with_indent": func(text string, indent int) string {
			newText := strings.ReplaceAll(text, "\n", "\n"+strings.Repeat(" ", indent))
			return newText
		},
	}

	templateString := string(templateData)
	tmpl := template.New("report.md")
	tmpl.Funcs(funcmap)
	tmpl, err = tmpl.Parse(templateString)
	if err != nil {
		return nil, err
	}
	templateValue := r.toTemplateValue()
	buffer := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buffer, templateValue)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
