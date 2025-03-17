/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	PoliciesDir      string `mapstructure:"policy-dir"`
	PolicyResultsDir string `mapstructure:"policy-results-dir"`
	TempDir          string `mapstructure:"temp-dir"`
	OutputDir        string `mapstructure:"output-dir"`
}

func (c Config) Validate() error {
	var errs []error
	if err := checkPath(&c.PoliciesDir); err != nil {
		errs = append(errs, err)
	}
	if err := checkPath(&c.PolicyResultsDir); err != nil {
		errs = append(errs, err)
	}
	if err := checkPath(&c.TempDir); err != nil {
		errs = append(errs, err)
	}
	if err := checkPath(&c.OutputDir); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func checkPath(path *string) error {
	if path != nil && *path != "" {
		cleanedPath := filepath.Clean(*path)
		path = &cleanedPath
		_, err := os.Stat(*path)
		if err != nil {
			return fmt.Errorf("path %q: %w", *path, err)
		}
	}
	return nil
}
