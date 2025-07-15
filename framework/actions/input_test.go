/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package actions

import (
	"os"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/models/components"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/stretchr/testify/require"

	"github.com/oscal-compass/compliance-to-policy-go/v2/internal/utils"
	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
)

func TestGetPluginIDFromComponent(t *testing.T) {
	tests := []struct {
		name      string
		component oscalTypes.DefinedComponent
		expected  plugin.ID
		wantError string
	}{
		{
			name: "Valid/ExactID",
			component: oscalTypes.DefinedComponent{
				Title: "myplugin",
			},
			expected:  "myplugin",
			wantError: "",
		},
		{
			name: "Valid/WithWhiteSpace",
			component: oscalTypes.DefinedComponent{
				Title: " myplugin ",
			},
			expected:  "myplugin",
			wantError: "",
		},
		{
			name: "Valid/UpperCase",
			component: oscalTypes.DefinedComponent{
				Title: "MYPLUGIN",
			},
			expected:  "myplugin",
			wantError: "",
		},
		{
			name: "Invalid/PluginNotMatchPattern",
			component: oscalTypes.DefinedComponent{
				Title: "my plugin",
			},
			expected:  "",
			wantError: "invalid plugin id my plugin",
		},
		{
			name: "Invalid/EmptyTitle",
			component: oscalTypes.DefinedComponent{
				Title: "",
			},
			expected:  "",
			wantError: "component is missing a title",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			compAdapter := components.NewDefinedComponentAdapter(c.component)
			id, err := GetPluginIDFromComponent(compAdapter)
			if c.wantError == "" {
				require.NoError(t, err)
				require.Equal(t, c.expected, id)
			} else {
				require.EqualError(t, err, c.wantError)
			}
		})
	}
}

// inputContextHelper to support other testing in the package
func inputContextHelper(t *testing.T) *InputContext {
	testDataPath := utils.PathFromInternalDirectory("./testdata/oscal/component-definition-test.json")
	file, err := os.Open(testDataPath)
	require.NoError(t, err)
	definition, err := models.NewComponentDefinition(file, validation.NoopValidator{})
	require.NoError(t, err)

	var allComponents []components.Component
	for _, component := range *definition.Components {
		compAdapter := components.NewDefinedComponentAdapter(component)
		allComponents = append(allComponents, compAdapter)
	}

	inputContext, err := NewContext(allComponents)
	require.NoError(t, err)
	return inputContext
}
