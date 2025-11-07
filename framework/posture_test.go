/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"os"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	posturemd := NewPosture(&assessmentResults, &oscalTypes.Catalog{Metadata: oscalTypes.Metadata{
		Title: "Catalog Title",
	}}, &assessmentPlan, hclog.NewNullLogger())

	// Read the expected markdown file before running the test
	expectedmd, err := os.ReadFile("./testdata/assessment-results.md")
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", "../test/testdata/assessment-results.md", err)
	}

	// Run test
	assessmentResultsMd, err := posturemd.Generate("assessment-results.md")
	if err != nil {
		t.Errorf("Error generating markdown: %v", err)
	}

	// Compare the generated markdown with the expected markdown contents
	require.Equal(t, expectedmd, assessmentResultsMd)

	// Check No Findings
	assessmentResultsNoFindings := oscalTypes.AssessmentResults{
		Results: []oscalTypes.Result{
			{
				Findings: &[]oscalTypes.Finding{},
			},
		},
	}
	posturemd.assessmentResults = &assessmentResultsNoFindings

	expectedmd, err = os.ReadFile("./testdata/assessment-results-without-findings.md")
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", "../test/testdata/assessment-results-without-findings.md", err)
	}

	assessmentResultsMd, err = posturemd.Generate("assessment-result.md")
	if err != nil {
		t.Errorf("Error generating markdown: %v", err)
	}
	require.Equal(t, expectedmd, assessmentResultsMd)

	posturemd.assessmentPlan = &assessmentPlanMulti
	posturemd.assessmentResults = &assessmentResultsMulti
	expectedmd, err = os.ReadFile("./testdata/assessment-results-multi-comp.md")
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", "../test/testdata/assessment-results-multi-comp", err)
	}

	assessmentResultsMd, err = posturemd.Generate("assessment-result.md")
	if err != nil {
		t.Errorf("Error generating markdown: %v", err)
	}
	require.Equal(t, expectedmd, assessmentResultsMd)

	// Test table template
	posturemdTable := NewPosture(&assessmentResults, &oscalTypes.Catalog{Metadata: oscalTypes.Metadata{
		Title: "Catalog Title",
	}}, &assessmentPlan, hclog.NewNullLogger())
	posturemdTable.SetUseTableTemplate(true)

	expectedmd, err = os.ReadFile("./testdata/assessment-results-table.md")
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", "./testdata/assessment-results-table.md", err)
	}

	assessmentResultsMd, err = posturemdTable.Generate("assessment-results-table.md")
	require.NoError(t, err)
	require.Equal(t, string(expectedmd), string(assessmentResultsMd))
}

// Mock data for testing
var (
	assessmentPlan = oscalTypes.AssessmentPlan{
		LocalDefinitions: &oscalTypes.LocalDefinitions{
			Components: &[]oscalTypes.SystemComponent{
				{
					Title: "Component Title",
					Props: &[]oscalTypes.Property{
						{
							Name:  extensions.RuleIdProp,
							Value: "rule-value",
							Ns:    extensions.TrestleNameSpace,
						},
						{
							Name:  extensions.RuleIdProp,
							Value: "rule-needs-review",
							Ns:    extensions.TrestleNameSpace,
						},
					},
				},
			},
		},
	}
	assessmentResults = oscalTypes.AssessmentResults{
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
							{
								ObservationUuid: "observationuuid-review",
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
								Ns:    extensions.TrestleNameSpace,
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
					{
						UUID: "observationuuid-review",
						Props: &[]oscalTypes.Property{
							{
								Name:  "assessment-rule-id",
								Value: "rule-needs-review",
								Ns:    extensions.TrestleNameSpace,
							},
						},
						Subjects: &[]oscalTypes.SubjectReference{
							{
								SubjectUuid: "subject-5678",
								Title:       "configuration component",
								Props: &[]oscalTypes.Property{
									{
										Name:  "result",
										Value: "requires-remediation",
									},
									{
										Name:  "reason",
										Value: "Configuration partially compliant but requires remediation",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	// Multi component data
	assessmentPlanMulti = oscalTypes.AssessmentPlan{
		LocalDefinitions: &oscalTypes.LocalDefinitions{
			Components: &[]oscalTypes.SystemComponent{
				{
					Title: "Component Title",
					Props: &[]oscalTypes.Property{
						{
							Name:  extensions.RuleIdProp,
							Value: "rule-value",
							Ns:    extensions.TrestleNameSpace,
						},
					},
				},
				{
					Title: "Component Title 2",
					Props: &[]oscalTypes.Property{
						{
							Name:  extensions.RuleIdProp,
							Value: "rule-value-2",
							Ns:    extensions.TrestleNameSpace,
						},
						{
							Name:  extensions.RuleIdProp,
							Value: "rule-error-state",
							Ns:    extensions.TrestleNameSpace,
						},
					},
				},
			},
		},
	}
	assessmentResultsMulti = oscalTypes.AssessmentResults{
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
							{
								ObservationUuid: "observationuuid2",
							},
							{
								ObservationUuid: "observationuuid-error",
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
								Ns:    extensions.TrestleNameSpace,
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
					{
						UUID: "observationuuid2",
						Props: &[]oscalTypes.Property{
							{
								Name:  "assessment-rule-id",
								Value: "rule-value-2",
								Ns:    extensions.TrestleNameSpace,
							},
						},
						Subjects: &[]oscalTypes.SubjectReference{
							{
								SubjectUuid: "subject-1234",
								Title:       "my resource",
								Props: &[]oscalTypes.Property{
									{
										Name:  "result",
										Value: "pass",
									},
									{
										Name:  "reason",
										Value: "my reason",
									},
								},
							},
						},
					},
					{
						UUID: "observationuuid-error",
						Props: &[]oscalTypes.Property{
							{
								Name:  "assessment-rule-id",
								Value: "rule-error-state",
								Ns:    extensions.TrestleNameSpace,
							},
						},
						Subjects: &[]oscalTypes.SubjectReference{
							{
								SubjectUuid: "subject-9999",
								Title:       "network component",
								Props: &[]oscalTypes.Property{
									{
										Name:  "result",
										Value: "error",
									},
									{
										Name:  "reason",
										Value: "Network connectivity issue during evaluation",
									},
								},
							},
						},
					},
				},
			},
		},
	}
)
