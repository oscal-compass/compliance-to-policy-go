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
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/config"
)

func NewOSCAL2Policy() *cobra.Command {
	options := NewOptions()
	command := &cobra.Command{
		Use:   "oscal2policy",
		Short: "Transform OSCAL to policy artifacts.",
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
			if options.Name == "" {
				return errors.New("name option must be set")
			}
			return runOSCAL2Policy(cmd.Context(), options)
		},
	}
	BindCommonFlags(command.Flags())
	return command
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
