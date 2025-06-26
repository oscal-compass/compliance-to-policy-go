/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package actions

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/oscal-compass/oscal-sdk-go/settings"
	"golang.org/x/sync/semaphore"

	"github.com/oscal-compass/compliance-to-policy-go/v2/logging"
	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

// GeneratePolicy action identifies policy configuration for each provider in the given pluginSet to execute the Generate() method
// each policy.Provider.
//
// The rule set passed to each plugin can be configured with compliance specific settings based on the InputContext.
func GeneratePolicy(ctx context.Context, inputContext *InputContext, pluginSet map[plugin.ID]policy.Provider) error {
	log := logging.GetLogger("generator")

	sem := semaphore.NewWeighted(inputContext.MaxConcurrentWeight)
	var wg sync.WaitGroup
	errorCh := make(chan error, len(pluginSet))

	for providerId, policyPlugin := range pluginSet {
		wg.Add(1)

		go func(providerId plugin.ID, plugin policy.Provider) {
			defer wg.Done()

			if err := sem.Acquire(ctx, 1); err != nil {
				errorCh <- fmt.Errorf("%s failed to acquire semaphore: %w", providerId.String(), err)
				return
			}
			defer sem.Release(1)

			componentTitle, err := inputContext.ProviderTitle(providerId)
			if err != nil {
				if errors.Is(err, ErrMissingProvider) {
					log.Warn(fmt.Sprintf("skipping %s provider: missing validation component", providerId))
					return
				}
				errorCh <- err
				return
			}
			log.Debug(fmt.Sprintf("Generating policy for provider %s", providerId))

			appliedRuleSet, err := settings.ApplyToComponent(ctx, componentTitle, inputContext.Store(), inputContext.Settings)
			if err != nil {
				errorCh <- fmt.Errorf("failed to get rule sets for component %s: %w", componentTitle, err)
				return
			}
			if err := policyPlugin.Generate(appliedRuleSet); err != nil {
				errorCh <- fmt.Errorf("plugin %s: %w", providerId, err)
				return
			}
		}(providerId, policyPlugin)
	}

	go func() {
		wg.Wait()
		close(errorCh)
	}()

	var errs []error
	for err := range errorCh {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
