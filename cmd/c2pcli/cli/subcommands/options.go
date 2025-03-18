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
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const PluginConfigPath = "config-path"

func BindCommonFlags(fs *pflag.FlagSet) {
	fs.StringP("name", "n", "", "short name of the control source for the implementation to be evaluated.")
	fs.StringP("component-definition", "d", "", "path to component definition")
	fs.StringP("plugin-dir", "p", "", "Path to plugin directory. Defaults to `c2p-plugins`.")
	fs.StringP(PluginConfigPath, "c", "plugins.yaml", "Path to the configuration file for plugins.")
}

func SetupViper(cmd *cobra.Command) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			panic(err)
		}
	})
	if viper.IsSet(PluginConfigPath) {
		viper.SetConfigFile(viper.GetString(PluginConfigPath))
		return viper.ReadInConfig()
	}
	return nil
}

type Options struct {
	PluginDir         string                       `yaml:"plugin-dir" mapstructure:"plugin-dir"`
	Name              string                       `yaml:"name" mapstructure:"name"`
	Definition        string                       `yaml:"component-definition" mapstructure:"component-definition"`
	Catalog           string                       `yaml:"catalog" mapstructure:"catalog"`
	AssessmentResults string                       `yaml:"assessment-results" mapstructure:"assessment-results"`
	Plugins           map[string]map[string]string `yaml:"plugins" mapstructure:"plugins"`
	Output            string                       `yaml:"out" mapstructure:"out"`
}

func NewOptions() *Options {
	return &Options{
		Plugins: make(map[string]map[string]string),
	}
}

func (o *Options) Validate() error {
	if o.Definition == "" {
		return errors.New("component-definition option must be set")
	}
	return nil
}
