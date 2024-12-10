/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"crypto/sha256"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// Register a set of implemented plugins.
// This function should be called last during plugin initialization in the main function.
func Register(plugins map[string]plugin.Plugin) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins:         plugins,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

// ClientFactoryFunc defines a function signature for creating
// new go-plugin clients.
type ClientFactoryFunc func(manifest Manifest) (*plugin.Client, error)

// ClientFactory returns a factory function for creating new plugin-specific
// clients with consistent plugin config settings.
//
// The returned factory function takes a Manifest object as input and returns
// a new plugin client configured with the specified logger, allowed protocols,
// and security settings.
func ClientFactory(logger hclog.Logger) ClientFactoryFunc {
	return func(manifest Manifest) (*plugin.Client, error) {
		config := &plugin.ClientConfig{
			HandshakeConfig: Handshake,
			Logger:          logger,
			// Enabling this will ensure that client.Kill() is run when this is cleaned up.
			Managed:          true,
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
			Cmd:              exec.Command(manifest.ExecutablePath),
			Plugins:          SupportedPlugins,
			SecureConfig: &plugin.SecureConfig{
				Checksum: []byte(manifest.Checksum),
				Hash:     sha256.New(),
			},
		}

		client := plugin.NewClient(config)
		return client, nil
	}
}
