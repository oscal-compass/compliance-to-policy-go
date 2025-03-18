/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
)

var logger hclog.Logger

func init() {
	logger = defaultLogger()
	logWriter := logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true})
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(logWriter)
}

func defaultLogger() hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Output: os.Stdout,
		Level:  hclog.Info,
	})
}

// GetLogger returns a named hcl.Logger.
func GetLogger(name string) hclog.Logger {
	return logger.Named(name)
}

// NewPluginLogger returns a configured hcl.Logger for plugins to use and
// pass in the plugin.ServeConfig.
func NewPluginLogger() hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Level:           hclog.Debug,
		Output:          os.Stderr,
		SyncParentLevel: true,
		JSONFormat:      true,
	})
}
