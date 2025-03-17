/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"fmt"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/models/components"
	"github.com/oscal-compass/oscal-sdk-go/rules"

	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
)

// ResolveOptions returns processed options from the C2PConfig to configure framework.PluginManager.
// The returned options include a loaded rules.MemoryStore and a plugin identifier map to the original
// OSCAL component title.
func ResolveOptions(config *C2PConfig) (*rules.MemoryStore, map[string]string, error) {
	allComponents, titleByID, err := resolveOptions(config)
	if err != nil {
		return nil, nil, err
	}
	store := rules.NewMemoryStore()
	err = store.IndexAll(allComponents)
	if err != nil {
		return store, titleByID, err
	}
	return store, titleByID, nil
}

// resolveOptions returns processed OSCAL Components and a plugin identifier map. This performs most
// of the logic in ResolveOptions, but is broken out to make it easier to test.
func resolveOptions(config *C2PConfig) ([]components.Component, map[string]string, error) {
	var allComponents []components.Component
	titleByID := make(map[string]string)
	for _, compDef := range config.ComponentDefinitions {
		if compDef.Components == nil {
			continue
		}
		for _, component := range *compDef.Components {
			if component.Type == pluginComponentType {
				pluginId, err := GetPluginIDFromComponent(component)
				if err != nil {
					return nil, nil, err
				}
				titleByID[pluginId] = component.Title
			}
			compAdapter := components.NewDefinedComponentAdapter(component)
			allComponents = append(allComponents, compAdapter)
		}
	}
	return allComponents, titleByID, nil
}

// GetPluginIDFromComponent returns the normalized plugin identifier defined by the OSCAL Component
// of type "validation".
func GetPluginIDFromComponent(component oscalTypes.DefinedComponent) (string, error) {
	title := strings.TrimSpace(component.Title)
	if title == "" {
		return "", fmt.Errorf("component is missing a title")
	}

	title = strings.ToLower(title)
	if !plugin.IdentifierPattern.MatchString(title) {
		return "", fmt.Errorf("invalid plugin id %s", title)
	}
	return title, nil
}
