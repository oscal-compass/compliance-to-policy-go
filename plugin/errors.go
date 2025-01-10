package plugin

import (
	"errors"
	"fmt"
)

// ErrPluginsNotFound should be used when there are no discoverable
// in the defined location.
var ErrPluginsNotFound = errors.New("no plugins found")

// NotFoundError indicates that a requested plugin if not found
// in the list of discovered plugins.
type NotFoundError struct {
	PluginID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("failed to find plugin %q in plugin installation location", e.PluginID)
}

// ManifestNotFoundError indicated that a plugin manifest
// cannot be located for a discovered plugin.
type ManifestNotFoundError struct {
	PluginID string
	File     string
}

func (e *ManifestNotFoundError) Error() string {
	return fmt.Sprintf("failed to open manifest file %s for plugin %q", e.File, e.PluginID)
}
