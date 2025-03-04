/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
)

//go:embed template/*.md
var embeddedResources embed.FS

type TemplateValues struct {
	Catalog           string
	Component         string
	AssessmentResults oscalTypes.AssessmentResults
}

// Get the catalog title as the template.md catalog info
func getCatalogTitle(catalog oscalTypes.Catalog) (string, error) {
	if catalog.Metadata.Title != "" {
		return catalog.Metadata.Title, nil
	} else {
		return "", fmt.Errorf("Error getting catalog title")
	}
}

// Get the component title as the template.md component info
// At that stage, it only supports the Components length is 1. When the
// observation links to the component in assessment plan, it will be improved.
func getComponentTitle(assessmentPlan oscalTypes.AssessmentPlan) (string, error) {
	if len(*assessmentPlan.LocalDefinitions.Components) == 1 {
		component := (*assessmentPlan.LocalDefinitions.Components)[0]
		return component.Title, nil
	} else {
		return "", fmt.Errorf("Error getting component title")
	}
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
		// Loop through the Props slice to find the assessment-rule-id
		for _, prop := range *ob.Props { // Dereference the pointer to access the slice
			if prop.Name == "assessment-rule-id" {
				return prop.Value // Return the value if the property is found
			}
		}
	}
	// Return empty string if not found or UUID doesn't match
	return ""
}

func CreateTemplateValues(catalog oscalTypes.Catalog, assessmentPlan oscalTypes.AssessmentPlan, assessmentResults oscalTypes.AssessmentResults) (*TemplateValues, error) {
	catalogTitle, err := getCatalogTitle(catalog)
        if err != nil {
                return nil, err
        }
	componentTitle, err := getComponentTitle(assessmentPlan)
        if err != nil {
                return nil, err
        }

	return &TemplateValues{
		Catalog:           catalogTitle,
		Component:         componentTitle,
		AssessmentResults: assessmentResults,
	}, nil
}

func (p *TemplateValues) GenerateAssessmentResultsMd(mdfilepath string) ([]byte, error) {
	// Read the template file
	templateData, err := embeddedResources.ReadFile("template/template.md")
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
