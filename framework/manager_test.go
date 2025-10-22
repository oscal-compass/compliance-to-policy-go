/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"bytes"
	"context"
	"sort"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/oscal-compass/compliance-to-policy-go/v2/logging"
	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

func TestNewPluginManager(t *testing.T) {
	cfg := &C2PConfig{
		PluginDir:         ".",
		PluginManifestDir: ".",
	}

	manager, err := NewPluginManager(cfg)
	require.NoError(t, err)
	require.NotNil(t, manager)

	require.Equal(t, ".", manager.pluginDir)
}

func TestPluginManager_Configure(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PluginDir = "."
	cfg.PluginManifestDir = "."

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
	pluginMap := func(id plugin.ID) map[string]string {
		return map[string]string{"option1": "override"}
	}

	// Create pluginSet
	providerTestObj := new(policyProvider)
	providerTestObj.
		On("Configure", map[string]string{"option 2": "value", "option1": "override"}).
		Return(nil)
	err = pluginManager.configurePlugin(context.Background(), providerTestObj, manifest, pluginMap)
	require.NoError(t, err)
	providerTestObj.AssertExpectations(t)
}

// policyProvider is a mocked implementation of policy.Provider.
type policyProvider struct {
	mock.Mock
}

func (p *policyProvider) Configure(_ context.Context, option map[string]string) error {
	args := p.Called(option)
	return args.Error(0)
}

func (p *policyProvider) Generate(_ context.Context, policyRules policy.Policy) error {
	sort.SliceStable(policyRules, func(i, j int) bool {
		return policyRules[i].Rule.ID > policyRules[j].Rule.ID
	})
	args := p.Called(policyRules)
	return args.Error(0)
}

func (p *policyProvider) GetResults(_ context.Context, policyRules policy.Policy) (policy.PVPResult, error) {
	sort.SliceStable(policyRules, func(i, j int) bool {
		return policyRules[i].Rule.ID > policyRules[j].Rule.ID
	})
	args := p.Called(policyRules)
	return args.Get(0).(policy.PVPResult), args.Error(1)
}

func TestNewPluginManager_ConfiguresGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	customLogger := hclog.New(&hclog.LoggerOptions{
		Name:   "test-logger",
		Output: &buf,
		Level:  hclog.Debug,
	})

	cfg := &C2PConfig{
		PluginDir:         ".",
		PluginManifestDir: ".",
		Logger:            customLogger,
	}

	_, err := NewPluginManager(cfg)
	require.NoError(t, err)

	log := logging.GetLogger("test-component")
	log.Debug("test message")

	require.Contains(t, buf.String(), "test message")
	require.Contains(t, buf.String(), "test-logger.test-component")
}
