// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"testing"

	repo_model "forgejo.org/models/repo"

	"github.com/stretchr/testify/assert"
)

func TestContextIntegrationGitRepoVocabulary(t *testing.T) {
	data := make(map[string]any)
	repo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeGit}
	InjectContextData(data, repo)

	// Git repos use "Branch" vocabulary
	assert.Equal(t, false, data["IsJJRepo"])
	assert.Equal(t, "Branch", data["BranchLabel"])
	assert.Equal(t, "Branches", data["BranchesLabel"])
	assert.Equal(t, "Branch", data["RefLabel"])

	// Git repos do not show jj-specific features
	assert.Equal(t, false, data["ShowChangeID"])
	assert.Equal(t, false, data["ShowConflicts"])
	assert.Equal(t, false, data["ShowOperationLog"])
}

func TestContextIntegrationJJRepoVocabulary(t *testing.T) {
	data := make(map[string]any)
	repo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeJJ}
	InjectContextData(data, repo)

	// JJ repos use "Bookmark" vocabulary
	assert.Equal(t, true, data["IsJJRepo"])
	assert.Equal(t, "Bookmark", data["BranchLabel"])
	assert.Equal(t, "Bookmarks", data["BranchesLabel"])
	assert.Equal(t, "Bookmark", data["RefLabel"])

	// JJ repos show jj-specific features
	assert.Equal(t, true, data["ShowChangeID"])
	assert.Equal(t, true, data["ShowConflicts"])
	assert.Equal(t, true, data["ShowOperationLog"])
}

func TestContextIntegrationDefaultRepoTreatedAsGit(t *testing.T) {
	data := make(map[string]any)
	repo := &repo_model.Repository{} // empty VCSBackendType
	InjectContextData(data, repo)

	assert.Equal(t, false, data["IsJJRepo"])
	assert.Equal(t, "Branch", data["BranchLabel"])
	assert.Equal(t, "Branches", data["BranchesLabel"])
	assert.Equal(t, false, data["ShowChangeID"])
	assert.Equal(t, false, data["ShowConflicts"])
	assert.Equal(t, false, data["ShowOperationLog"])
}

func TestContextIntegrationAllExpectedKeysPresent(t *testing.T) {
	expectedKeys := []string{
		"IsJJRepo",
		"BranchLabel",
		"BranchesLabel",
		"RefLabel",
		"ShowChangeID",
		"ShowConflicts",
		"ShowOperationLog",
	}

	tests := []struct {
		name    string
		vcsType repo_model.VCSType
	}{
		{"git", repo_model.VCSTypeGit},
		{"jj", repo_model.VCSTypeJJ},
		{"default", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make(map[string]any)
			repo := &repo_model.Repository{VCSBackendType: tt.vcsType}
			InjectContextData(data, repo)

			for _, key := range expectedKeys {
				_, exists := data[key]
				assert.True(t, exists, "expected key %q in context data for %s repo", key, tt.name)
			}
		})
	}
}

func TestContextIntegrationNoExtraKeysInjected(t *testing.T) {
	allowedKeys := map[string]bool{
		"IsJJRepo":         true,
		"BranchLabel":      true,
		"BranchesLabel":    true,
		"RefLabel":         true,
		"ShowChangeID":     true,
		"ShowConflicts":    true,
		"ShowOperationLog": true,
	}

	tests := []struct {
		name    string
		vcsType repo_model.VCSType
	}{
		{"git", repo_model.VCSTypeGit},
		{"jj", repo_model.VCSTypeJJ},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make(map[string]any)
			repo := &repo_model.Repository{VCSBackendType: tt.vcsType}
			InjectContextData(data, repo)

			for key := range data {
				assert.True(t, allowedKeys[key],
					"unexpected key %q in context data for %s repo", key, tt.name)
			}
		})
	}
}

func TestContextIntegrationBooleanTypesCorrect(t *testing.T) {
	// Verify that boolean values are actual bools, not strings or ints
	data := make(map[string]any)
	repo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeJJ}
	InjectContextData(data, repo)

	boolKeys := []string{"IsJJRepo", "ShowChangeID", "ShowConflicts", "ShowOperationLog"}
	for _, key := range boolKeys {
		_, ok := data[key].(bool)
		assert.True(t, ok, "%q should be a bool, got %T", key, data[key])
	}

	stringKeys := []string{"BranchLabel", "BranchesLabel", "RefLabel"}
	for _, key := range stringKeys {
		_, ok := data[key].(string)
		assert.True(t, ok, "%q should be a string, got %T", key, data[key])
	}
}

func TestContextIntegrationGetContextDataConsistentWithInject(t *testing.T) {
	// Verify that GetContextData and InjectContextData produce consistent results
	tests := []struct {
		name    string
		vcsType repo_model.VCSType
	}{
		{"git", repo_model.VCSTypeGit},
		{"jj", repo_model.VCSTypeJJ},
		{"default", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repo_model.Repository{VCSBackendType: tt.vcsType}

			cd := GetContextData(repo)
			data := make(map[string]any)
			InjectContextData(data, repo)

			assert.Equal(t, cd.IsJJRepo, data["IsJJRepo"])
			assert.Equal(t, cd.BranchLabel, data["BranchLabel"])
			assert.Equal(t, cd.BranchesLabel, data["BranchesLabel"])
			assert.Equal(t, cd.RefLabel, data["RefLabel"])
			assert.Equal(t, cd.ShowChangeID, data["ShowChangeID"])
			assert.Equal(t, cd.ShowConflicts, data["ShowConflicts"])
			assert.Equal(t, cd.ShowOperationLog, data["ShowOperationLog"])
		})
	}
}

func TestContextIntegrationVocabularyDiffersBetweenGitAndJJ(t *testing.T) {
	gitRepo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeGit}
	jjRepo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeJJ}

	gitCtx := GetContextData(gitRepo)
	jjCtx := GetContextData(jjRepo)

	// These must differ between git and jj
	assert.NotEqual(t, gitCtx.BranchLabel, jjCtx.BranchLabel,
		"BranchLabel should differ: git=%q jj=%q", gitCtx.BranchLabel, jjCtx.BranchLabel)
	assert.NotEqual(t, gitCtx.BranchesLabel, jjCtx.BranchesLabel,
		"BranchesLabel should differ: git=%q jj=%q", gitCtx.BranchesLabel, jjCtx.BranchesLabel)
	assert.NotEqual(t, gitCtx.IsJJRepo, jjCtx.IsJJRepo,
		"IsJJRepo should differ")
	assert.NotEqual(t, gitCtx.ShowChangeID, jjCtx.ShowChangeID,
		"ShowChangeID should differ")
	assert.NotEqual(t, gitCtx.ShowConflicts, jjCtx.ShowConflicts,
		"ShowConflicts should differ")
	assert.NotEqual(t, gitCtx.ShowOperationLog, jjCtx.ShowOperationLog,
		"ShowOperationLog should differ")
}

func TestContextIntegrationInjectDoesNotClobberExistingData(t *testing.T) {
	data := map[string]any{
		"Title":     "My Repository",
		"PageCount": 42,
		"Owner":     "alice",
	}

	repo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeJJ}
	InjectContextData(data, repo)

	// Existing keys should still be present
	assert.Equal(t, "My Repository", data["Title"])
	assert.Equal(t, 42, data["PageCount"])
	assert.Equal(t, "alice", data["Owner"])

	// VCS keys should also be present
	assert.Equal(t, true, data["IsJJRepo"])
	assert.Equal(t, "Bookmark", data["BranchLabel"])
}
