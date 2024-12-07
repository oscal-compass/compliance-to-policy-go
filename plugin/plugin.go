/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"github.com/hashicorp/go-plugin"
)

const (
	// PVPPluginName is used to dispense policy validation point plugin type
	PVPPluginName = "pvp"
	// The ProtocolVersion is the version that must match between the core
	// and plugins.
	ProtocolVersion = 1
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion: ProtocolVersion,

	// These magic cookie values should only be set one time.
	// Please do NOT change.
	MagicCookieKey:   "C2P_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "4fc73041107cf346f76f14d178c3ce63ebb7f6d09d7e2e3983a5737e149e3bfb",
}

// SupportedPlugins is the map of plugins we can dispense.
var SupportedPlugins = map[string]plugin.Plugin{}
