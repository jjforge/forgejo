// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package web

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDownloadsDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create jj binary with VERSION
	jjDir := filepath.Join(dir, "jj")
	require.NoError(t, os.MkdirAll(filepath.Join(jjDir, "linux-x86_64"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(jjDir, "darwin-arm64"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(jjDir, "VERSION"), []byte("v0.1.0 (abc1234)"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(jjDir, "linux-x86_64", "jj"), []byte("fake-binary-content"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(jjDir, "darwin-arm64", "jj"), []byte("fake-binary"), 0o755))

	// Create jjf binary with VERSION, only one platform
	jjfDir := filepath.Join(dir, "jjf")
	require.NoError(t, os.MkdirAll(filepath.Join(jjfDir, "linux-x86_64"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(jjfDir, "VERSION"), []byte("v0.2.0 (def5678)"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(jjfDir, "linux-x86_64", "jjf"), []byte("fake-jjf-binary"), 0o755))

	return dir
}

func TestScanDownloadsDir(t *testing.T) {
	dir := setupTestDownloadsDir(t)

	binaries := scanDownloadsDir(dir)

	require.Len(t, binaries, 2)

	// jj should be first (alphabetical)
	assert.Equal(t, "jj", binaries[0].Name)
	assert.Equal(t, "v0.1.0 (abc1234)", binaries[0].Version)
	require.Len(t, binaries[0].Platforms, 2)

	// Check darwin-arm64 platform (alphabetical first)
	assert.Equal(t, "macOS", binaries[0].Platforms[0].OS)
	assert.Equal(t, "arm64", binaries[0].Platforms[0].Arch)
	assert.Equal(t, "jj", binaries[0].Platforms[0].Filename)
	assert.Equal(t, "/downloads/jj/darwin-arm64/jj", binaries[0].Platforms[0].URL)

	// Check linux-x86_64 platform
	assert.Equal(t, "Linux", binaries[0].Platforms[1].OS)
	assert.Equal(t, "x86_64", binaries[0].Platforms[1].Arch)

	// jjf
	assert.Equal(t, "jjf", binaries[1].Name)
	assert.Equal(t, "v0.2.0 (def5678)", binaries[1].Version)
	require.Len(t, binaries[1].Platforms, 1)
}

func TestScanDownloadsDirEmpty(t *testing.T) {
	dir := t.TempDir()
	binaries := scanDownloadsDir(dir)
	assert.Empty(t, binaries)
}

func TestScanDownloadsDirMissing(t *testing.T) {
	binaries := scanDownloadsDir("/nonexistent/path")
	assert.Empty(t, binaries)
}
