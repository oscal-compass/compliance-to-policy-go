/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"bytes"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestSetLogger(t *testing.T) {
	var buf bytes.Buffer
	customLogger := hclog.New(&hclog.LoggerOptions{
		Name:   "custom",
		Output: &buf,
		Level:  hclog.Debug,
	})

	SetLogger(customLogger)

	log := GetLogger("test")
	require.NotNil(t, log)

	log.Debug("debug message")
	require.Contains(t, buf.String(), "debug message")
	require.Contains(t, buf.String(), "custom.test")
}

func TestGetLogger_DefaultLogger(t *testing.T) {
	logger = defaultLogger()

	log := GetLogger("component")
	require.NotNil(t, log)
}
