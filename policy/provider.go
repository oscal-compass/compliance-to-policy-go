/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package policy

import "github.com/oscal-compass/oscal-sdk-go/extensions"

/*
Provider defines methods for a policy validation engine.
Defined uses cases include the following:
1. A scanning plugin may contact a remote API for scanning
2. A scanning plugin may exec out to another tool for scanning in a new process
3. A scanning plugin may be a self-contained scanning tool
*/
type Provider interface {
	Generate(rules []extensions.RuleSet) error
	GetResults(rules []extensions.RuleSet) (PVPResult, error)
}
