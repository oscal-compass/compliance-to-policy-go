/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestManifest_ResolvePath(t *testing.T) {
	tmpDir := t.TempDir()
	copyPlugin(t, tmpDir, "testdata/plugins/testplugin")

	tests := []struct {
		name         string
		testManifest Manifest
		wantError    string
		wantPath     string
	}{
		{
			name: "Valid/RelativePathLocation",
			testManifest: Manifest{
				ExecutablePath: "testplugin",
			},
			wantPath: fmt.Sprintf("%s/testplugin", tmpDir),
		},
		{
			name: "Valid/AbsolutePathLocation",
			testManifest: Manifest{
				ExecutablePath: fmt.Sprintf("%s/testplugin", tmpDir),
			},
			wantPath: fmt.Sprintf("%s/testplugin", tmpDir),
		},
		{
			name: "Invalid/PluginNotInExpectedDir",
			testManifest: Manifest{
				ExecutablePath: "/dir/testplugin",
			},
			wantError: fmt.Sprintf("absolute path /dir/testplugin is not under the plugin directory %s", tmpDir),
		},
		{
			name: "Invalid/PluginDoesNotExist",
			testManifest: Manifest{
				ExecutablePath: "notatestplugin",
			},
			wantError: fmt.Sprintf(`plugin executable %s/notatestplugin`+
				` does not exist: stat %s/notatestplugin: no such file or directory`, tmpDir, tmpDir),
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			err := c.testManifest.ResolvePath(tmpDir)
			if c.wantError != "" {
				require.EqualError(t, err, c.wantError)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.wantPath, c.testManifest.ExecutablePath)
			}
		})
	}
}

func TestManifest_ResolveOptions(t *testing.T) {

	defaultValue := "default"
	tests := []struct {
		name         string
		testManifest Manifest
		selections   map[string]string
		wantError    string
		wantOptions  map[string]string
	}{
		{
			name: "Success/AllDefaults",
			testManifest: Manifest{
				ExecutablePath: "testplugin",
				Configuration: []ConfigurationOption{
					{
						Name:        "default",
						Description: "A not required option",
						Required:    false,
						Default:     &defaultValue,
					},
				},
			},
			wantOptions: map[string]string{"default": "default"},
		},
		{
			name: "Success/WithSelections",
			testManifest: Manifest{
				ExecutablePath: "testplugin",
				Configuration: []ConfigurationOption{
					{
						Name:        "required",
						Description: "A required option",
						Required:    true,
					},
					{
						Name:        "default",
						Description: "A default option",
						Required:    false,
						Default:     &defaultValue,
					},
					{
						Name:        "default2",
						Description: "A default option",
						Required:    false,
						Default:     &defaultValue,
					},
				},
			},
			selections: map[string]string{
				"required": "myvalue",
				"default":  "override",
			},
			wantOptions: map[string]string{
				"required": "myvalue",
				"default":  "override",
				"default2": "default",
			},
		},
		{
			name: "Success/RequiredWithDefault",
			testManifest: Manifest{
				ExecutablePath: "testplugin",
				Configuration: []ConfigurationOption{
					{
						Name:        "required_with_default",
						Description: "A required option with a default that needs to be explicitly chosen.",
						Required:    true,
						Default:     &defaultValue,
					},
				},
			},
			selections: map[string]string{
				"required_with_default": "default",
			},
			wantOptions: map[string]string{
				"required_with_default": "default",
			},
		},
		{
			name: "Success/NoConfiguration",
			testManifest: Manifest{
				ExecutablePath: "testplugin",
			},
			wantOptions: map[string]string{},
		},
		{
			name: "Failure/RequiredMissing",
			testManifest: Manifest{
				ExecutablePath: "testplugin",
				Configuration: []ConfigurationOption{
					{
						Name:        "required",
						Description: "A required option",
						Required:    true,
					},
				},
			},
			wantError: "required value not supplied for option \"required\"",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			gotOptions, err := c.testManifest.ResolveOptions(c.selections, hclog.NewNullLogger())
			if c.wantError != "" {
				require.EqualError(t, err, c.wantError)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.wantOptions, gotOptions)
			}
		})
	}
}

func copyPlugin(t *testing.T, tmpDir, srcFile string) {
	dstFile := filepath.Join(tmpDir, filepath.Base(srcFile))

	source, err := os.Open(srcFile)
	require.NoError(t, err)
	defer source.Close()

	destination, err := os.Create(dstFile)
	require.NoError(t, err)
	defer destination.Close()

	_, err = io.Copy(destination, source)
	require.NoError(t, err)

	// Retain the permissions
	srcFileInfo, err := os.Stat(srcFile)
	require.NoError(t, err)
	srcMode := srcFileInfo.Mode()

	err = os.Chmod(dstFile, srcMode)
	require.NoError(t, err)
}
