/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"context"
	"os"
	"testing"
	"time"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/settings"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/stretchr/testify/require"

	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

var (
	pvpResults = []policy.PVPResult{
		{
			ObservationsByCheck: []policy.ObservationByCheck{
				{
					Title:       "etcd_cert_file",
					Description: "Ensure that the --cert-file argument is set as appropriate",
					CheckID:     "etcd_cert_file",
					Methods:     []string{"test_method_1"},
					Subjects: []policy.Subject{
						{
							Title:       "test_subject_1",
							Result:      policy.ResultFail,
							Reason:      "not-satisfied",
							ResourceID:  "test_resource_1",
							EvaluatedOn: time.Now(),
						},
					},
					RelevantEvidences: []policy.Link{
						{
							Description: "test_related_evidence_1",
							Href:        "https://test-related-evidence-1",
						},
					},
				},
			},
			Links: []policy.Link{
				{
					Href:        "https:...",
					Description: "test_link_1",
				},
			},
		},
	}
	testDataPath = pkg.PathFromPkgDirectory("./testdata/oscal/component-definition-test.json")
)

func TestReporter_GenereateAssessmentResults(t *testing.T) {

	cfg := prepConfig(t)
	r, err := NewReporter(cfg)
	require.NoError(t, err)

	compDef := readCompDef(t)
	implementationSettings := prepImplementationSettings(t, compDef)

	testTitle := "test-title"
	planHref := "https://test-plan-href"
	opts := WithTitle(testTitle)

	ar, err := r.GenerateAssessmentResults(context.TODO(), planHref, &implementationSettings, pvpResults, opts)
	require.NoError(t, err)
	require.Equal(t, ar.Metadata.Title, testTitle)
	require.Equal(t, ar.ImportAp.Href, planHref)

	// verify lenght of Results attributes
	require.Len(t, ar.Results, 1)
	require.Len(t, *ar.Results[0].Observations, 1)
	require.Len(t, *ar.Results[0].Findings, 1)
	require.Len(t, ar.Results[0].ReviewedControls.ControlSelections, 1)

}

func TestReporter_FindControls(t *testing.T) {
	cfg := prepConfig(t)
	r, err := NewReporter(cfg)
	require.NoError(t, err)

	compDef := readCompDef(t)
	implementationSettings := prepImplementationSettings(t, compDef)
	foundControls := r.findControls(implementationSettings)
	includeControls := *foundControls.ControlSelections[0].IncludeControls

	require.Len(t, foundControls.ControlSelections, 1)
	require.Equal(t, includeControls[0], implementationSettings.AllControls()[0])
}

func TestReporter_ToOscalObservation(t *testing.T) {
	cfg := prepConfig(t)
	r, err := NewReporter(cfg)
	require.NoError(t, err)

	observationByCheck := pvpResults[0].ObservationsByCheck[0]
	ruleSet, err := r.rulesStore.GetByCheckID(context.TODO(), observationByCheck.CheckID)
	require.NoError(t, err)

	oscalObs := r.toOscalObservation(observationByCheck, ruleSet)
	require.Equal(t, oscalObs.Title, pvpResults[0].ObservationsByCheck[0].Title)
	require.Equal(t, oscalObs.Description, pvpResults[0].ObservationsByCheck[0].Description)

	oscalObsSubjects := *oscalObs.Subjects
	require.Len(t, oscalObsSubjects, 1)
	require.Equal(t, oscalObsSubjects[0].Title, pvpResults[0].ObservationsByCheck[0].Subjects[0].Title)
	require.Equal(t, oscalObsSubjects[0].Type, pvpResults[0].ObservationsByCheck[0].Subjects[0].Type)

	oscalObsRelEv := *oscalObs.RelevantEvidence
	require.Len(t, oscalObsRelEv, 1)
	require.Equal(t, oscalObsRelEv[0].Href, pvpResults[0].ObservationsByCheck[0].RelevantEvidences[0].Href)
	require.Equal(t, oscalObsRelEv[0].Description, pvpResults[0].ObservationsByCheck[0].RelevantEvidences[0].Description)

	oscalObsSubjectProps := *oscalObsSubjects[0].Props
	require.Len(t, oscalObsSubjectProps, 4)

	// iterate over OSCAL observation subject properties and verify match with PVP result
	for _, v := range oscalObsSubjectProps {
		if v.Name == "resource-id" {
			require.Equal(t, v.Value, pvpResults[0].ObservationsByCheck[0].Subjects[0].ResourceID)
		} else if v.Name == "result" {
			require.Equal(t, v.Value, pvpResults[0].ObservationsByCheck[0].Subjects[0].Result.String())
		} else if v.Name == "evaluated-on" {
			require.Equal(t, v.Value, pvpResults[0].ObservationsByCheck[0].Subjects[0].EvaluatedOn.String())
		} else if v.Name == "reason" {
			require.Equal(t, v.Value, pvpResults[0].ObservationsByCheck[0].Subjects[0].Reason)
		}
	}

}

func TestReporter_GenerateFindings(t *testing.T) {
	cfg := prepConfig(t)
	r, err := NewReporter(cfg)
	require.NoError(t, err)

	compDef := readCompDef(t)
	implementationSettings := prepImplementationSettings(t, compDef)

	observationByCheck := pvpResults[0].ObservationsByCheck[0]
	ruleSet, err := r.rulesStore.GetByCheckID(context.TODO(), observationByCheck.CheckID)
	require.NoError(t, err)

	oscalObservation := r.toOscalObservation(observationByCheck, ruleSet)

	tests := []struct {
		name         string
		initFindings []oscalTypes.Finding
		assertFunc   func(*testing.T, []oscalTypes.Finding)
	}{
		{
			name:         "Success/NewFinding",
			initFindings: []oscalTypes.Finding{},
			assertFunc: func(t *testing.T, findings []oscalTypes.Finding) {
				require.Len(t, findings, 1)
				relObs := *findings[0].RelatedObservations
				require.Len(t, relObs, 1)
				require.Equal(t, relObs[0].ObservationUuid, oscalObservation.UUID)
			},
		},
		{
			name: "Success/ExistingFindingWithMatchingControl",
			initFindings: []oscalTypes.Finding{
				{
					Target: oscalTypes.FindingTarget{
						TargetId: "CIS-2.1_smt",
						Type:     "statement-id",
						Status: oscalTypes.ObjectiveStatus{
							State: "not-satisfied",
						},
					},
					RelatedObservations: &[]oscalTypes.RelatedObservation{
						{
							ObservationUuid: "1234",
						},
					},
				},
			},
			assertFunc: func(t *testing.T, findings []oscalTypes.Finding) {
				require.Len(t, findings, 1)
				relObs := *findings[0].RelatedObservations
				require.Len(t, relObs, 2)
			},
		},
		{
			name: "Success/ExistingFindingNoMatchingControl",
			initFindings: []oscalTypes.Finding{
				{
					Target: oscalTypes.FindingTarget{
						TargetId: "X_smt",
						Type:     "statement-id",
						Status: oscalTypes.ObjectiveStatus{
							State: "not-satisfied",
						},
					},
					RelatedObservations: &[]oscalTypes.RelatedObservation{
						{
							ObservationUuid: "1234",
						},
					},
				},
			},
			assertFunc: func(t *testing.T, findings []oscalTypes.Finding) {
				require.Len(t, findings, 2)
				for _, f := range findings {
					relObs := *f.RelatedObservations
					require.Len(t, relObs, 1)
				}
			},
		},
	}

	for _, c := range tests {
		findings, err := r.generateFindings(c.initFindings, oscalObservation, ruleSet, implementationSettings)
		require.NoError(t, err)
		c.assertFunc(t, findings)
	}
}

// Load test component definition JSON
func readCompDef(t *testing.T) oscalTypes.ComponentDefinition {
	file, err := os.Open(testDataPath)
	require.NoError(t, err)

	definition, err := models.NewComponentDefinition(file, validation.NoopValidator{})
	require.NoError(t, err)
	require.NotNil(t, definition)

	return *definition
}

// Create implementation settings using test compdef
func prepImplementationSettings(t *testing.T, testComp oscalTypes.ComponentDefinition) settings.ImplementationSettings {

	var allImplementations []oscalTypes.ControlImplementationSet
	for _, component := range *testComp.Components {
		if component.ControlImplementations == nil {
			continue
		}
		allImplementations = append(allImplementations, *component.ControlImplementations...)
	}

	implementationSettings, err := settings.Framework("cis", allImplementations)
	require.NoError(t, err)

	return *implementationSettings

}
