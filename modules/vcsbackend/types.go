// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import "time"

// TreeEntry represents a single entry in a directory listing.
type TreeEntry struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	Type          string `json:"type"` // "file", "directory", "symlink"
	Size          int64  `json:"size,omitempty"`
	IsExecutable  bool   `json:"is_executable,omitempty"`
	HasConflict   bool   `json:"has_conflict,omitempty"` // jj only
	CommitID      string `json:"commit_id,omitempty"`
	CommitMessage string `json:"commit_message,omitempty"`
}

// TreeResponse is the result of listing a directory tree.
type TreeResponse struct {
	Path         string      `json:"path"`
	Ref          string      `json:"ref"`
	RefName      string      `json:"ref_name,omitempty"`
	Entries      []TreeEntry `json:"entries"`
	TotalEntries int         `json:"total_entries"`
	Readme       string      `json:"readme,omitempty"`
	ReadmePath   string      `json:"readme_path,omitempty"`
}

// BlobResponse is the result of fetching a single file's content.
type BlobResponse struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	IsBinary    bool   `json:"is_binary"`
	IsLarge     bool   `json:"is_large"`
	HasConflict bool   `json:"has_conflict,omitempty"`
	Content     string `json:"content,omitempty"`
	Language    string `json:"language,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
}

// Signature represents a commit author or committer.
type Signature struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

// CommitInfo represents a single commit in a list.
type CommitInfo struct {
	CommitID      string    `json:"commit_id"`
	ChangeID      string    `json:"change_id,omitempty"` // jj only
	ShortID       string    `json:"short_id"`
	ShortChangeID string    `json:"short_change_id,omitempty"` // jj only
	Author        Signature `json:"author"`
	Subject       string    `json:"subject"`
	HasConflict   bool      `json:"has_conflict,omitempty"`    // jj only
	IsWorkingCopy bool      `json:"is_working_copy,omitempty"` // jj only
	Parents       []string  `json:"parents"`
	Bookmarks     []string  `json:"bookmarks,omitempty"` // jj only
}

// CommitsResponse is the result of listing commits.
type CommitsResponse struct {
	Commits []CommitInfo `json:"commits"`
	Total   int          `json:"total"`
	Page    int          `json:"page"`
	PerPage int          `json:"per_page"`
}

// CommitDetailResponse is the detailed view of a single commit.
type CommitDetailResponse struct {
	CommitInfo
	Committer     *Signature `json:"committer,omitempty"`
	Message       string     `json:"message"`
	Stats         *DiffStats `json:"stats,omitempty"`
	Diff          []DiffFile `json:"diff,omitempty"`
	ConflictFiles []string   `json:"conflict_files,omitempty"` // jj only
}

// DiffStats holds aggregate diff statistics.
type DiffStats struct {
	FilesChanged int `json:"files_changed"`
	Insertions   int `json:"insertions"`
	Deletions    int `json:"deletions"`
}

// DiffFile represents a single file in a diff.
type DiffFile struct {
	OldPath string     `json:"old_path,omitempty"`
	NewPath string     `json:"new_path,omitempty"`
	Status  string     `json:"status"` // "added", "modified", "deleted", "renamed"
	Hunks   []DiffHunk `json:"hunks"`
}

// DiffHunk represents a single diff hunk.
type DiffHunk struct {
	Header string     `json:"header"`
	Lines  []DiffLine `json:"lines"`
}

// DiffLine represents a single line in a diff hunk.
type DiffLine struct {
	Type    string `json:"type"` // "add", "delete", "context"
	Content string `json:"content"`
	OldLine int    `json:"old_line,omitempty"`
	NewLine int    `json:"new_line,omitempty"`
}

// DiffResponse is the result of getting a diff.
type DiffResponse struct {
	Stats *DiffStats `json:"stats,omitempty"`
	Files []DiffFile `json:"files"`
}

// RefInfo represents a single ref (branch/bookmark/tag).
type RefInfo struct {
	Name        string `json:"name,omitempty"`
	CommitID    string `json:"commit_id"`
	ChangeID    string `json:"change_id,omitempty"` // jj only
	Type        string `json:"type"`                // "branch", "bookmark", "tag", "anonymous", "working_copy"
	IsCurrent   bool   `json:"is_current,omitempty"`
	Description string `json:"description,omitempty"` // jj anonymous heads
}

// RefsResponse is the result of listing all refs.
type RefsResponse struct {
	Branches []RefInfo `json:"bookmarks,omitempty"` // git branches or jj bookmarks
	Tags     []RefInfo `json:"tags,omitempty"`
	Heads    []RefInfo `json:"heads,omitempty"`          // jj anonymous heads
	OpHead   string    `json:"operation_head,omitempty"` // jj only
}

// BlameEntry represents a single blame line range.
type BlameEntry struct {
	CommitID  string    `json:"commit_id"`
	Author    Signature `json:"author"`
	StartLine int       `json:"start_line"`
	EndLine   int       `json:"end_line"`
}

// BlameResponse is the result of blame/annotate on a file.
type BlameResponse struct {
	Path    string       `json:"path"`
	Entries []BlameEntry `json:"entries"`
}

// OperationInfo represents a single jj operation.
type OperationInfo struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Timestamp   string   `json:"timestamp"`
	Hostname    string   `json:"hostname"`
	Username    string   `json:"username"`
	Parents     []string `json:"parents"`
	Heads       []string `json:"heads"`
}

// OperationsResponse is the result of listing operations.
type OperationsResponse struct {
	Operations []OperationInfo `json:"operations"`
	Total      int             `json:"total,omitempty"`
	Page       int             `json:"page,omitempty"`
	PerPage    int             `json:"per_page,omitempty"`
}

// MergeStrategy defines how merges are performed.
type MergeStrategy string

const (
	MergeStrategyMerge  MergeStrategy = "merge"
	MergeStrategyRebase MergeStrategy = "rebase"
	MergeStrategySquash MergeStrategy = "squash"
)

// MergeResult is the outcome of a merge operation.
type MergeResult struct {
	CommitID    string `json:"commit_id"`
	HasConflict bool   `json:"has_conflict"`
}

// RebaseResult is the outcome of a rebase operation.
type RebaseResult struct {
	NewCommitIDs []string `json:"new_commit_ids"`
	HasConflict  bool     `json:"has_conflict"`
}
