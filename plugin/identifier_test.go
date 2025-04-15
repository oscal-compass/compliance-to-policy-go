/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestID_Validate(t *testing.T) {
	passingMetadata := Metadata{
		ID: "test-plugin",
	}
	require.True(t, passingMetadata.ID.Validate())
	failingMetadata := Metadata{
		ID: "TEST-PLUGIN",
	}
	require.False(t, failingMetadata.ID.Validate())
}
