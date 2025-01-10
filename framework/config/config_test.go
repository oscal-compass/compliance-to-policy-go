/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"os"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/generators"
	"github.com/stretchr/testify/require"
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

func TestC2PConfig_Validate(t *testing.T) {
	config := DefaultConfig()
	require.EqualError(t, config.Validate(), "plugin directory c2p-plugins does not exist: stat c2p-plugins: no such file or directory")
	config.PluginDir = "."
	require.EqualError(t, config.Validate(), "component definitions not set")
	config.Logger = nil
	config.ComponentDefinitions = []oscalTypes.ComponentDefinition{
		{},
	}
	require.NoError(t, config.Validate())
	require.NotNil(t, config.Logger)
}

func TestResolveOptions(t *testing.T) {
	testDataPath := "../../test/testdata/component-definition-test.json"
	testFile, err := os.Open(testDataPath)
	require.NoError(t, err)
	compDef, err := generators.NewComponentDefinition(testFile)
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

func TestDefaultConfig(t *testing.T) {
	defaultConfig := DefaultConfig()
	require.Equal(t, defaultConfig.PluginDir, DefaultPluginPath)
	require.NotNil(t, defaultConfig.Logger)
}
