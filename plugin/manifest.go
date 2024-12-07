/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

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

// ManifestSet defines the Manifest by plugin id.
type ManifestSet map[string]Manifest
