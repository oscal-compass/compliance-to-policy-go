/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"os"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/generators"
	"github.com/stretchr/testify/require"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/config"
)

func TestNewPluginManager(t *testing.T) {
	testDataPath := "../test/testdata/component-definition-test.json"
	testFile, err := os.Open(testDataPath)
	require.NoError(t, err)
	compDef, err := generators.NewComponentDefinition(testFile)
	require.NoError(t, err)
	cfg := &config.C2PConfig{
		PluginDir: ".",
		ComponentDefinitions: []oscalTypes.ComponentDefinition{
			*compDef,
		},
	}
	manager, err := NewPluginManager(cfg)
	require.NoError(t, err)
	require.NotNil(t, manager)
}
