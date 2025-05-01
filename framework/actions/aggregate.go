/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package actions

import (
	"context"
	"fmt"

	"github.com/oscal-compass/oscal-sdk-go/settings"

	"github.com/oscal-compass/compliance-to-policy-go/v2/logging"
	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

// AggregateResults action identifies policy configuration for each provider in the given pluginSet to execute the GetResults() method
// each policy.Provider.
//
// The rule set passed to each plugin can be configured with compliance specific settings based on the InputContext.
func AggregateResults(ctx context.Context, inputContext *InputContext, pluginSet map[plugin.ID]policy.Provider) ([]policy.PVPResult, error) {
	var allResults []policy.PVPResult
	log := logging.GetLogger("aggregator")
	for providerId, policyPlugin := range pluginSet {
		componentTitle, err := inputContext.ProviderTitle(providerId)
		if err != nil {
			return nil, err
		}
		log.Debug(fmt.Sprintf("Aggregating results for provider %s", providerId))
		appliedRuleSet, err := settings.ApplyToComponent(ctx, componentTitle, inputContext.Store(), inputContext.Settings)
		if err != nil {
			return allResults, fmt.Errorf("failed to get rule sets for component %s: %w", componentTitle, err)
		}

		pluginResults, err := policyPlugin.GetResults(appliedRuleSet)
		if err != nil {
			return allResults, fmt.Errorf("plugin %s: %w", providerId, err)
		}
		allResults = append(allResults, pluginResults)
	}
	return allResults, nil
}
