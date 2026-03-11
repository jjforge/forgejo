// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadJJForgeSettingsDefaults(t *testing.T) {
	cfg, _ := NewConfigProviderFromData("")
	loadJJForgeFrom(cfg)

	assert.Equal(t, "http://localhost:8080", JJForge.SidecarURL)
	assert.Equal(t, "", JJForge.InternalToken)
	assert.False(t, JJForge.Enabled)
}

func TestLoadJJForgeSettingsFromIni(t *testing.T) {
	cfg, _ := NewConfigProviderFromData(`
[jjforge]
SIDECAR_URL = http://jj-sidecar:8080
INTERNAL_TOKEN = my-secret-token
ENABLED = true
`)
	loadJJForgeFrom(cfg)

	assert.Equal(t, "http://jj-sidecar:8080", JJForge.SidecarURL)
	assert.Equal(t, "my-secret-token", JJForge.InternalToken)
	assert.True(t, JJForge.Enabled)
}

func TestLoadJJForgeSettingsTrailingSlashTrimmed(t *testing.T) {
	cfg, _ := NewConfigProviderFromData(`
[jjforge]
SIDECAR_URL = http://jj-sidecar:8080/
`)
	loadJJForgeFrom(cfg)

	assert.Equal(t, "http://jj-sidecar:8080", JJForge.SidecarURL)
}
