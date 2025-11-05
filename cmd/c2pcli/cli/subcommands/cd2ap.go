/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package subcommands

import (
	"context"
	"fmt"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/transformers"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/spf13/cobra"

	"github.com/oscal-compass/compliance-to-policy-go/v2/internal/utils"
)

func NewCD2AP(logger hclog.Logger) *cobra.Command {
	options := NewOptions()
	options.logger = logger

	command := &cobra.Command{
		Use:   "cd2ap",
		Short: "Create an Assessment Plan from a Component Definition.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Complete(cmd); err != nil {
				return err
			}
			if err := validateCD2AP(options); err != nil {
				return err
			}
			return runCD2AP(cmd.Context(), options)
		},
	}

	fs := command.Flags()
	fs.StringP(ComponentDefinition, "d", "", "path to component-definition.json file")
	fs.StringP(Name, "n", "", "short name of the control source for the implementation to be evaluated")
	fs.StringP("out", "o", "./assessment-plan.json", "path to output OSCAL Assessment Plan")

	return command
}

// validateCD2AP runs validation specific to the CD2AP command.
func validateCD2AP(options *Options) error {
	if options.Definition == "" {
		return &ConfigError{Option: ComponentDefinition}
	}
	if options.Name == "" {
		return &ConfigError{Option: Name}
	}
	return nil
}

func runCD2AP(ctx context.Context, option *Options) error {
	// Load component definition
	compDef, err := loadCompDef(option.Definition)
	if err != nil {
		return fmt.Errorf("error loading component definition: %w", err)
	}

	// Transform component definition to assessment plan
	option.logger.Info("Converting component definition to assessment plan", "framework", option.Name)
	ap, err := transformers.ComponentDefinitionsToAssessmentPlan(ctx, []oscalTypes.ComponentDefinition{compDef}, option.Name)
	if err != nil {
		return fmt.Errorf("error converting component definition to assessment plan: %w", err)
	}

	// Validate the assessment plan
	option.logger.Info("Validating generated assessment plan")
	validator := validation.NewSchemaValidator()
	oscalModels := oscalTypes.OscalModels{
		AssessmentPlan: ap,
	}
	if err := validator.Validate(oscalModels); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Write the assessment plan to file
	option.logger.Info(fmt.Sprintf("Writing assessment plan to %s", option.Output))
	err = utils.WriteObjToJsonFile(option.Output, oscalModels)
	if err != nil {
		return fmt.Errorf("error writing assessment plan to file: %w", err)
	}

	return nil
}
