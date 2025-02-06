/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package policy

// Provider defines methods for a policy engine C2P plugin.
type Provider interface {
	// Configure send configuration options and selected values to the
	// plugin.
	Configure(map[string]string) error
	// Generate policy artifacts for a specific policy engine.
	Generate(Policy) error
	// GetResults from a specific policy engine and transform into
	// PVPResults.
	GetResults(Policy) (PVPResult, error)
}
