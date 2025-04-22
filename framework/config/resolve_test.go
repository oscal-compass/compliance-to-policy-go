/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"os"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/stretchr/testify/require"

	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
)

func TestGetPluginIDFromComponent(t *testing.T) {
	tests := []struct {
		name      string
		component oscalTypes.DefinedComponent
		expected  string
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
			id, err := GetPluginIDFromComponent(c.component)
			if c.wantError == "" {
				require.NoError(t, err)
				require.Equal(t, c.expected, id)
			} else {
				require.EqualError(t, err, c.wantError)
			}
		})
	}
}

func TestResolveOptions(t *testing.T) {
	testDataPath := pkg.PathFromPkgDirectory("./testdata/oscal/component-definition-test.json")
	testFile, err := os.Open(testDataPath)
	require.NoError(t, err)
	compDef, err := models.NewComponentDefinition(testFile, validation.NoopValidator{})
	require.NoError(t, err)
	config := &C2PConfig{
		ComponentDefinitions: []oscalTypes.ComponentDefinition{
			*compDef,
		},
	}

	components, pluginMap, err := resolveOptions(config)
	require.NoError(t, err)

	expectedPluginMap := map[string]string{
		"mypvpvalidator": "MyPVPValidator",
	}
	require.Len(t, components, 2)
	require.Equal(t, expectedPluginMap, pluginMap)
}
