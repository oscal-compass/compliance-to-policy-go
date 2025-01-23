/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
	"github.com/oscal-compass/oscal-sdk-go/generators"
	"github.com/oscal-compass/oscal-sdk-go/models/components"
	"github.com/oscal-compass/oscal-sdk-go/rules"
	"github.com/oscal-compass/oscal-sdk-go/settings"
	"github.com/stretchr/testify/require"
)

func TestGenereateAssessmentResults(t *testing.T) {

	pvpResults := []*policy.PVPResult{
		{
			ObservationsByCheck: []policy.ObservationByCheck{
				{
					Title:       "etcd_cert_file",
					Description: "Ensure that the --cert-file argument is set as appropriate",
					CheckID:     "etcd_cert_file",
					Methods:     []string{"test_method_1"},
					Subjects:    []policy.Subject{{Title: "test_subject_1"}},
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

	compDef := readCompDef(t)

	r := Reporter{
		store: prepMemoryStore(t, compDef),
	}

	implementationSettings := prepImplementationSettings(t, compDef)

	ar, err := r.GenerateAssessmentResults(context.TODO(), "https://...", &implementationSettings, pvpResults)
	require.NoError(t, err)

	require.Len(t, ar.Results, 1)

	require.Len(t, *ar.Results[0].Observations, 1)

	oscalObs := *ar.Results[0].Observations
	require.Equal(t, oscalObs[0].Title, pvpResults[0].ObservationsByCheck[0].Title)

}

// Load test component definition JSON
func readCompDef(t *testing.T) oscalTypes.ComponentDefinition {
	testDataPath := filepath.Join("../test/testdata", "component-definition-test.json")

	file, err := os.Open(testDataPath)
	require.NoError(t, err)

	definition, err := generators.NewComponentDefinition(file)
	require.NoError(t, err)
	require.NotNil(t, definition)

	return *definition
}

// Create a memory store using test compdef
func prepMemoryStore(t *testing.T, testComp oscalTypes.ComponentDefinition) *rules.MemoryStore {

	testMemoryStore := rules.NewMemoryStore()

	var comps []components.Component
	for _, cp := range *testComp.Components {
		adapters := components.NewDefinedComponentAdapter(cp)
		comps = append(comps, adapters)
	}
	err := testMemoryStore.IndexAll(comps)
	require.NoError(t, err)

	return testMemoryStore
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
