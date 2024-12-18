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

	"github.com/stretchr/testify/require"
)

func TestMetadata_ValidateID(t *testing.T) {
	passingMetadata := Metadata{
		ID: "test-plugin",
	}
	require.True(t, passingMetadata.ValidateID())
	failingMetadata := Metadata{
		ID: "TEST-PLUGIN",
	}
	require.False(t, failingMetadata.ValidateID())
}

func TestManifest_ResolvePath(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

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
