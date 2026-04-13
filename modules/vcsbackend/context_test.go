// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"testing"

	repo_model "forgejo.org/models/repo"

	"github.com/stretchr/testify/assert"
)

func TestGetContextDataGitRepo(t *testing.T) {
	repo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeGit}
	cd := GetContextData(repo)
	assert.False(t, cd.IsJJRepo)
	assert.Equal(t, "Branch", cd.BranchLabel)
	assert.Equal(t, "Branches", cd.BranchesLabel)
	assert.False(t, cd.ShowChangeID)
	assert.False(t, cd.ShowConflicts)
	assert.False(t, cd.ShowOperationLog)
}

func TestGetContextDataJJRepo(t *testing.T) {
	repo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeJJ}
	cd := GetContextData(repo)
	assert.True(t, cd.IsJJRepo)
	assert.Equal(t, "Bookmark", cd.BranchLabel)
	assert.Equal(t, "Bookmarks", cd.BranchesLabel)
	assert.True(t, cd.ShowChangeID)
	assert.True(t, cd.ShowConflicts)
	assert.True(t, cd.ShowOperationLog)
}

func TestGetContextDataDefaultRepo(t *testing.T) {
	repo := &repo_model.Repository{} // empty VCSBackendType
	cd := GetContextData(repo)
	assert.False(t, cd.IsJJRepo)
	assert.Equal(t, "Branch", cd.BranchLabel)
}

func TestInjectContextData(t *testing.T) {
	data := make(map[string]any)
	repo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeJJ}
	InjectContextData(data, repo)

	assert.Equal(t, true, data["IsJJRepo"])
	assert.Equal(t, "Bookmark", data["BranchLabel"])
	assert.Equal(t, "Bookmarks", data["BranchesLabel"])
	assert.Equal(t, true, data["ShowChangeID"])
	assert.Equal(t, true, data["ShowConflicts"])
	assert.Equal(t, true, data["ShowOperationLog"])
}

func TestInjectContextDataGit(t *testing.T) {
	data := make(map[string]any)
	repo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeGit}
	InjectContextData(data, repo)

	assert.Equal(t, false, data["IsJJRepo"])
	assert.Equal(t, "Branch", data["BranchLabel"])
}

func TestGetContextData_JJ_HasCommitLabels(t *testing.T) {
	repo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeJJ}
	cd := GetContextData(repo)

	assert.Equal(t, "Change", cd.CommitLabel)
	assert.Equal(t, "Changes", cd.CommitsLabel)
	assert.Equal(t, "Revision", cd.RevisionLabel)
}

func TestGetContextData_Git_HasCommitLabels(t *testing.T) {
	repo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeGit}
	cd := GetContextData(repo)

	assert.Equal(t, "Commit", cd.CommitLabel)
	assert.Equal(t, "Commits", cd.CommitsLabel)
	assert.Equal(t, "Ref", cd.RevisionLabel)
}

func TestInjectContextData_CommitLabels(t *testing.T) {
	dataJJ := make(map[string]any)
	InjectContextData(dataJJ, &repo_model.Repository{VCSBackendType: repo_model.VCSTypeJJ})
	assert.Equal(t, "Change", dataJJ["CommitLabel"])
	assert.Equal(t, "Changes", dataJJ["CommitsLabel"])
	assert.Equal(t, "Revision", dataJJ["RevisionLabel"])

	dataGit := make(map[string]any)
	InjectContextData(dataGit, &repo_model.Repository{VCSBackendType: repo_model.VCSTypeGit})
	assert.Equal(t, "Commit", dataGit["CommitLabel"])
	assert.Equal(t, "Commits", dataGit["CommitsLabel"])
	assert.Equal(t, "Ref", dataGit["RevisionLabel"])
}
