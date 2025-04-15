/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import "regexp"

// IdentifierPattern defines criteria the plugin id must comply with.
// It includes the following criteria:
//  1. Consist of lowercase alphanumeric characters
//  2. May contain underscore (_) or hyphen (-) characters.
var IdentifierPattern = regexp.MustCompile("^[a-z0-9_-]+$")

// ID is a unique identifier for a plugin.
type ID string

// String implements the Stringer interface
func (i ID) String() string {
	return string(i)
}

// Validate ensures the plugin id is valid based on the
// plugin IdentifierPattern.
func (i ID) Validate() bool {
	return IdentifierPattern.MatchString(i.String())
}
