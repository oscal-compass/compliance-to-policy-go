/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/models/components"
	"github.com/oscal-compass/oscal-sdk-go/rules"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

func TestOscal2Policy(t *testing.T) {
	policyDir := pkg.PathFromPkgDirectory("./testdata/ocm/policies")
	tmpOutputDir := t.TempDir()

	tempDirPath := pkg.PathFromPkgDirectory("./testdata/_test")
	err := os.MkdirAll(tempDirPath, os.ModePerm)
	assert.NoError(t, err, "Should not happen")
	tempDir := pkg.NewTempDirectory(tempDirPath)

	testPolicy := createPolicy(t)
	plugin := NewPlugin()
	plugin.config.PoliciesDir = policyDir
	plugin.config.Namespace = "test"
	plugin.config.PolicySetName = "Managed Kubernetes"
	plugin.config.TempDir = tempDir.GetTempDir()
	plugin.config.OutputDir = tmpOutputDir
	plugin.config.PolicyResultsDir = tmpOutputDir
	require.NoError(t, plugin.Generate(testPolicy))
}

func TestResult2Oscal(t *testing.T) {
	policyResultsDir := pkg.PathFromPkgDirectory("./testdata/ocm/policy-results")
	tempDirPath := pkg.PathFromPkgDirectory("./testdata/_test")
	err := os.MkdirAll(tempDirPath, os.ModePerm)
	assert.NoError(t, err, "Should not happen")
	testPolicy := createPolicy(t)
	reporter := NewResultToOscal(testPolicy, policyResultsDir, "c2p", "Managed Kubernetes")
	results, err := reporter.GenerateResults()
	assert.NoError(t, err, "Should not happen")
	expected := policy.PVPResult{
		ObservationsByCheck: []policy.ObservationByCheck{
			{
				Title:       "test_configuration_check",
				Description: "Observation of policy policy-high-scan",
				CheckID:     "policy-high-scan",
				Methods:     []string{"TEST-AUTOMATED"},
				Subjects: []policy.Subject{
					{
						Title:      "Cluster Name: cluster1",
						Type:       "resource",
						ResourceID: "322b6a68-006e-11f0-a98b-88a4c2f0e4d9",
						Result:     policy.ResultFail,
						Reason:     "",
					},
					{
						Title:      "Cluster Name: cluster2",
						Type:       "resource",
						ResourceID: "322b6a68-006e-11f0-a98b-88a4c2f0e4d9",
						Result:     policy.ResultFail,
						Reason:     "",
					},
				},
				Props: []policy.Property{
					{
						Name:  "assessment-rule-id",
						Value: "test_configuration_check",
					},
				},
			},
			{
				Title:       "test_proxy_check",
				Description: "Observation of policy policy-deployment",
				CheckID:     "policy-deployment",
				Methods:     []string{"TEST-AUTOMATED"},
				Subjects: []policy.Subject{
					{
						Title:      "Cluster Name: cluster1",
						Type:       "resource",
						ResourceID: "322b6a68-006e-11f0-a98b-88a4c2f0e4d9",
						Result:     policy.ResultFail,
						Reason:     "ddd",
					},
					{
						Title:      "Cluster Name: cluster2",
						Type:       "resource",
						ResourceID: "322b6a68-006e-11f0-a98b-88a4c2f0e4d9",
						Result:     policy.ResultFail,
						Reason:     "",
					},
				},
				Props: []policy.Property{
					{
						Name:  "assessment-rule-id",
						Value: "test_proxy_check",
					},
				},
			},
			{
				Title:       "test_rbac_check",
				Description: "Observation of policy policy-disallowed-roles",
				CheckID:     "policy-disallowed-roles",
				Methods:     []string{"TEST-AUTOMATED"},
				Subjects: []policy.Subject{
					{
						Title:      "Cluster Name: cluster1",
						Type:       "resource",
						ResourceID: "322b6a68-006e-11f0-a98b-88a4c2f0e4d9",
						Result:     policy.ResultPass,
						Reason:     "",
					},
					{
						Title:      "Cluster Name: cluster2",
						Type:       "resource",
						ResourceID: "322b6a68-006e-11f0-a98b-88a4c2f0e4d9",
						Result:     policy.ResultPass,
						Reason:     "",
					},
				},
				Props: []policy.Property{
					{Name: "assessment-rule-id", Value: "test_rbac_check"}}},
		},
	}
	diff := cmp.Diff(expected, results,
		cmpopts.IgnoreFields(policy.ObservationByCheck{}, "Collected"),
		cmpopts.IgnoreFields(policy.Subject{}, "ResourceID"),
		cmpopts.IgnoreFields(policy.Subject{}, "Reason"),
		cmpopts.SortSlices(func(i, j policy.ObservationByCheck) bool {
			return i.Title < j.Title
		}),
	)
	require.Equal(t, diff, "")
}

func TestConfigure(t *testing.T) {
	plugin := NewPlugin()
	policyDir := pkg.PathFromPkgDirectory("./testdata/ocm/policies")
	configuration := map[string]string{
		"policy-dir": policyDir,
	}
	err := plugin.Configure(configuration)
	require.EqualError(t, err, "policy set name must be set")

	configuration["policy-set-name"] = "set"

	configuration["policy-dir"] = "not-exist"
	err = plugin.Configure(configuration)
	require.EqualError(t, err, "path \"not-exist\": stat not-exist: no such file or directory")

	configuration["policy-dir"] = policyDir
	err = plugin.Configure(configuration)
	require.NoError(t, err)
}

func createPolicy(t *testing.T) []extensions.RuleSet {
	cdPath := pkg.PathFromPkgDirectory("./testdata/ocm/component-definition.json")

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

	ruleSets, err := store.FindByComponent(context.TODO(), "Managed Kubernetes")
	require.NoError(t, err)
	return ruleSets
}
