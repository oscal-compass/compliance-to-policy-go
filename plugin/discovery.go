/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	manifestPrefix = "c2p-"
	manifestSuffix = "-manifest.json"
)

type findOptions struct {
	providerIds []string
	pluginType  string
}

// FindOption represents a filtering criteria for plugin discovery in plugin.FindPlugins.
type FindOption func(options *findOptions)

// WithProviderIds filters plugins by their provider IDs.
func WithProviderIds(providerIds []string) FindOption {
	return func(options *findOptions) {
		options.providerIds = providerIds
	}
}

// WithPluginType filters available plugins based on the plugin type
// implemented.
func WithPluginType(pluginType string) FindOption {
	return func(options *findOptions) {
		options.pluginType = pluginType
	}
}

// FindPlugins searches for plugins in the specified directory, optionally applying filters.
//
// The function expects plugin manifests in the format "c2p-$PLUGIN-ID-manifest.json".
//
// Available filters:
//   - `WithProviderIds`: Filters by a list of provider IDs.
//   - `WithPluginType`: Filters by plugin type.
//
// If no filters are applied, all discovered plugins are returned.
func FindPlugins(pluginDir string, opts ...FindOption) (ManifestSet, error) {
	config := &findOptions{}
	for _, opt := range opts {
		opt(config)
	}

	matchingPlugins, err := findAllPluginMatches(pluginDir)
	if err != nil {
		return nil, err
	}

	if len(matchingPlugins) == 0 {
		return nil, fmt.Errorf("%w in %s", ErrPluginsNotFound, pluginDir)
	}

	metaDataSet := make(ManifestSet)
	var errs []error

	fmt.Println(matchingPlugins)

	// Filter plugins by provider IDs if provided
	if len(config.providerIds) != 0 {
		filteredIds := make(map[string]string)
		for _, providerId := range config.providerIds {
			if _, ok := matchingPlugins[providerId]; !ok {
				errs = append(errs, &NotFoundError{providerId})
			}
			filteredIds[providerId] = matchingPlugins[providerId]
		}
		matchingPlugins = filteredIds

		// Return early if there are errors to avoid unnecessary processing
		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
	}

	// Process remaining plugins, filtering by plugin type if necessary
	for id, manifestName := range matchingPlugins {
		manifestPath := filepath.Join(pluginDir, manifestName)
		manifest, err := readManifestFile(id, manifestPath)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if config.pluginType != "" && !manifestMatchesType(manifest, config.pluginType) {
			continue
		}

		// sanitize the executable path in the manifest
		cleanPath, err := sanitizeAndResolvePath(pluginDir, manifest.ExecutablePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		manifest.ExecutablePath = cleanPath
		metaDataSet[id] = manifest
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	if len(metaDataSet) == 0 {
		return nil, fmt.Errorf("%w in %s with matching criteria", ErrPluginsNotFound, pluginDir)
	}

	return metaDataSet, nil
}

// findAllPluginsMatches locates the manifests in the plugin directory that match
// the prefix name scheme and return the plugin ID and file name.
func findAllPluginMatches(pluginDir string) (map[string]string, error) {
	items, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, err
	}

	matchingPlugins := make(map[string]string)
	for _, item := range items {
		name := item.Name()
		if !strings.HasPrefix(name, manifestPrefix) {
			continue
		}
		trimmedName := strings.TrimPrefix(name, manifestPrefix)
		trimmedName = strings.TrimSuffix(trimmedName, manifestSuffix)
		matchingPlugins[trimmedName] = name
	}
	return matchingPlugins, nil
}

// manifestMatchesType checks if the plugin manifest defines
// the plugin type being searched for.
func manifestMatchesType(manifest Manifest, pluginType string) bool {
	for _, typ := range manifest.Types {
		if typ == pluginType {
			return true
		}
	}
	return false
}

// readManifestFile reads and parses the manifest from JSON.
func readManifestFile(pluginName, manifestPath string) (Manifest, error) {
	manifestFile, err := os.Open(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Manifest{}, &ManifestNotFoundError{File: manifestPath, PluginID: pluginName}
		}
		return Manifest{}, err
	}
	defer manifestFile.Close()

	jsonParser := json.NewDecoder(manifestFile)
	var manifest Manifest
	err = jsonParser.Decode(&manifest)
	if err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

// sanitizeAndResolvePath returns the absolute, cleaned filepath.
func sanitizeAndResolvePath(pluginDir, path string) (string, error) {
	absPluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		return "", err
	}

	cleanPath := filepath.Clean(filepath.Join(absPluginDir, path))

	// Ensure the path exists and is executable
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		return "", err
	}

	if fileInfo.Mode()&0100 == 0 {
		return "", fmt.Errorf("path %s is not executable", cleanPath)
	}

	return cleanPath, nil
}
