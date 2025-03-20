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
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Options with no defaults
const (
	ConfigPath          = "config-path"
	ComponentDefinition = "component-definition"
	Name                = "name"
	Catalog             = "catalog"
)

// BindCommonFlags binds common flags for all commands.
func BindCommonFlags(fs *pflag.FlagSet) {
	fs.StringP(ComponentDefinition, "d", "", "path to component definition")
	fs.StringP(ConfigPath, "c", "c2p-config.yaml", "Path to the configuration for the C2P CLI.")
}

// BindPluginFlags binds flags for command that interact with the plugin manager.
func BindPluginFlags(fs *pflag.FlagSet) {
	BindCommonFlags(fs)
	fs.StringP("plugin-dir", "p", "c2p-plugins", "Path to plugin directory. Defaults to `c2p-plugins`.")
	fs.StringP(Name, "n", "", "short name of the control source for the implementation to be evaluated.")
}

// ConfigError is an error for missing configuration options
type ConfigError struct {
	Option string
}

func (c *ConfigError) Error() string {
	return fmt.Sprintf("%q option is not set", c.Option)
}

// Options define config options when for the CLI commands.
type Options struct {
	PluginDir         string                       `yaml:"plugin-dir" mapstructure:"plugin-dir"`
	Name              string                       `yaml:"name" mapstructure:"name"`
	Definition        string                       `yaml:"component-definition" mapstructure:"component-definition"`
	Catalog           string                       `yaml:"catalog" mapstructure:"catalog"`
	AssessmentResults string                       `yaml:"assessment-results" mapstructure:"assessment-results"`
	Plugins           map[string]map[string]string `yaml:"plugins" mapstructure:"plugins"`
	Output            string                       `yaml:"out" mapstructure:"out"`
	logger            hclog.Logger
}

// NewOptions returns an initialized Options struct.
func NewOptions() *Options {
	return &Options{
		Plugins: make(map[string]map[string]string),
	}
}

// Complete the options from the given command.
func (o *Options) Complete(cmd *cobra.Command) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			panic(err)
		}
	})

	// If a config path is set, read options from the config
	if viper.IsSet(ConfigPath) {
		viper.SetConfigFile(viper.GetString(ConfigPath))
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
		return viper.Unmarshal(o)
	}
	return nil
}
