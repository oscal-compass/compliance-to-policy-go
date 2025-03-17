/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package subcommands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "none"
	commit  = "none"
	date    = "unknown"
)

func NewVersionSubCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "version",
		Short: "Display version",
		RunE: func(cmd *cobra.Command, args []string) error {
			message := fmt.Sprintf("version: %s, commit: %s, date: %s", version, commit, date)
			fmt.Fprintln(os.Stdout, message)
			return nil
		},
	}
	return command
}
