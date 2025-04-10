/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/extensions"

	tp "github.com/oscal-compass/compliance-to-policy-go/v2/framework/template"
)

// ResultsTemplateValues defines values for a plan-based posture report.
type ResultsTemplateValues struct {
	Catalog           string
	Component         string
	AssessmentResults oscalTypes.AssessmentResults
}

func CreateResultsValues(
	catalog oscalTypes.Catalog,
	assessmentPlan oscalTypes.AssessmentPlan,
	assessmentResults oscalTypes.AssessmentResults,
) (*ResultsTemplateValues, error) {
	catalogTitle, err := getCatalogTitle(catalog)
	if err != nil {
		return nil, err
	}
	componentTitle, err := getComponentTitle(assessmentPlan)
	if err != nil {
		return nil, err
	}

	return &ResultsTemplateValues{
		Catalog:           catalogTitle,
		Component:         componentTitle,
		AssessmentResults: assessmentResults,
	}, nil
}

func (p *ResultsTemplateValues) GenerateAssessmentResultsMd(mdfilepath string) ([]byte, error) {
	// Read the template file
	templateData, err := embeddedResources.ReadFile("template/results.md")
	if err != nil {
		return nil, err
	}

	// Custom function to add indentation for newlines
	funcmap := template.FuncMap{
		"extractControlId": extractControlId,
		"extractRuleId":    extractRuleId,
		"newline_with_indent": func(text string, indent int) string {
			return strings.ReplaceAll(text, "\n", "\n"+strings.Repeat(" ", indent))
		},
	}

	// Convert templateData to string
	templateString := string(templateData)

	// Create a new template and parse the string
	tmpl, err := template.New(mdfilepath).Funcs(funcmap).Parse(templateString)
	if err != nil {
		return nil, err
	}

	// Generate the markdown content using the struct data
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, p)
	if err != nil {
		return nil, err
	}

	// Return the generated markdown content as a byte slice
	return buffer.Bytes(), nil
}

// ComponentTemplateValues defined values for a component-based posture report.
type ComponentTemplateValues struct {
	CatalogTitle string
	Components   []tp.Component
}

func CreateComponentValues(
	catalog *oscalTypes.Catalog,
	compDef *oscalTypes.ComponentDefinition,
	assessmentResults *oscalTypes.AssessmentResults,
	logger hclog.Logger,
) (ComponentTemplateValues, error) {

	if catalog == nil {
		return ComponentTemplateValues{}, errors.New("catalog cannot be nil")
	}

	catalogTitle, err := getCatalogTitle(*catalog)
	if err != nil {
		return ComponentTemplateValues{}, err
	}
	templateValue := ComponentTemplateValues{
		CatalogTitle: catalogTitle,
		Components:   []tp.Component{},
	}

	if assessmentResults == nil {
		return templateValue, errors.New("assessment results cannot be nil")
	}

	// Process assessment results
	subjectsByRule := findSubjects(*assessmentResults, logger)

	if compDef.Components == nil {
		return templateValue, nil
	}

	for _, componentObject := range *compDef.Components {
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
						rawSubjects, ok := subjectsByRule[ruleId.Value]
						if !ok {
							logger.Debug(fmt.Sprintf("no subjects found for rule %s", ruleId.Value))
						}
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
	return templateValue, nil
}

// Get the catalog title as the template.md catalog info
func getCatalogTitle(catalog oscalTypes.Catalog) (string, error) {
	if catalog.Metadata.Title != "" {
		return catalog.Metadata.Title, nil
	} else {
		return "", fmt.Errorf("error getting catalog title")
	}
}

// Get the component title as the template.md component info
// At that stage, it only supports the Components length is 1. When the
// observation links to the component in assessment plan, it will be improved.
func getComponentTitle(assessmentPlan oscalTypes.AssessmentPlan) (string, error) {
	if assessmentPlan.LocalDefinitions != nil {
		if len(*assessmentPlan.LocalDefinitions.Components) == 1 {
			component := (*assessmentPlan.LocalDefinitions.Components)[0]
			return component.Title, nil
		}
	}
	return "", fmt.Errorf("error getting component title")
}

// Get controlId info from finding.Target.TargetId
func extractControlId(targetId string) string {
	controlId, _ := strings.CutSuffix(targetId, "_smt")
	return controlId
}

// Get the controlId mapping Rules from result.Observations base on finding.RelatedObservations
func extractRuleId(ob oscalTypes.Observation, observationUuid string) string {
	// Check if the UUID matches
	if ob.UUID == observationUuid {
		// Check if Props is not nil
		if ob.Props != nil {
			// Loop through the Props slice to find the assessment-rule-id
			for _, prop := range *ob.Props { // Dereference the pointer to access the slice
				if prop.Name == "assessment-rule-id" {
					return prop.Value // Return the value if the property is found
				}
			}
		}
	}
	// Return empty string if not found or UUID doesn't match
	return ""
}

func findSubjects(assessmentResults oscalTypes.AssessmentResults, logger hclog.Logger) map[string][]oscalTypes.SubjectReference {
	subjectsByRule := make(map[string][]oscalTypes.SubjectReference)
	for _, ar := range assessmentResults.Results {
		if ar.Observations == nil {
			continue
		}
		for _, ob := range *ar.Observations {
			if ob.Props == nil || ob.Subjects == nil {
				logger.Debug(fmt.Sprintf("no subjects found for %s", ob.Title))
				continue
			}
			prop, found := extensions.GetTrestleProp("assessment-rule-id", *ob.Props)
			if found {
				subjectsByRule[prop.Value] = *ob.Subjects
			}
		}
	}
	return subjectsByRule
}
