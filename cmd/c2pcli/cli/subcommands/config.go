/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package subcommands

import (
	"os"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/settings"
	"github.com/oscal-compass/oscal-sdk-go/validation"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/config"
)

// Config returns a populated C2PConfig for the CLI to use.
func Config(option *Options) (*config.C2PConfig, error) {
	c2pConfig := config.DefaultConfig()
	componentPath := option.Definition
	pluginsPath := option.PluginDir
	if pluginsPath != "" {
		c2pConfig.PluginDir = pluginsPath
	}

	compDef, err := loadCompDef(componentPath)
	if err != nil {
		return nil, err
	}
	c2pConfig.ComponentDefinitions = []oscalTypes.ComponentDefinition{*compDef}
	return c2pConfig, nil
}

func loadCompDef(path string) (*oscalTypes.ComponentDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	compDef, err := models.NewComponentDefinition(file, validation.NewSchemaValidator())
	if err != nil {
		return nil, err
	}
	return compDef, nil
}

// Settings returns extracted compliance settings from a given component definition implementation using the C2PConfig.
func Settings(frameworkConfig *config.C2PConfig, option *Options) (*settings.ImplementationSettings, error) {
	var implementation []oscalTypes.ControlImplementationSet
	for _, comp := range frameworkConfig.ComponentDefinitions {
		for _, cp := range *comp.Components {
			if cp.ControlImplementations != nil {
				implementation = append(implementation, *cp.ControlImplementations...)
			}
		}
	}
	return settings.Framework(option.Name, implementation)
}
