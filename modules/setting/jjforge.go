// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"strings"

	"forgejo.org/modules/log"
)

// JJForge holds configuration for jjforge integration with the jj-sidecar.
// These settings are read from the [jjforge] section of app.ini and can be
// overridden via environment variables using the FORGEJO__jjforge__ prefix.
var JJForge = struct {
	// Enabled controls whether jj repository support is active.
	// When false, all repositories are treated as git repos (stock Forgejo behavior).
	Enabled bool `ini:"ENABLED"`

	// SidecarURL is the base URL for the jj-sidecar Browse API.
	// Forgejo calls this URL for jj repository content (tree, blob, commits, refs).
	// In Docker Compose, this is typically http://jj-sidecar:8080.
	SidecarURL string `ini:"SIDECAR_URL"`

	// InternalToken is the shared secret for Forgejo -> sidecar authentication.
	// Sent as a Bearer token in the Authorization header on every sidecar request.
	InternalToken string `ini:"INTERNAL_TOKEN"`
}{
	Enabled:    false,
	SidecarURL: "http://localhost:8080",
}

func loadJJForgeFrom(rootCfg ConfigProvider) {
	if err := rootCfg.Section("jjforge").MapTo(&JJForge); err != nil {
		log.Fatal("Failed to map JJForge settings: %v", err)
	}

	// Trim trailing slash from SidecarURL for consistent URL construction
	JJForge.SidecarURL = strings.TrimRight(JJForge.SidecarURL, "/")
}
