# Plugins

C2P supports external plugins to allow extension of functionality without needing to recompile.

This package contains shared code needed for building C2P plugins to interact with the C2P Plugin Manager.
The `hashicorp/go-plugin` library is leveraged to support gRPC-based plugins.

## How To Write a Plugin

### Example Code for a C2P plugin in Go
```go
package main

import (
	hplugin "github.com/hashicorp/go-plugin"

	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
)

var _ policy.Provider = (*PluginServer)(nil)

type PluginServer struct {}

func (s *PluginServer) Configure(m map[string]string) error {
	// Configure send configuration options and selected values to the
	// plugin.
	panic("implement me")
}

func (s *PluginServer) Generate(p policy.Policy) error {
	// Generate policy artifacts for a specific policy engine.
	panic("implement me")
}

func (s *PluginServer) GetResults(p policy.Policy) (policy.PVPResult, error) {
	// GetResults from a specific policy engine and transform into
	// PVPResults.
	panic("implement me")
}


func main() {
	myPlugin := &PluginServer{}
	pluginByType := map[string]hplugin.Plugin{
		plugin.PVPPluginName: &plugin.PVPPlugin{Impl: myPlugin},
	}
	plugin.Register(pluginByType)
}
```

### Manifest

The plugin manifest is a JSON file that provides metadata about the plugin. It can optionally include global plugin
configuration options and defaults.

```json
{
  "metadata": {
    "id": "myplugin",
    "description": "My C2P plugin",
    "version": "0.0.1",
    "types": [
      "pvp"
    ]
  },
  "executablePath": "myplugin",
  "sha256": "63784a675a475b0e93865eae1028626a90bab7c66f55c3d8a510f06874e0924a",
  "configuration": [
    {
      "name": "myoption",
      "description": "My plugin option",
      "required": false,
      "default": "defaultvalue"
    }
  ]
}
```
