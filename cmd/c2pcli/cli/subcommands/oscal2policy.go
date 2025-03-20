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

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/config"
)

func NewOSCAL2Policy(logger hclog.Logger) *cobra.Command {
	options := NewOptions()
	options.logger = logger

	command := &cobra.Command{
		Use:   "oscal2policy",
		Short: "Transform OSCAL to policy artifacts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Complete(cmd); err != nil {
				return err
			}
			if err := validateOSCAL2Policy(options); err != nil {
				return err
			}
			return runOSCAL2Policy(cmd.Context(), options)
		},
	}
	BindPluginFlags(command.Flags())
	return command
}

// validateOSCAL2Policy required options with no defaults
// are in place.
func validateOSCAL2Policy(options *Options) error {
	if options.Name == "" {
		return &ConfigError{Option: Name}
	}
	if options.Definition == "" {
		return &ConfigError{Option: ComponentDefinition}
	}
	return nil
}

func runOSCAL2Policy(ctx context.Context, option *Options) error {
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

	err = manager.GeneratePolicy(ctx, launchedPlugins, settings.AllSettings())
	if err != nil {
		return err
	}

	return nil
}
