/*
Copyright 2025 The OSCAL Compass Authors
SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"bytes"
	"os"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/stretchr/testify/require"
)

func TestGetCatalogTitle(t *testing.T) {
	tests := []struct {
		catalog  oscalTypes.Catalog
		expected string
		hasError bool
	}{
		{
			catalog: oscalTypes.Catalog{
				Metadata: oscalTypes.Metadata{
					Title: "Catalog Title",
				},
			},
			expected: "Catalog Title",
			hasError: false,
		},
		{
			catalog: oscalTypes.Catalog{
				Metadata: oscalTypes.Metadata{
					Title: "",
				},
			},
			expected: "",
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result, err := getCatalogTitle(test.catalog)

			if test.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, result)
			}
		})
	}
}

func TestGetComponentTitle(t *testing.T) {
	tests := []struct {
		assessmentPlan oscalTypes.AssessmentPlan
		expected       string
		hasError       bool
	}{
		{
			assessmentPlan: oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Components: &[]oscalTypes.SystemComponent{
						{
							Title: "Component Title",
						},
					},
				},
			},
			expected: "Component Title",
			hasError: false,
		},
		{
			assessmentPlan: oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Components: &[]oscalTypes.SystemComponent{},
				},
			},
			expected: "",
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result, err := getComponentTitle(test.assessmentPlan)
			if test.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, result)
			}
		})
	}
}

func TestExtractControlId(t *testing.T) {
	test := struct {
		targetId string
		expected string
	}{
		targetId: "control-1_smt",
		expected: "control-1",
	}
	result := extractControlId(test.targetId)
	require.Equal(t, test.expected, result)
}

func TestExtractRuleId(t *testing.T) {
	// Setup mock data
	ob := oscalTypes.Observation{
		UUID: "1234-uuid",
		Props: &[]oscalTypes.Property{
			{
				Name:  "assessment-rule-id",
				Value: "rule-1",
			},
		},
	}
	tests := []struct {
		observation     oscalTypes.Observation
		observationUuid string
		expected        string
	}{
		{
			observation:     ob,
			observationUuid: "1234-uuid",
			expected:        "rule-1",
		},
		{
			observation:     ob,
			observationUuid: "wrong-uuid",
			expected:        "",
		},
	}

	for _, test := range tests {
		t.Run(test.observationUuid, func(t *testing.T) {
			result := extractRuleId(test.observation, test.observationUuid)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestCreateTemplateValues(t *testing.T) {
	catalog := oscalTypes.Catalog{
		Metadata: oscalTypes.Metadata{
			Title: "Catalog Title",
		},
	}
	assessmentPlan := oscalTypes.AssessmentPlan{
		LocalDefinitions: &oscalTypes.LocalDefinitions{
			Components: &[]oscalTypes.SystemComponent{
				{Title: "Component Title"},
			},
		},
	}
	assessmentResults := oscalTypes.AssessmentResults{}

	test := struct {
		catalog           oscalTypes.Catalog
		assessmentPlan    oscalTypes.AssessmentPlan
		assessmentResults oscalTypes.AssessmentResults
		expected          *ResultsTemplateValues
	}{
		catalog:           catalog,
		assessmentPlan:    assessmentPlan,
		assessmentResults: assessmentResults,
		expected: &ResultsTemplateValues{
			Catalog:           "Catalog Title",
			Component:         "Component Title",
			AssessmentResults: assessmentResults,
		},
	}

	// Run test
	result, err := CreateResultsValues(test.catalog, test.assessmentPlan, test.assessmentResults)
	if err != nil {
		t.Errorf("Error creating ResultsTemplateValues: %v", err)
	}
	require.Equal(t, test.expected, result)
}

func TestGenerateAssessmentResultsMd(t *testing.T) {
	// Mock data for testing
	assessmentResults := oscalTypes.AssessmentResults{
		Results: []oscalTypes.Result{
			{
				Findings: &[]oscalTypes.Finding{
					{
						Target: oscalTypes.FindingTarget{
							TargetId: "control-1_smt",
						},
						RelatedObservations: &[]oscalTypes.RelatedObservation{
							{
								ObservationUuid: "observationuuid",
							},
						},
					},
				},
				Observations: &[]oscalTypes.Observation{
					{
						UUID: "observationuuid",
						Props: &[]oscalTypes.Property{
							{
								Name:  "assessment-rule-id",
								Value: "rule-value",
							},
						},
						Subjects: &[]oscalTypes.SubjectReference{
							{
								SubjectUuid: "subject-1234",
								Title:       "my component",
								Props: &[]oscalTypes.Property{
									{
										Name:  "result",
										Value: "fail",
									},
									{
										Name:  "reason",
										Value: "my reason",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	templateValues := &ResultsTemplateValues{
		Catalog:           "Catalog Title",
		Component:         "Component Title",
		AssessmentResults: assessmentResults,
	}

	// Read the expected markdown file before running the test
	expectedmd, err := os.ReadFile("./testdata/assessment-results.md")
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", "../test/testdata/assessment-results.md", err)
	}

	// Run test
	assessmentResultsMd, err := templateValues.GenerateAssessmentResultsMd("assessment-results.md")
	if err != nil {
		t.Errorf("Error generating markdown: %v", err)
	}

	// Compare the generated markdown with the expected markdown contents
	if !bytes.Equal(expectedmd, assessmentResultsMd) {
		t.Errorf("The generated markdown file is not equal to expected markdown")
	}
}
