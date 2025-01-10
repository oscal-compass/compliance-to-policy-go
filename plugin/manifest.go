/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manifest is metadata about a plugin to support discovering and
// launching plugins. This should be provided with the plugin on-disk.
type Manifest struct {
	// Metadata has required information for plugin launch and discovery.
	Metadata `json:"metadata"`
	// ExecutablePath is the path to the plugin binary.
	ExecutablePath string `json:"executablePath"`
	// Checksum is the SHA256 hash of the content.
	// This checked against the calculated value at plugin launch.
	Checksum string `json:"sha256"`
}

// ResolvePath validates and sanitizes the Manifest.ExecutablePath.
//
// If the path is not absolute, it updates Manifest.ExecutablePath field
// to a location under the given plugin directory. If the path is absolute,
// it validates it is under the given plugin directory.
func (m *Manifest) ResolvePath(pluginDir string) error {
	absPluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute plugin directory: %w", err)
	}

	// Handle different path types (relative and absolute)
	var cleanedPath string
	if filepath.IsAbs(m.ExecutablePath) {
		if !strings.HasPrefix(m.ExecutablePath, absPluginDir+string(os.PathSeparator)) {
			return fmt.Errorf("absolute path %s is not under the plugin directory %s", m.ExecutablePath, absPluginDir)
		}
		cleanedPath = filepath.Clean(m.ExecutablePath)
	} else {
		cleanedPath = filepath.Clean(filepath.Join(absPluginDir, m.ExecutablePath))
	}

	fileInfo, err := os.Stat(cleanedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("plugin executable %s does not exist: %w", cleanedPath, err)
		}
		return fmt.Errorf("failed to stat plugin executable: %w", err)
	}

	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("plugin executable %s is not a file", cleanedPath)
	}

	if fileInfo.Mode()&0100 == 0 {
		return fmt.Errorf("plugin file %s is not executable", cleanedPath)
	}
	m.ExecutablePath = cleanedPath

	return nil
}

// Metadata has required information for plugin launch and discovery.
type Metadata struct {
	// ID is the name of the plugin. This is the information used
	// when a plugin is requested.
	ID string `json:"id"`
	// Description is a short description for the plugin.
	Description string `json:"description"`
	// Version is the semantic version of the
	// plugin.
	Version string `json:"version"`
	// Type defined which supported plugin types
	// are implemented by this plugin. It should match
	// on or more of the values in plugin.SupportedPlugin.
	Types []string `json:"types"`
}

// ValidateID ensure the plugin id is valid based on the
// plugin IdentifierPattern.
func (m Metadata) ValidateID() bool {
	return IdentifierPattern.MatchString(m.ID)
}

// Manifests defines the Manifest by plugin id.
type Manifests map[string]Manifest
