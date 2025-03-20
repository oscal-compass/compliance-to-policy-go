/*
Copyright 2023 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package subcommands

import (
	"context"
	"fmt"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/spf13/cobra"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/config"
	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
)

func NewResult2OSCAL(logger hclog.Logger) *cobra.Command {
	options := NewOptions()
	options.logger = logger

	command := &cobra.Command{
		Use:   "result2oscal",
		Short: "Transform policy result artifacts to OSCAL Assessment Results.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Complete(cmd); err != nil {
				return err
			}
			if err := validateResult2OSCAL(options); err != nil {
				return err
			}
			return runResult2Policy(cmd.Context(), options)
		},
	}

	fs := command.Flags()
	fs.StringP("out", "o", "./assessment-results.json", "path to output OSCAL Assessment Results")
	BindPluginFlags(fs)

	return command
}

// validateResult2OSCAL required options with no defaults
// are in place.
func validateResult2OSCAL(options *Options) error {
	if options.Name == "" {
		return &ConfigError{Option: Name}
	}
	if options.Definition == "" {
		return &ConfigError{Option: ComponentDefinition}
	}
	return nil
}

func runResult2Policy(ctx context.Context, option *Options) error {
	frameworkConfig, err := Config(option)
	if err != nil {
		return err
	}

	settings, err := Settings(frameworkConfig, option)
	if err != nil {
		return err
	}

	manager, err := framework.NewPluginManager(frameworkConfig)
	if err != nil {
		return err
	}
	foundPlugins, err := manager.FindRequestedPlugins()
	if err != nil {
		return err
	}

	var configSelections config.PluginConfig = func(pluginID string) map[string]string {
		return option.Plugins[pluginID]
	}
	launchedPlugins, err := manager.LaunchPolicyPlugins(foundPlugins, configSelections)
	if err != nil {
		return err
	}
	defer manager.Clean()

	results, err := manager.AggregateResults(ctx, launchedPlugins, settings.AllSettings())
	if err != nil {
		return err
	}

	reporter, err := framework.NewReporter(frameworkConfig)
	if err != nil {
		return err
	}

	assessmentResults, err := reporter.GenerateAssessmentResults(ctx, "REPLACE_ME", settings, results)
	oscalModels := oscalTypes.OscalModels{
		AssessmentResults: &assessmentResults,
	}

	// Validate before writing out
	option.logger.Info("Validating generated assessment results")
	validator := validation.NewSchemaValidator()
	if err := validator.Validate(oscalModels); err != nil {
		return err
	}

	option.logger.Info(fmt.Sprintf("Writing assessment results to %s.", option.Output))
	err = pkg.WriteObjToJsonFile(option.Output, oscalModels)
	if err != nil {
		return err
	}
	return nil
}
