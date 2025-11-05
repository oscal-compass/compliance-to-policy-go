/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package subcommands

import (
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

// NewTools creates a new tools command that groups utility commands
func NewTools(logger hclog.Logger) *cobra.Command {
	command := &cobra.Command{
		Use:   "tools",
		Short: "Utility tools for OSCAL transformations",
		Long:  "Utility tools for transforming and working with OSCAL artifacts",
	}

	command.AddCommand(
		NewCD2AP(logger),
	)

	return command
}
