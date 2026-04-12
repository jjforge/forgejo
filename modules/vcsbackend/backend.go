// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// Package vcsbackend provides the VCSBackend interface for abstracting
// repository content operations across different version control systems.
// GitBackend wraps modules/git/ for git repos. JjBackend calls the
// jj-sidecar HTTP API for jj repos. Handlers dispatch via GetBackend().
package vcsbackend

// VCSBackend abstracts repository content read and write operations.
// All repository page rendering goes through this interface, allowing
// transparent dispatch between git and jj backends.
type VCSBackend interface {
	// ListTree returns a directory listing at the given ref and path.
	ListTree(ref, path string) (*TreeResponse, error)

	// GetBlob returns the content of a single file at the given ref and path.
	GetBlob(ref, path string) (*BlobResponse, error)

	// GetCommits returns paginated commit history starting from the given ref.
	// If path is non-empty, only commits touching that path are returned.
	GetCommits(ref string, path string, page, perPage int) (*CommitsResponse, error)

	// GetCommit returns detailed information about a single commit by ID.
	GetCommit(id string) (*CommitDetailResponse, error)

	// GetRefs returns all refs (branches/bookmarks, tags, anonymous heads).
	GetRefs() (*RefsResponse, error)

	// GetDefaultRef returns the name of the default ref (e.g., "main").
	GetDefaultRef() (string, error)

	// GetDiff returns the diff for a single commit.
	GetDiff(commitID string) (*DiffResponse, error)

	// CompareDiff returns the diff between two refs.
	CompareDiff(base, head string) (*DiffResponse, error)

	// BlameFile returns blame/annotate information for a file.
	BlameFile(ref, path string) (*BlameResponse, error)

	// Merge merges source commits into the target ref using the given strategy.
	Merge(targetRef string, sourceCommits []string, strategy MergeStrategy) (*MergeResult, error)

	// Rebase rebases commits onto a new base.
	Rebase(commits []string, onto string) (*RebaseResult, error)

	// CreateRef creates a new ref (branch/bookmark) pointing to the given commit.
	CreateRef(name, commitID string) error

	// DeleteRef deletes a ref (branch/bookmark).
	DeleteRef(name string) error

	// GetOperations returns paginated jj operation log entries.
	// Returns an error for git repos (not applicable).
	GetOperations(page, perPage int) (*OperationsResponse, error)
}
