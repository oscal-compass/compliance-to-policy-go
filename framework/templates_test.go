/*
Copyright 2025 The OSCAL Compass Authors
SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"

	tp "github.com/oscal-compass/compliance-to-policy-go/v2/framework/template"
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

func TestCreateTemplateValues(t *testing.T) {
	catalog := oscalTypes.Catalog{
		Metadata: oscalTypes.Metadata{
			Title: "Catalog Title",
		},
	}

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
			Catalog: "Catalog Title",
			Components: []tp.Component{
				{
					ComponentTitle: "Component Title",
					Findings: []tp.Findings{
						{
							ControlID: "control-1",
							Results: []tp.RuleResult{
								{
									RuleId: "rule-value",
									Subjects: []oscalTypes.SubjectReference{
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
				},
			},
		},
	}

	// Run test
	result, err := CreateResultsValues(test.catalog, test.assessmentPlan, test.assessmentResults, hclog.NewNullLogger())
	if err != nil {
		t.Errorf("Error creating ResultsTemplateValues: %v", err)
	}
	require.Equal(t, test.expected, result)
}
