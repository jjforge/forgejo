// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	repo_model "forgejo.org/models/repo"
)

// GetBackend returns the appropriate VCSBackend implementation for a repository.
// For jj repos (VCSBackendType == "jj"), it returns a JjBackend that calls
// the sidecar's Browse API over HTTP. For all other repos (including the
// default empty string), it returns a GitBackend wrapping modules/git/.
func GetBackend(repo *repo_model.Repository) VCSBackend {
	if repo.IsJJ() {
		return NewJjBackend(repo.OwnerName, repo.Name)
	}
	return NewGitBackend(repo)
}
