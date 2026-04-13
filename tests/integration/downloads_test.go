// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"forgejo.org/modules/setting"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadsPageRequiresAuth(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/downloads")
	MakeRequest(t, req, http.StatusSeeOther) // Redirects to login
}

func TestDownloadsPageEmpty(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user1")
	resp := session.MakeRequest(t, NewRequest(t, "GET", "/downloads"), http.StatusOK)

	htmlDoc := NewHTMLParser(t, resp.Body)
	htmlDoc.AssertElement(t, ".ui.placeholder.segment", true)
}

func TestDownloadsPageWithBinaries(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// Set up test binaries
	downloadsDir := filepath.Join(setting.AppDataPath, "downloads", "jj", "linux-x86_64")
	require.NoError(t, os.MkdirAll(downloadsDir, 0o755))
	defer os.RemoveAll(filepath.Join(setting.AppDataPath, "downloads"))

	require.NoError(t, os.WriteFile(
		filepath.Join(setting.AppDataPath, "downloads", "jj", "VERSION"),
		[]byte("v0.1.0 (abc1234)"), 0o644))
	require.NoError(t, os.WriteFile(
		filepath.Join(downloadsDir, "jj"),
		[]byte("fake-binary"), 0o755))

	session := loginUser(t, "user1")
	resp := session.MakeRequest(t, NewRequest(t, "GET", "/downloads"), http.StatusOK)

	htmlDoc := NewHTMLParser(t, resp.Body)
	// Should have a table with the binary
	htmlDoc.AssertElement(t, "table.ui.celled.table", true)
	// Should show version
	body := resp.Body.String()
	assert.Contains(t, body, "v0.1.0 (abc1234)")
}

func TestDownloadBinaryRequiresAuth(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/downloads/jj/linux-x86_64/jj")
	MakeRequest(t, req, http.StatusSeeOther)
}

func TestDownloadBinaryInvalidPath(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user1")

	// Invalid binary name
	session.MakeRequest(t, NewRequest(t, "GET", "/downloads/nope/linux-x86_64/nope"), http.StatusNotFound)

	// Invalid platform
	session.MakeRequest(t, NewRequest(t, "GET", "/downloads/jj/freebsd-x86_64/jj"), http.StatusNotFound)

	// Mismatched filename
	session.MakeRequest(t, NewRequest(t, "GET", "/downloads/jj/linux-x86_64/jjf"), http.StatusNotFound)
}

func TestDownloadBinaryServeFile(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// Set up test binary
	downloadsDir := filepath.Join(setting.AppDataPath, "downloads", "jj", "linux-x86_64")
	require.NoError(t, os.MkdirAll(downloadsDir, 0o755))
	defer os.RemoveAll(filepath.Join(setting.AppDataPath, "downloads"))

	binaryContent := []byte("fake-jj-binary-content")
	require.NoError(t, os.WriteFile(filepath.Join(downloadsDir, "jj"), binaryContent, 0o755))

	session := loginUser(t, "user1")
	resp := session.MakeRequest(t, NewRequest(t, "GET", "/downloads/jj/linux-x86_64/jj"), http.StatusOK)

	assert.Equal(t, "application/octet-stream", resp.Header().Get("Content-Type"))
	assert.Contains(t, resp.Header().Get("Content-Disposition"), `filename="jj"`)
	assert.Equal(t, binaryContent, resp.Body.Bytes())
}
