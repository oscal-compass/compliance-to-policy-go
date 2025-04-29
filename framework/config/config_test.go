/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/stretchr/testify/require"
)

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

func TestDefaultConfig(t *testing.T) {
	defaultConfig := DefaultConfig()
	require.Equal(t, defaultConfig.PluginDir, DefaultPluginPath)
	require.Equal(t, defaultConfig.PluginManifestDir, DefaultPluginManifestPath)
	require.NotNil(t, defaultConfig.Logger)
}
