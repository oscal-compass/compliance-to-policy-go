/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindPlugins(t *testing.T) {
	tests := []struct {
		name         string
		testDataPath string
		options      []FindOption
		wantError    string
		wantMeta     []Metadata
	}{
		{
			name:         "Valid/AllPlugins",
			testDataPath: "testdata/plugins",
			options:      []FindOption{},
			wantMeta: []Metadata{
				{
					ID:          "another-testplugin",
					Description: "My example test plugin",
					Version:     "0.0.1",
					Types:       []string{"pvp", "remediation"},
				},
				{
					ID:          "testplugin",
					Description: "My test plugin",
					Version:     "0.0.0",
					Types:       []string{"pvp"},
				},
			},
		},
		{
			name:         "Valid/MatchingPlugins",
			testDataPath: "testdata/plugins",
			options: []FindOption{
				WithProviderIds([]string{"testplugin"}),
			},
			wantMeta: []Metadata{
				{
					ID:          "testplugin",
					Description: "My test plugin",
					Version:     "0.0.0",
					Types:       []string{"pvp"},
				},
			},
		},
		{
			name:         "Valid/MatchingPluginOfType",
			testDataPath: "testdata/plugins",
			options: []FindOption{
				WithPluginType("remediation"),
			},
			wantMeta: []Metadata{
				{
					ID:          "another-testplugin",
					Description: "My example test plugin",
					Version:     "0.0.1",
					Types:       []string{"pvp", "remediation"},
				},
			},
		},
		{
			name:         "InValid/PluginNameInvalid",
			testDataPath: "testdata/invalid-plugins",
			options: []FindOption{
				WithProviderIds([]string{"INVALID"}),
			},
			wantError: "invalid plugin id \"INVALID\" in manifest c2p-INVALID-manifest.json",
		},
		{
			name:         "InValid/PluginNameMismatch",
			testDataPath: "testdata/invalid-plugins",
			options: []FindOption{
				WithProviderIds([]string{"testplugin"}),
			},
			wantError: "invalid plugin id \"testplugin2\" in manifest c2p-testplugin-manifest.json",
		},
		{
			name:         "Failure/NoPlugins",
			testDataPath: "testdata/",
			wantError:    "no plugins found in testdata/",
		},
		{
			name:         "Failure/NoMatchingPlugins",
			testDataPath: "testdata/plugins",
			options: []FindOption{
				WithProviderIds([]string{"example"}),
			},
			wantError: "failed to find plugin \"example\" in plugin installation location",
		},
		{
			name:         "Failure/NoPluginsOfType",
			testDataPath: "testdata/plugins",
			options: []FindOption{
				WithProviderIds([]string{"testplugin"}),
				WithPluginType("remediation"),
			},
			wantError: "no plugins found in testdata/plugins with matching criteria",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			manifests, err := FindPlugins(c.testDataPath, c.options...)
			if c.wantError != "" {
				require.EqualError(t, err, c.wantError)
			} else {
				require.NoError(t, err)
				var foundMeta []Metadata
				for _, m := range manifests {
					foundMeta = append(foundMeta, m.Metadata)
				}
				// Eliminate flakiness with sorting
				sort.SliceStable(foundMeta, func(i, j int) bool {
					return foundMeta[i].ID < foundMeta[j].ID
				})
				require.Equal(t, c.wantMeta, foundMeta)
			}
		})
	}
}
