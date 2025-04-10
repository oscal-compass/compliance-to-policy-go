/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"context"
	"os"
	"testing"

	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/models/components"
	"github.com/oscal-compass/oscal-sdk-go/rules"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
)

func TestOscal2Policy(t *testing.T) {
	policyDir := pkg.PathFromPkgDirectory("./testdata/kyverno/policy-resources")

	tempDirPath := pkg.PathFromPkgDirectory("./testdata/_test")
	err := os.MkdirAll(tempDirPath, os.ModePerm)
	assert.NoError(t, err, "Should not happen")
	tempDir := pkg.NewTempDirectory(tempDirPath)

	policyExample := createPolicy(t)
	o2p := NewOscal2Policy(policyDir, tempDir)
	err = o2p.Generate(policyExample)
	assert.NoError(t, err, "Should not happen")
}

func TestConfigure(t *testing.T) {
	plugin := NewPlugin()
	configuration := map[string]string{
		"policy-dir": "not-exist",
	}
	err := plugin.Configure(configuration)
	require.EqualError(t, err, "path \"not-exist\": stat not-exist: no such file or directory")

	policyDir := pkg.PathFromPkgDirectory("./testdata/kyverno/policy-resources")
	configuration["policy-dir"] = policyDir
	err = plugin.Configure(configuration)
	require.NoError(t, err)
}

func createPolicy(t *testing.T) []extensions.RuleSet {
	cdPath := pkg.PathFromPkgDirectory("./testdata/kyverno/component-definition.json")

	file, err := os.Open(cdPath)
	require.NoError(t, err)
	defer file.Close()

	compDef, err := models.NewComponentDefinition(file, validation.NoopValidator{})

	require.NotNil(t, compDef)
	require.NotNil(t, compDef.Components)

	var allComponents []components.Component
	for _, comp := range *compDef.Components {
		adapter := components.NewDefinedComponentAdapter(comp)
		allComponents = append(allComponents, adapter)
	}

	store := rules.NewMemoryStore()
	require.NoError(t, store.IndexAll(allComponents))

	ruleSets, err := store.FindByComponent(context.TODO(), "Kyverno")
	require.NoError(t, err)
	return ruleSets
}
