/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package subcommands

import (
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/spf13/cobra"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
)

func NewOSCAL2Posture(logger hclog.Logger) *cobra.Command {
	options := NewOptions()
	options.logger = logger

	command := &cobra.Command{
		Use:   "oscal2posture",
		Short: "Generate Compliance Posture from OSCAL artifacts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Complete(cmd); err != nil {
				return err
			}
			if err := validateOSCAL2Posture(options); err != nil {
				return err
			}
			return runOSCAL2Posture(options)
		},
	}
	fs := command.Flags()
	BindCommonFlags(fs)
	fs.String(Catalog, "", "path to catalog.json")
	fs.StringP("assessment-results", "a", "./assessment-results.json", "path to assessment-results.json")
	fs.StringP("out", "o", "-", "path to output file. Use '-' for stdout. Default '-'.")
	return command
}

// validateOSCAL2Posture runs validation specific to the OSCAL2Posture command.
func validateOSCAL2Posture(options *Options) error {
	var errs []error
	if options.Catalog == "" {
		errs = append(errs, &ConfigError{Option: Catalog})
	}
	if options.Definition == "" {
		errs = append(errs, &ConfigError{Option: ComponentDefinition})
	}
	return errors.Join(errs...)
}

func runOSCAL2Posture(option *Options) error {
	schemaValidator := validation.NewSchemaValidator()
	arFile, err := os.Open(option.AssessmentResults)
	if err != nil {
		return err
	}
	defer arFile.Close()
	assessmentResults, err := models.NewAssessmentResults(arFile, schemaValidator)
	if err != nil {
		return fmt.Errorf("error loading assessment results: %w", err)
	}

	catalogFile, err := os.Open(option.Catalog)
	if err != nil {
		return err
	}
	defer catalogFile.Close()
	catalog, err := models.NewCatalog(catalogFile, schemaValidator)
	if err != nil {
		return fmt.Errorf("error loading catalog: %w", err)
	}

	compDefFile, err := os.Open(option.Definition)
	if err != nil {
		return err
	}
	defer compDefFile.Close()
	compDef, err := models.NewComponentDefinition(compDefFile, schemaValidator)
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
