/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"context"
	"os"
	"sort"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/settings"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/config"
	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

var (
	expectedCertFileRule = extensions.RuleSet{
		Rule: extensions.Rule{
			ID:          "etcd_cert_file",
			Description: "Ensure that the --cert-file argument is set as appropriate",
		},
		Checks: []extensions.Check{
			{
				ID:          "etcd_cert_file",
				Description: "Check that the --cert-file argument is set as appropriate",
			},
		},
	}
	expectedKeyFileRule = extensions.RuleSet{
		Rule: extensions.Rule{
			ID:          "etcd_key_file",
			Description: "Ensure that the --key-file argument is set as appropriate",
			Parameter: &extensions.Parameter{
				ID:          "file_name",
				Description: "A parameter for a file name",
			},
		},
		Checks: []extensions.Check{
			{
				ID:          "etcd_key_file",
				Description: "Check that the --key-file argument is set as appropriate",
			},
		},
	}
)

func TestNewPluginManager(t *testing.T) {
	testFile, err := os.Open(testDataPath)
	require.NoError(t, err)
	compDef, err := models.NewComponentDefinition(testFile, validation.NoopValidator{})
	require.NoError(t, err)
	cfg := &config.C2PConfig{
		PluginDir:         ".",
		PluginManifestDir: ".",
		ComponentDefinitions: []oscalTypes.ComponentDefinition{
			*compDef,
		},
	}
	manager, err := NewPluginManager(cfg)
	require.NoError(t, err)
	require.NotNil(t, manager)
}

func TestPluginManager_GeneratePolicy(t *testing.T) {
	cfg := prepConfig(t)
	pluginManager, err := NewPluginManager(cfg)
	require.NoError(t, err)

	// Create pluginSet
	providerTestObj := new(policyProvider)
	providerTestObj.On("Generate", policy.Policy{expectedCertFileRule}).Return(nil)
	pluginSet := map[string]policy.Provider{
		"mypvpvalidator": providerTestObj,
	}

	testSettings := settings.NewSettings(map[string]struct{}{"etcd_cert_file": {}}, map[string]string{})

	err = pluginManager.GeneratePolicy(context.TODO(), pluginSet, testSettings)
	require.NoError(t, err)
	providerTestObj.AssertExpectations(t)
}

func TestPluginManager_AggregateResults(t *testing.T) {
	cfg := prepConfig(t)
	pluginManager, err := NewPluginManager(cfg)
	require.NoError(t, err)

	wantResults := policy.PVPResult{
		ObservationsByCheck: []policy.ObservationByCheck{
			{
				Title:       "Example",
				Description: "Example description",
				CheckID:     "test-check",
			},
		},
	}

	updatedParam := &extensions.Parameter{
		ID:          "file_name",
		Description: "A parameter for a file name",
		Value:       "my_file",
	}

	updatedKeyFileRule := expectedKeyFileRule
	updatedKeyFileRule.Rule.Parameter = updatedParam

	// Create pluginSet
	providerTestObj := new(policyProvider)
	providerTestObj.On("GetResults", policy.Policy{updatedKeyFileRule}).Return(wantResults, nil)
	pluginSet := map[string]policy.Provider{
		"mypvpvalidator": providerTestObj,
	}

	testSettings := settings.NewSettings(map[string]struct{}{"etcd_key_file": {}}, map[string]string{"file_name": "my_file"})

	gotResults, err := pluginManager.AggregateResults(context.TODO(), pluginSet, testSettings)
	require.NoError(t, err)
	providerTestObj.AssertExpectations(t)
	require.Len(t, gotResults, 1)
}

func TestPluginManager_Configure(t *testing.T) {
	cfg := prepConfig(t)
	pluginManager, err := NewPluginManager(cfg)
	require.NoError(t, err)

	defaultValue := "value"
	// test options and manifest
	manifest := plugin.Manifest{
		Metadata: plugin.Metadata{
			ID: "myplugin",
		},
		Configuration: []plugin.ConfigurationOption{
			{
				Name:        "option1",
				Description: "Option 1",
				Required:    true,
			},
			{
				Name:        "option 2",
				Description: "Option 2",
				Required:    false,
				Default:     &defaultValue,
			},
		},
	}
	pluginMap := func(string) map[string]string {
		return map[string]string{"option1": "override"}
	}

	// Create pluginSet
	providerTestObj := new(policyProvider)
	providerTestObj.
		On("Configure", map[string]string{"option 2": "value", "option1": "override"}).
		Return(nil)
	err = pluginManager.configurePlugin(providerTestObj, manifest, pluginMap)
	require.NoError(t, err)
	providerTestObj.AssertExpectations(t)
}

// prepConfig returns an initialized C2PConfig to support the
// unit tests.
func prepConfig(t *testing.T) *config.C2PConfig {
	cfg := config.DefaultConfig()
	cfg.PluginDir = "."
	cfg.PluginManifestDir = "."
	file, err := os.Open(testDataPath)
	require.NoError(t, err)
	definition, err := models.NewComponentDefinition(file, validation.NoopValidator{})
	require.NoError(t, err)
	cfg.ComponentDefinitions = append(cfg.ComponentDefinitions, *definition)
	return cfg
}

// policyProvider is a mocked implementation of policy.Provider.
type policyProvider struct {
	mock.Mock
}

func (p *policyProvider) Configure(option map[string]string) error {
	args := p.Called(option)
	return args.Error(0)
}

func (p *policyProvider) Generate(policyRules policy.Policy) error {
	sort.SliceStable(policyRules, func(i, j int) bool {
		return policyRules[i].Rule.ID > policyRules[j].Rule.ID
	})
	args := p.Called(policyRules)
	return args.Error(0)
}

func (p *policyProvider) GetResults(policyRules policy.Policy) (policy.PVPResult, error) {
	sort.SliceStable(policyRules, func(i, j int) bool {
		return policyRules[i].Rule.ID > policyRules[j].Rule.ID
	})
	args := p.Called(policyRules)
	return args.Get(0).(policy.PVPResult), args.Error(1)
}
