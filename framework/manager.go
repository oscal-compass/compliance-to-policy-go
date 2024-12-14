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

// NewPluginManager creates a new instance of a Plugin Manager from a C2PConfig.
func NewPluginManager(cfg *config.C2PConfig) (*PluginManager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

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

// LoadPolicyPlugins retrieves information for the plugins that have been requested
// in the C2PConfig.
func (m *PluginManager) LoadPolicyPlugins() (map[string]policy.Provider, error) {
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

func (m *PluginManager) Stop() {
	plugin.Cleanup()
}
