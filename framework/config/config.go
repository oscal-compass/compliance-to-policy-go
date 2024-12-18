/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/rules"

	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
)

const (
	// Only validation components are plugins
	pluginComponentType = "validation"
	// DefaultPluginPath default location c2p will look for plugins
	DefaultPluginPath = "c2p-plugins"
)

// C2PConfig represents configuration options for the C2P framework.PluginManager.
type C2PConfig struct {
	// PluginDir is the directory where the PluginManager searches
	// for installed plugins.
	PluginDir string
	// Logger is the logging implementation using in the PluginManager and
	// plugin clients.
	Logger hclog.Logger
	// ComponentDefinitions
	ComponentDefinitions []oscalTypes.ComponentDefinition
}

var defaultLogger = hclog.New(&hclog.LoggerOptions{
	Name:   "c2p",
	Output: os.Stdout,
	Level:  hclog.Info,
})

// DefaultConfig returns the default configuration.
func DefaultConfig() *C2PConfig {
	return &C2PConfig{
		PluginDir:            DefaultPluginPath,
		Logger:               defaultLogger,
		ComponentDefinitions: make([]oscalTypes.ComponentDefinition, 0),
	}
}

// Validate returns an error if C2PConfig has invalid fields.
func (c *C2PConfig) Validate() error {
	// Sanitize the plugin directory input
	c.PluginDir = strings.TrimSpace(c.PluginDir)
	c.PluginDir = filepath.Clean(c.PluginDir)
	if strings.TrimSpace(c.PluginDir) == "" {
		return fmt.Errorf("plugin directory cannot be empty")
	}
	if _, err := os.Stat(c.PluginDir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("plugin directory %s does not exist: %w", c.PluginDir, err)
		}
		return err
	}
	if len(c.ComponentDefinitions) == 0 {
		return fmt.Errorf("component definitions not set")
	}
	if c.Logger == nil {
		c.Logger = defaultLogger
	}
	return nil
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

// ResolveOptions returns processed options from the C2PConfig to configure framework.PluginManager.
// The returned options include a loaded rules.MemoryStore and a plugin identifier map to the original
// OSCAL component title.
func ResolveOptions(config *C2PConfig) (*rules.MemoryStore, map[string]string, error) {
	allComponents, titleByID, err := resolveOptions(config)
	if err != nil {
		return nil, nil, err
	}
	store, err := rules.NewMemoryStoreFromComponents(allComponents)
	if err != nil {
		return store, titleByID, err
	}
	return store, titleByID, nil
}

// resolveOptions returns processed OSCAL Components and a plugin identifier map. This performs most
// of the logic in ResolveOptions, but is broken out to make it easier to test.
func resolveOptions(config *C2PConfig) ([]oscalTypes.DefinedComponent, map[string]string, error) {
	var allComponents []oscalTypes.DefinedComponent
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
			allComponents = append(allComponents, component)
		}
	}
	return allComponents, titleByID, nil
}
