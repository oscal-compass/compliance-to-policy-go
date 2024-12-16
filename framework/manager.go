/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"context"
	"fmt"

	"github.com/oscal-compass/oscal-sdk-go/rules"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/config"
	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

// PluginManager manages the plugin lifecycle and compliance-to-policy
// workflows.
type PluginManager struct {
	pluginDir     string
	store         rules.Store
	titleByIds    map[string]string
	clientFactory plugin.ClientFactoryFunc
}

// NewPluginManager creates a new instance of a PluginManager from a C2PConfig that can be used to
// interact with support plugins.
//
// It supports the plugin lifecycle with the following methods:
//   - Finding and initializing plugins: LaunchPolicyPlugins()
//   - Execution - GeneratePolicy() and AggregateResults()
//   - Clean/Stop - Clean()
func NewPluginManager(cfg *config.C2PConfig) (*PluginManager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Resolve all the options that were set in C2P into loaded
	// that are immediately usable for the PluginManager.
	rulesStore, pluginIDMap, err := config.ResolveOptions(cfg)
	if err != nil {
		return nil, err
	}

	return &PluginManager{
		pluginDir:     cfg.PluginDir,
		store:         rulesStore,
		clientFactory: plugin.ClientFactory(cfg.Logger),
		titleByIds:    pluginIDMap,
	}, nil
}

// LaunchPolicyPlugins retrieves information for the plugins that have been requested
// in the C2PConfig and launches each plugin to make it ready for use with GeneratePolicy() and
// AggregateResults().
func (m *PluginManager) LaunchPolicyPlugins() (map[string]policy.Provider, error) {
	var providerIds []string
	for id := range m.titleByIds {
		providerIds = append(providerIds, id)
	}
	pluginsByIds := make(map[string]policy.Provider)

	pluginManifests, err := plugin.FindPlugins(
		m.pluginDir,
		plugin.WithProviderIds(providerIds),
		plugin.WithPluginType(plugin.PVPPluginName),
	)
	if err != nil {
		return pluginsByIds, err
	}

	for _, manifest := range pluginManifests {
		policyPlugin, err := plugin.NewPolicyPlugin(manifest, m.clientFactory)
		if err != nil {
			return pluginsByIds, err
		}
		pluginsByIds[manifest.ID] = policyPlugin
	}
	return pluginsByIds, nil
}

// GeneratePolicy identifies policy configuration for each provider in the given pluginSet to execute the Generate() method
// each policy.Provider.
func (m *PluginManager) GeneratePolicy(ctx context.Context, pluginSet map[string]policy.Provider) error {
	for providerId, policyPlugin := range pluginSet {
		componentTitle, ok := m.titleByIds[providerId]
		if !ok {
			return fmt.Errorf("missing title for provider %s", providerId)
		}

		ruleSets, err := m.store.FindByComponent(ctx, componentTitle)
		if err != nil {
			return err
		}
		if err := policyPlugin.Generate(ruleSets); err != nil {
			return err
		}
	}
	return nil
}

// AggregateResults identifies policy configuration for each provider in the given pluginSet to execute the GetResults() method
// each policy.Provider.
func (m *PluginManager) AggregateResults(ctx context.Context, pluginSet map[string]policy.Provider) ([]policy.PVPResult, error) {
	var allResults []policy.PVPResult
	for providerId, policyPlugin := range pluginSet {
		// get the provider ids here to grab the policy
		componentTitle, ok := m.titleByIds[providerId]
		if !ok {
			return allResults, fmt.Errorf("missing title for provider %s", providerId)
		}

		ruleSets, err := m.store.FindByComponent(ctx, componentTitle)
		if err != nil {
			return allResults, err
		}

		pluginResults, err := policyPlugin.GetResults(ruleSets)
		if err != nil {
			return allResults, err
		}
		allResults = append(allResults, pluginResults)
	}
	return allResults, nil
}

// Clean deletes instances of plugin clients that have been created using LaunchPolicyPlugins.
func (m *PluginManager) Clean() {
	plugin.Cleanup()
}
