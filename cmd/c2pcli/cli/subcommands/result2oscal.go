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

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/config"
	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
)

func NewResult2OSCAL() *cobra.Command {
	options := NewOptions()
	command := &cobra.Command{
		Use:   "result2oscal",
		Short: "Transform policy result artifacts to OSCAL Assessment Results.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return setupViper(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.Unmarshal(options); err != nil {
				return err
			}
			if err := options.Validate(); err != nil {
				return err
			}
			return runResult2Policy(cmd.Context(), options)
		},
	}

	fs := command.Flags()
	fs.StringP("out", "o", "./assessment-results.json", "path to output OSCAL Assessment Results")
	BindCommonFlags(fs)

	return command
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

	err = pkg.WriteObjToJsonFile(viper.GetString("out"), oscalModels)
	if err != nil {
		return err
	}
	return nil
}
