// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"testing"

	repo_model "forgejo.org/models/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitBackend(t *testing.T) {
	repo := &repo_model.Repository{
		OwnerName:     "alice",
		Name:          "myrepo",
		DefaultBranch: "main",
	}
	backend := NewGitBackend(repo)
	require.NotNil(t, backend)
	assert.Equal(t, repo, backend.repo)
}

func TestGitBackendGetDefaultRef(t *testing.T) {
	repo := &repo_model.Repository{
		DefaultBranch: "main",
	}
	backend := NewGitBackend(repo)
	ref, err := backend.GetDefaultRef()
	require.NoError(t, err)
	assert.Equal(t, "main", ref)
}

func TestGitBackendGetDefaultRefEmpty(t *testing.T) {
	repo := &repo_model.Repository{}
	backend := NewGitBackend(repo)
	ref, err := backend.GetDefaultRef()
	require.NoError(t, err)
	assert.Equal(t, "", ref)
}

func TestGitBackendImplementsVCSBackend(t *testing.T) {
	repo := &repo_model.Repository{}
	var _ VCSBackend = NewGitBackend(repo)
}

func TestGitTreeEntryTypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		isDir    bool
		isLink   bool
		expected string
	}{
		{"directory", true, false, "directory"},
		{"symlink", false, true, "symlink"},
		{"file", false, false, "file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entryType := gitEntryType(tt.isDir, tt.isLink)
			assert.Equal(t, tt.expected, entryType)
		})
	}
}
