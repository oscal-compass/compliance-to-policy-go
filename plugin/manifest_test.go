/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetadata_ValidateID(t *testing.T) {
	passingMetadata := Metadata{
		ID: "test-plugin",
	}
	require.True(t, passingMetadata.ValidateID())
	failingMetadata := Metadata{
		ID: "TEST-PLUGIN",
	}
	require.False(t, failingMetadata.ValidateID())
}
