// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"testing"

	repo_model "forgejo.org/models/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBackendReturnsGitBackendForGitRepo(t *testing.T) {
	repo := &repo_model.Repository{
		OwnerName:      "alice",
		Name:           "myrepo",
		VCSBackendType: repo_model.VCSTypeGit,
	}
	backend := GetBackend(repo)
	require.NotNil(t, backend)
	_, ok := backend.(*GitBackend)
	assert.True(t, ok, "expected GitBackend for git repo")
}

func TestGetBackendReturnsGitBackendForDefaultRepo(t *testing.T) {
	repo := &repo_model.Repository{
		OwnerName: "alice",
		Name:      "myrepo",
		// VCSBackendType is empty string (zero value), should default to git
	}
	backend := GetBackend(repo)
	require.NotNil(t, backend)
	_, ok := backend.(*GitBackend)
	assert.True(t, ok, "expected GitBackend for default repo type")
}

func TestGetBackendReturnsJjBackendForJjRepo(t *testing.T) {
	repo := &repo_model.Repository{
		OwnerName:      "alice",
		Name:           "myrepo",
		VCSBackendType: repo_model.VCSTypeJJ,
	}
	backend := GetBackend(repo)
	require.NotNil(t, backend)
	_, ok := backend.(*JjBackend)
	assert.True(t, ok, "expected JjBackend for jj repo")
}
