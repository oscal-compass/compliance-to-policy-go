/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package actions

import (
	"context"
	"os"
	"testing"
	"time"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/models/components"
	"github.com/oscal-compass/oscal-sdk-go/transformers"
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
							Type:        "inventory-item",
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
	defaultLogger = hclog.NewNullLogger()
)

func TestReport(t *testing.T) {
	inputContext, plan := inputContextHelperPlan(t)

	planHref := "https://test-plan-href"
	ar, err := Report(context.TODO(), inputContext, planHref, plan, pvpResults)
	require.NoError(t, err)
	require.Equal(t, ar.ImportAp.Href, planHref)

	// verify length of Results attributes
	require.Len(t, ar.Results, 1)
	require.Len(t, *ar.Results[0].Observations, 2)
	require.Len(t, *ar.Results[0].Findings, 1)
	findings := *ar.Results[0].Findings
	relatedObs := *findings[0].RelatedObservations

	// require that the observation is properly linked to the finding
	require.Len(t, relatedObs, 1)
	observationUUID := relatedObs[0].ObservationUuid

	var found bool
	for _, obs := range *ar.Results[0].Observations {
		if obs.UUID == observationUUID {
			found = true
			break
		}
	}
	require.True(t, found)
	require.Len(t, *ar.Results[0].LocalDefinitions.InventoryItems, 1)
}

func TestToOscalObservation(t *testing.T) {
	inputContext := inputContextHelper(t)
	rulesStore := inputContext.Store()

	observationByCheck := pvpResults[0].ObservationsByCheck[0]
	ruleSet, err := rulesStore.GetByCheckID(context.TODO(), observationByCheck.CheckID)
	require.NoError(t, err)

	idMap := make(map[string]string)
	oscalObs, err := toOscalObservation(observationByCheck, ruleSet, &idMap)
	require.NoError(t, err)
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

func TestGenerateFindings(t *testing.T) {
	inputContext := inputContextHelper(t)
	rulesStore := inputContext.Store()

	observationByCheck := pvpResults[0].ObservationsByCheck[0]
	ruleSet, err := rulesStore.GetByCheckID(context.TODO(), observationByCheck.CheckID)
	require.NoError(t, err)

	idMap := make(map[string]string)
	oscalObservation, err := toOscalObservation(observationByCheck, ruleSet, &idMap)
	require.NoError(t, err)

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
		findings, err := generateFindings(c.initFindings, oscalObservation, []string{"CIS-2.1_smt"})
		require.NoError(t, err)
		c.assertFunc(t, findings)
	}
}

func TestGenerateInventoryItem(t *testing.T) {

	tests := []struct {
		name            string
		subject         oscalTypes.SubjectReference
		expectedInvItem oscalTypes.InventoryItem
	}{
		{
			name: "Valid/InventoryItemWithProps",
			subject: oscalTypes.SubjectReference{
				SubjectUuid: "10a7c4ed-cb2e-4932-b993-905513d4789d",
				Title:       "Test Subject",
				Type:        "inventory-item",
				Props: &[]oscalTypes.Property{
					{
						Name:  "fqdn",
						Value: "test.com",
					},
					{
						Name:  "ipv4-address",
						Value: "10.1.1.1",
					},
					{
						Name:  "do-not-include",
						Value: "invalid",
					},
				},
			},
			expectedInvItem: oscalTypes.InventoryItem{
				UUID:        "10a7c4ed-cb2e-4932-b993-905513d4789d",
				Description: "Test Subject",
				Props: &[]oscalTypes.Property{
					{
						Name:  "fqdn",
						Value: "test.com",
					},
					{
						Name:  "ipv4-address",
						Value: "10.1.1.1",
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			gotInvItem := generateInventoryItem(&c.subject)
			require.Equal(t, c.expectedInvItem, gotInvItem)
		})
	}
}

func TestGenerateResource(t *testing.T) {

	tests := []struct {
		name             string
		subject          oscalTypes.SubjectReference
		expectedResource oscalTypes.Resource
	}{
		{
			name: "Valid/Resource",
			subject: oscalTypes.SubjectReference{
				SubjectUuid: "10a7c4ed-cb2e-4932-b993-905513d4789d",
				Title:       "Test Subject",
				Type:        "inventory-item",
			},
			expectedResource: oscalTypes.Resource{
				UUID:  "10a7c4ed-cb2e-4932-b993-905513d4789d",
				Title: "Test Subject",
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			gotResource := generateResource(&c.subject)
			require.Equal(t, c.expectedResource, gotResource)
		})
	}
}

// inputContextHelperPlan created input context from a plan.
func inputContextHelperPlan(t *testing.T) (*InputContext, oscalTypes.AssessmentPlan) {
	testDataPath := pkg.PathFromPkgDirectory("./testdata/oscal/component-definition-test.json")
	file, err := os.Open(testDataPath)
	require.NoError(t, err)
	definition, err := models.NewComponentDefinition(file, validation.NoopValidator{})
	require.NoError(t, err)
	require.NotNil(t, definition)

	ap, err := transformers.ComponentDefinitionsToAssessmentPlan(context.TODO(), []oscalTypes.ComponentDefinition{*definition}, "cis")
	require.NoError(t, err)

	if ap.LocalDefinitions == nil || ap.LocalDefinitions.Activities == nil || ap.AssessmentAssets.Components == nil {
		t.Error("error converting component definition to assessment plan")
	}

	var allComponents []components.Component
	for _, component := range *ap.AssessmentAssets.Components {
		compAdapter := components.NewSystemComponentAdapter(component)
		allComponents = append(allComponents, compAdapter)
	}

	inputContext, err := NewContext(allComponents)
	require.NoError(t, err)
	return inputContext, *ap
}
