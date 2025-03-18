/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package subcommands

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
)

func NewOSCAL2Posture(logger hclog.Logger) *cobra.Command {
	options := NewOptions()
	command := &cobra.Command{
		Use:   "oscal2posture",
		Short: "Generate Compliance Posture from OSCAL artifacts.",
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

			// Extra validation for this command
			if options.Catalog == "" {
				return fmt.Errorf("\"catalog\" flag must be set")
			}
			options.logger = logger
			return runOSCAL2Posture(options)
		},
	}
	fs := command.Flags()
	fs.String("catalog", "", "path to catalog.json")
	fs.StringP("assessment-results", "a", "./assessment-results.json", "path to assessment-results.json")
	fs.StringP("component-definition", "d", "", "path to component-definition.json")
	fs.StringP("out", "o", "-", "path to output file. Use '-' for stdout. Default '-'.")
	fs.StringP(PluginConfigPath, "c", "plugins.yaml", "Path to the configuration file for plugins.")
	return command
}

func runOSCAL2Posture(option *Options) error {
	arFile, err := os.Open(option.AssessmentResults)
	if err != nil {
		return err
	}
	defer arFile.Close()
	assessmentResults, err := models.NewAssessmentResults(arFile, validation.NewSchemaValidator())
	if err != nil {
		return fmt.Errorf("error loading assessment results: %w", err)
	}

	catalogFile, err := os.Open(option.Catalog)
	if err != nil {
		return err
	}
	defer catalogFile.Close()
	catalog, err := models.NewCatalog(catalogFile, validation.NewSchemaValidator())
	if err != nil {
		return fmt.Errorf("error loading catalog: %w", err)
	}

	compDefFile, err := os.Open(option.Definition)
	if err != nil {
		return err
	}
	defer compDefFile.Close()
	compDef, err := models.NewComponentDefinition(compDefFile, validation.NewSchemaValidator())
	if err != nil {
		return fmt.Errorf("error loading component definition: %w", err)
	}

	r := framework.NewOscal2Posture(assessmentResults, catalog, compDef, option.logger)
	data, err := r.Generate()
	if err != nil {
		return err
	}

	out := option.Output
	if out == "-" {
		fmt.Fprintln(os.Stdout, string(data))
	} else {
		return os.WriteFile(out, data, os.ModePerm)
	}
	return nil
}
