/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package subcommands

import (
	"context"
	"errors"
	"os"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/models/components"
	"github.com/oscal-compass/oscal-sdk-go/settings"
	"github.com/oscal-compass/oscal-sdk-go/transformers"
	"github.com/oscal-compass/oscal-sdk-go/validation"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/actions"
)

// Config returns a populated C2PConfig for the CLI to use.
func Config(option *Options) (*framework.C2PConfig, error) {
	c2pConfig := framework.DefaultConfig()
	pluginsPath := option.PluginDir
	if pluginsPath != "" {
		c2pConfig.PluginDir = pluginsPath
		c2pConfig.PluginManifestDir = pluginsPath
	}
	// Set logger
	c2pConfig.Logger = option.logger
	return c2pConfig, nil
}

func Context(ctx context.Context, option *Options) (*actions.InputContext, oscalTypes.AssessmentPlan, error) {
	compDef, err := loadCompDef(option.Definition)
	if err != nil {
		return nil, oscalTypes.AssessmentPlan{}, err
	}

	ap, err := transformers.ComponentDefinitionsToAssessmentPlan(ctx, []oscalTypes.ComponentDefinition{compDef}, option.Name)
	if err != nil {
		return nil, oscalTypes.AssessmentPlan{}, err
	}

	if ap.LocalDefinitions == nil || ap.LocalDefinitions.Activities == nil || ap.AssessmentAssets.Components == nil {
		return nil, oscalTypes.AssessmentPlan{}, errors.New("error converting component definition to assessment plan")
	}

	var allComponents []components.Component
	for _, component := range *ap.AssessmentAssets.Components {
		compAdapter := components.NewSystemComponentAdapter(component)
		allComponents = append(allComponents, compAdapter)
	}

	inputCtx, err := actions.NewContext(allComponents)
	if err != nil {
		return nil, oscalTypes.AssessmentPlan{}, err
	}

	apSettings := settings.NewAssessmentActivitiesSettings(*ap.LocalDefinitions.Activities)
	inputCtx.Settings = apSettings

	return inputCtx, *ap, nil
}

func loadCompDef(path string) (oscalTypes.ComponentDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return oscalTypes.ComponentDefinition{}, err
	}
	defer file.Close()
	compDef, err := models.NewComponentDefinition(file, validation.NewSchemaValidator())
	if err != nil {
		return oscalTypes.ComponentDefinition{}, err
	}

	if compDef == nil {
		return oscalTypes.ComponentDefinition{}, errors.New("component definition cannot be nil")
	}
	return *compDef, nil
}
