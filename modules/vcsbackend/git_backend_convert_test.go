// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"main.go", "go"},
		{"lib.rs", "rust"},
		{"app.py", "python"},
		{"index.js", "javascript"},
		{"app.ts", "typescript"},
		{"Main.java", "java"},
		{"config.yml", "yaml"},
		{"data.json", "json"},
		{"Cargo.toml", "toml"},
		{"README.md", "markdown"},
		{"page.html", "html"},
		{"style.css", "css"},
		{"hello.c", "c"},
		{"hello.cpp", "cpp"},
		{"script.sh", "bash"},
		{"noext", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectLanguage(tt.path))
		})
	}
}

func TestGitEntryType(t *testing.T) {
	assert.Equal(t, "directory", gitEntryType(true, false))
	assert.Equal(t, "symlink", gitEntryType(false, true))
	assert.Equal(t, "file", gitEntryType(false, false))
	// directory takes precedence over link
	assert.Equal(t, "directory", gitEntryType(true, true))
}
