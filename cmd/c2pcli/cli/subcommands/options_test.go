/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package subcommands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		name      string
		options   *Options
		wantError string
	}{
		{
			name: "Invalid/BothOptionsSet",
			options: &Options{
				Definition: "set",
				Plan:       "also-set",
			},
			wantError: "cannot set both component-definition and assessment-plan values",
		},
		{
			name:      "Invalid/NoOptionsSet",
			options:   &Options{},
			wantError: "must set component-definition or assessment-plan",
		},
		{
			name: "Invalid/InvalidOptionsSet",
			options: &Options{
				Definition: "set",
			},
			wantError: "\"name\" option is not set",
		},
		{
			name: "Valid/PlanSet",
			options: &Options{
				Plan: "also-set",
			},
		},
		{
			name: "Valid/DefinitionSet",
			options: &Options{
				Definition: "set",
				Name:       "set",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.options.Validate()

			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}

}
