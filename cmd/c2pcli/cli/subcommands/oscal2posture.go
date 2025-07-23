/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package subcommands

import (
	"context"
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
			if err := options.Validate(); err != nil {
				return err
			}
			if err := validateOSCAL2Posture(options); err != nil {
				return err
			}
			return runOSCAL2Posture(cmd.Context(), options)
		},
	}
	fs := command.Flags()
	BindCommonFlags(fs)
	fs.String(Catalog, "", "path to catalog.json")
	fs.StringP("assessment-results", "r", "./assessment-results.json", "path to assessment-results.json")
	fs.StringP("out", "o", "-", "path to output file. Use '-' for stdout. Default '-'.")
	return command
}

// validateOSCAL2Posture runs validation specific to the OSCAL2Posture command.
func validateOSCAL2Posture(options *Options) error {
	var errs []error
	if options.Catalog == "" {
		errs = append(errs, &ConfigError{Option: Catalog})
	}
	return errors.Join(errs...)
}

func runOSCAL2Posture(ctx context.Context, option *Options) error {
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

	plan, _, err := createOrGetPlan(ctx, option)
	if err != nil {
		return err
	}

	r := framework.NewPosture(assessmentResults, catalog, plan, option.logger)
	data, err := r.Generate(option.Output)
	if err != nil {
		return err
	}

	out := option.Output
	if out == "-" {
		fmt.Fprintln(os.Stdout, string(data))
	} else {
		return os.WriteFile(out, data, 0600)
	}
	return nil
}
