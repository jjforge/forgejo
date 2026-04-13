// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	repo_model "forgejo.org/models/repo"
)

// ContextData holds jj-specific template context data that handlers
// inject into ctx.Data for conditional rendering in templates.
type ContextData struct {
	// IsJJRepo is true if this repository uses the jj backend.
	IsJJRepo bool

	// BranchLabel is "Bookmark" for jj repos, "Branch" for git repos.
	BranchLabel string

	// BranchesLabel is "Bookmarks" for jj repos, "Branches" for git repos.
	BranchesLabel string

	// RefLabel is how to label refs: "Bookmark" vs "Branch".
	RefLabel string

	// CommitLabel is "Change" for jj repos, "Commit" for git repos.
	CommitLabel string

	// CommitsLabel is "Changes" for jj repos, "Commits" for git repos.
	CommitsLabel string

	// RevisionLabel is "Revision" for jj repos, "Ref" for git repos.
	RevisionLabel string

	// ShowChangeID is true for jj repos (show change_id column in commits).
	ShowChangeID bool

	// ShowConflicts is true for jj repos (show conflict indicators).
	ShowConflicts bool

	// ShowOperationLog is true for jj repos (show operation log tab).
	ShowOperationLog bool
}

// GetContextData returns the template context data for a repository,
// setting vocabulary and feature flags based on the repo's VCS type.
func GetContextData(repo *repo_model.Repository) ContextData {
	if repo.IsJJ() {
		return ContextData{
			IsJJRepo:         true,
			BranchLabel:      "Bookmark",
			BranchesLabel:    "Bookmarks",
			RefLabel:         "Bookmark",
			CommitLabel:      "Change",
			CommitsLabel:     "Changes",
			RevisionLabel:    "Revision",
			ShowChangeID:     true,
			ShowConflicts:    true,
			ShowOperationLog: true,
		}
	}
	return ContextData{
		IsJJRepo:         false,
		BranchLabel:      "Branch",
		BranchesLabel:    "Branches",
		RefLabel:         "Branch",
		CommitLabel:      "Commit",
		CommitsLabel:     "Commits",
		RevisionLabel:    "Ref",
		ShowChangeID:     false,
		ShowConflicts:    false,
		ShowOperationLog: false,
	}
}

// InjectContextData sets the VCS-specific context data into a template data map.
// Called by repo handlers to make vocabulary and feature flags available to templates.
func InjectContextData(data map[string]any, repo *repo_model.Repository) {
	cd := GetContextData(repo)
	data["IsJJRepo"] = cd.IsJJRepo
	data["BranchLabel"] = cd.BranchLabel
	data["BranchesLabel"] = cd.BranchesLabel
	data["RefLabel"] = cd.RefLabel
	data["CommitLabel"] = cd.CommitLabel
	data["CommitsLabel"] = cd.CommitsLabel
	data["RevisionLabel"] = cd.RevisionLabel
	data["ShowChangeID"] = cd.ShowChangeID
	data["ShowConflicts"] = cd.ShowConflicts
	data["ShowOperationLog"] = cd.ShowOperationLog
}
