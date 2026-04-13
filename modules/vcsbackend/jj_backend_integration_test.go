// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"forgejo.org/modules/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSidecar creates an httptest server that mimics all sidecar Browse API endpoints.
// It returns realistic JSON responses for each endpoint based on the request path.
func mockSidecar(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Tree endpoint: /api/jj/:owner/:repo/tree/*
	mux.HandleFunc("/api/jj/alice/myrepo/tree/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		treePath := strings.TrimPrefix(r.URL.Path, "/api/jj/alice/myrepo/tree/")
		ref := r.URL.Query().Get("ref")
		if ref == "" {
			ref = "@"
		}

		var resp map[string]any

		switch treePath {
		case "src":
			resp = map[string]any{
				"path": "src",
				"ref":  "abc123def456",
				"entries": []map[string]any{
					{"name": "main.go", "path": "src/main.go", "type": "file", "size": 1024},
					{"name": "lib", "path": "src/lib", "type": "directory"},
					{"name": "util.go", "path": "src/util.go", "type": "file", "size": 512, "is_executable": false},
				},
				"total_entries": 3,
			}
		case "":
			resp = map[string]any{
				"path": "",
				"ref":  "abc123def456",
				"entries": []map[string]any{
					{"name": "src", "path": "src", "type": "directory"},
					{"name": "README.md", "path": "README.md", "type": "file", "size": 256},
					{"name": "go.mod", "path": "go.mod", "type": "file", "size": 128},
				},
				"total_entries": 3,
				"readme":        "# My Project\n\nThis is a test project.",
				"readme_path":   "README.md",
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "NOT_FOUND",
				"message": fmt.Sprintf("Path not found: %s", treePath),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Blob endpoint: /api/jj/:owner/:repo/blob/*
	mux.HandleFunc("/api/jj/alice/myrepo/blob/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		blobPath := strings.TrimPrefix(r.URL.Path, "/api/jj/alice/myrepo/blob/")

		var resp map[string]any

		switch blobPath {
		case "src/main.go":
			resp = map[string]any{
				"path":      "src/main.go",
				"size":      89,
				"is_binary": false,
				"is_large":  false,
				"content":   "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n",
				"language":  "go",
				"encoding":  "utf-8",
			}
		case "image.png":
			resp = map[string]any{
				"path":      "image.png",
				"size":      4096,
				"is_binary": true,
				"is_large":  false,
				"content":   "",
				"language":  "",
				"encoding":  "",
			}
		case "bigfile.dat":
			resp = map[string]any{
				"path":         "bigfile.dat",
				"size":         10485760,
				"is_binary":    true,
				"is_large":     true,
				"content":      "",
				"language":     "",
				"encoding":     "",
				"download_url": "http://sidecar/api/jj/alice/myrepo/raw/bigfile.dat",
			}
		case "conflicted.txt":
			resp = map[string]any{
				"path":         "conflicted.txt",
				"size":         200,
				"is_binary":    false,
				"is_large":     false,
				"has_conflict": true,
				"content":      "<<<<<<< Side #1\nfoo\n||||||||\nbar\n>>>>>>> Side #2\nbaz\n",
				"language":     "",
				"encoding":     "utf-8",
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "NOT_FOUND",
				"message": fmt.Sprintf("File not found: %s", blobPath),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Changes endpoint: /api/jj/:owner/:repo/changes
	mux.HandleFunc("/api/jj/alice/myrepo/changes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		page := r.URL.Query().Get("page")
		perPage := r.URL.Query().Get("per_page")

		var resp map[string]any

		if page == "2" {
			// Second page: empty
			resp = map[string]any{
				"changes":  []any{},
				"total":    3,
				"page":     2,
				"per_page": 2,
			}
		} else {
			// Default / first page
			pp := 20
			if perPage == "2" {
				pp = 2
			}
			resp = map[string]any{
				"changes": []map[string]any{
					{
						"commit_id":       "abc123def456789000",
						"change_id":       "xyz789abc",
						"short_id":        "abc123de",
						"short_change_id": "xyz789ab",
						"author": map[string]any{
							"name":      "Alice",
							"email":     "alice@example.com",
							"timestamp": "2026-03-10T14:30:00Z",
						},
						"subject":      "Add main entry point",
						"has_conflict": false,
						"parents":      []string{"parent111"},
						"bookmarks":    []string{"main"},
					},
					{
						"commit_id":       "def456abc789000111",
						"change_id":       "uvw456def",
						"short_id":        "def456ab",
						"short_change_id": "uvw456de",
						"author": map[string]any{
							"name":      "Bob",
							"email":     "bob@example.com",
							"timestamp": "2026-03-09T10:00:00Z",
						},
						"subject":         "Initial commit",
						"has_conflict":    false,
						"is_working_copy": false,
						"parents":         []string{},
						"bookmarks":       []string{},
					},
					{
						"commit_id":       "ghi789def012345678",
						"change_id":       "rst123ghi",
						"short_id":        "ghi789de",
						"short_change_id": "rst123gh",
						"author": map[string]any{
							"name":      "Alice",
							"email":     "alice@example.com",
							"timestamp": "2026-03-10T16:00:00Z",
						},
						"subject":         "WIP: add feature",
						"has_conflict":    true,
						"is_working_copy": true,
						"parents":         []string{"abc123def456789000"},
						"bookmarks":       []string{},
					},
				},
				"total":    3,
				"page":     1,
				"per_page": pp,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Change detail endpoint: /api/jj/:owner/:repo/changes/:id
	mux.HandleFunc("/api/jj/alice/myrepo/changes/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		commitID := strings.TrimPrefix(r.URL.Path, "/api/jj/alice/myrepo/changes/")

		var resp map[string]any

		switch commitID {
		case "abc123def456789000":
			resp = map[string]any{
				"commit_id":       "abc123def456789000",
				"change_id":       "xyz789abc",
				"short_id":        "abc123de",
				"short_change_id": "xyz789ab",
				"author": map[string]any{
					"name":      "Alice",
					"email":     "alice@example.com",
					"timestamp": "2026-03-10T14:30:00Z",
				},
				"committer": map[string]any{
					"name":      "Alice",
					"email":     "alice@example.com",
					"timestamp": "2026-03-10T14:30:00Z",
				},
				"subject":      "Add main entry point",
				"message":      "Add main entry point\n\nAdded the main.go file with hello world output.",
				"has_conflict": false,
				"parents":      []string{"parent111"},
				"bookmarks":    []string{"main"},
				"stats": map[string]any{
					"files_changed": 2,
					"insertions":    15,
					"deletions":     3,
				},
				"diff": []map[string]any{
					{
						"new_path": "src/main.go",
						"status":   "added",
						"hunks": []map[string]any{
							{
								"header": "@@ -0,0 +1,7 @@",
								"lines": []map[string]any{
									{"type": "add", "content": "package main", "new_line": 1},
									{"type": "add", "content": "", "new_line": 2},
									{"type": "add", "content": "import \"fmt\"", "new_line": 3},
								},
							},
						},
					},
				},
			}
		case "conflicted123":
			resp = map[string]any{
				"commit_id":       "conflicted123",
				"change_id":       "conflictchange",
				"short_id":        "conflic1",
				"short_change_id": "conflict",
				"author": map[string]any{
					"name":      "Bob",
					"email":     "bob@example.com",
					"timestamp": "2026-03-10T16:00:00Z",
				},
				"subject":        "Conflicted merge",
				"message":        "Conflicted merge\n\nThis commit has conflicts.",
				"has_conflict":   true,
				"parents":        []string{"parent111", "parent222"},
				"bookmarks":      []string{},
				"conflict_files": []string{"file1.txt", "file2.txt"},
				"stats": map[string]any{
					"files_changed": 2,
					"insertions":    10,
					"deletions":     5,
				},
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "NOT_FOUND",
				"message": fmt.Sprintf("Commit not found: %s", commitID),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Refs endpoint: /api/jj/:owner/:repo/refs
	mux.HandleFunc("/api/jj/alice/myrepo/refs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		resp := map[string]any{
			"heads": []map[string]any{
				{"commit_id": "abc123def456789000", "change_id": "xyz789abc", "type": "anonymous"},
				{"commit_id": "ghi789def012345678", "change_id": "rst123ghi", "type": "working_copy", "name": "@"},
			},
			"bookmarks": map[string]string{
				"main":      "abc123def456789000",
				"feature-x": "def456abc789000111",
				"develop":   "ghi789def012345678",
			},
			"tags": map[string]string{
				"v1.0.0": "release111",
				"v0.9.0": "release000",
			},
			"operation_head": "op-abc-123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	return httptest.NewServer(mux)
}

// withMockSidecar sets up a mock sidecar server and configures settings,
// restoring original settings on cleanup.
func withMockSidecar(t *testing.T) (*httptest.Server, func()) {
	t.Helper()
	server := mockSidecar(t)

	origURL := setting.JJForge.SidecarURL
	origToken := setting.JJForge.InternalToken
	setting.JJForge.SidecarURL = server.URL
	setting.JJForge.InternalToken = "integration-test-token"

	cleanup := func() {
		setting.JJForge.SidecarURL = origURL
		setting.JJForge.InternalToken = origToken
		server.Close()
	}

	return server, cleanup
}

// --- Full round-trip tests ---

func TestIntegrationJjBackendListTreeRootDirectory(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	tree, err := backend.ListTree("@", "")
	require.NoError(t, err)
	require.NotNil(t, tree)

	assert.Equal(t, "", tree.Path)
	assert.Equal(t, "abc123def456", tree.Ref)
	assert.Equal(t, 3, tree.TotalEntries)
	assert.Len(t, tree.Entries, 3)

	// Verify root-level entries
	assert.Equal(t, "src", tree.Entries[0].Name)
	assert.Equal(t, "directory", tree.Entries[0].Type)
	assert.Equal(t, "README.md", tree.Entries[1].Name)
	assert.Equal(t, "file", tree.Entries[1].Type)
	assert.Equal(t, int64(256), tree.Entries[1].Size)

	// Verify README detection
	assert.Equal(t, "# My Project\n\nThis is a test project.", tree.Readme)
	assert.Equal(t, "README.md", tree.ReadmePath)
}

func TestIntegrationJjBackendListTreeSubdirectory(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	tree, err := backend.ListTree("main", "src")
	require.NoError(t, err)
	require.NotNil(t, tree)

	assert.Equal(t, "src", tree.Path)
	assert.Equal(t, 3, tree.TotalEntries)
	assert.Len(t, tree.Entries, 3)

	// Verify subdirectory entries
	assert.Equal(t, "main.go", tree.Entries[0].Name)
	assert.Equal(t, "src/main.go", tree.Entries[0].Path)
	assert.Equal(t, "file", tree.Entries[0].Type)
	assert.Equal(t, int64(1024), tree.Entries[0].Size)

	assert.Equal(t, "lib", tree.Entries[1].Name)
	assert.Equal(t, "directory", tree.Entries[1].Type)

	assert.Equal(t, "util.go", tree.Entries[2].Name)
}

func TestIntegrationJjBackendListTreeNonexistentPath(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	_, err := backend.ListTree("main", "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestIntegrationJjBackendGetBlobTextFile(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	blob, err := backend.GetBlob("main", "src/main.go")
	require.NoError(t, err)
	require.NotNil(t, blob)

	assert.Equal(t, "src/main.go", blob.Path)
	assert.Equal(t, int64(89), blob.Size)
	assert.False(t, blob.IsBinary)
	assert.False(t, blob.IsLarge)
	assert.False(t, blob.HasConflict)
	assert.Contains(t, blob.Content, "package main")
	assert.Contains(t, blob.Content, "fmt.Println")
	assert.Equal(t, "go", blob.Language)
	assert.Equal(t, "utf-8", blob.Encoding)
}

func TestIntegrationJjBackendGetBlobBinaryFile(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	blob, err := backend.GetBlob("main", "image.png")
	require.NoError(t, err)
	require.NotNil(t, blob)

	assert.Equal(t, "image.png", blob.Path)
	assert.True(t, blob.IsBinary)
	assert.False(t, blob.IsLarge)
	assert.Equal(t, "", blob.Content)
}

func TestIntegrationJjBackendGetBlobLargeFile(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	blob, err := backend.GetBlob("main", "bigfile.dat")
	require.NoError(t, err)
	require.NotNil(t, blob)

	assert.Equal(t, "bigfile.dat", blob.Path)
	assert.Equal(t, int64(10485760), blob.Size)
	assert.True(t, blob.IsBinary)
	assert.True(t, blob.IsLarge)
	assert.NotEmpty(t, blob.DownloadURL)
}

func TestIntegrationJjBackendGetBlobConflictedFile(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	blob, err := backend.GetBlob("@", "conflicted.txt")
	require.NoError(t, err)
	require.NotNil(t, blob)

	assert.True(t, blob.HasConflict)
	assert.Contains(t, blob.Content, "<<<<<<<")
	assert.Contains(t, blob.Content, ">>>>>>>")
}

func TestIntegrationJjBackendGetBlobNonexistentFile(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	_, err := backend.GetBlob("main", "nonexistent.go")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestIntegrationJjBackendGetCommitsPaginated(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")

	// First page
	commits, err := backend.GetCommits("main", "", 1, 20)
	require.NoError(t, err)
	require.NotNil(t, commits)

	assert.Equal(t, 3, commits.Total)
	assert.Equal(t, 1, commits.Page)
	assert.Len(t, commits.Commits, 3)

	// Verify first commit
	c0 := commits.Commits[0]
	assert.Equal(t, "abc123def456789000", c0.CommitID)
	assert.Equal(t, "xyz789abc", c0.ChangeID)
	assert.Equal(t, "abc123de", c0.ShortID)
	assert.Equal(t, "xyz789ab", c0.ShortChangeID)
	assert.Equal(t, "Add main entry point", c0.Subject)
	assert.Equal(t, "Alice", c0.Author.Name)
	assert.Equal(t, "alice@example.com", c0.Author.Email)
	assert.False(t, c0.HasConflict)
	assert.Equal(t, []string{"parent111"}, c0.Parents)
	assert.Equal(t, []string{"main"}, c0.Bookmarks)

	// Verify second commit
	c1 := commits.Commits[1]
	assert.Equal(t, "def456abc789000111", c1.CommitID)
	assert.Equal(t, "uvw456def", c1.ChangeID)
	assert.Equal(t, "Initial commit", c1.Subject)
	assert.Equal(t, "Bob", c1.Author.Name)
	assert.Empty(t, c1.Parents)

	// Verify third commit (conflicted working copy)
	c2 := commits.Commits[2]
	assert.True(t, c2.HasConflict)
	assert.True(t, c2.IsWorkingCopy)
}

func TestIntegrationJjBackendGetCommitsSecondPageEmpty(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")

	commits, err := backend.GetCommits("main", "", 2, 2)
	require.NoError(t, err)
	require.NotNil(t, commits)

	assert.Equal(t, 3, commits.Total)
	assert.Equal(t, 2, commits.Page)
	assert.Empty(t, commits.Commits)
}

func TestIntegrationJjBackendGetCommitsWithChangeIDs(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	commits, err := backend.GetCommits("@", "", 1, 20)
	require.NoError(t, err)

	// All jj commits should have change IDs
	for _, c := range commits.Commits {
		assert.NotEmpty(t, c.ChangeID, "jj commit %s should have a change ID", c.CommitID)
		assert.NotEmpty(t, c.ShortChangeID, "jj commit %s should have a short change ID", c.CommitID)
	}
}

func TestIntegrationJjBackendGetCommitDetail(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	commit, err := backend.GetCommit("abc123def456789000")
	require.NoError(t, err)
	require.NotNil(t, commit)

	// Verify commit info
	assert.Equal(t, "abc123def456789000", commit.CommitID)
	assert.Equal(t, "xyz789abc", commit.ChangeID)
	assert.Equal(t, "Add main entry point", commit.Subject)
	assert.Equal(t, "Add main entry point\n\nAdded the main.go file with hello world output.", commit.Message)
	assert.False(t, commit.HasConflict)

	// Verify author
	assert.Equal(t, "Alice", commit.Author.Name)
	assert.Equal(t, "alice@example.com", commit.Author.Email)

	// Verify committer
	require.NotNil(t, commit.Committer)
	assert.Equal(t, "Alice", commit.Committer.Name)

	// Verify stats
	require.NotNil(t, commit.Stats)
	assert.Equal(t, 2, commit.Stats.FilesChanged)
	assert.Equal(t, 15, commit.Stats.Insertions)
	assert.Equal(t, 3, commit.Stats.Deletions)

	// Verify diff
	require.Len(t, commit.Diff, 1)
	assert.Equal(t, "src/main.go", commit.Diff[0].NewPath)
	assert.Equal(t, "added", commit.Diff[0].Status)
	require.Len(t, commit.Diff[0].Hunks, 1)
	assert.Equal(t, "@@ -0,0 +1,7 @@", commit.Diff[0].Hunks[0].Header)
	require.Len(t, commit.Diff[0].Hunks[0].Lines, 3)
	assert.Equal(t, "add", commit.Diff[0].Hunks[0].Lines[0].Type)
	assert.Equal(t, "package main", commit.Diff[0].Hunks[0].Lines[0].Content)

	// Verify bookmarks
	assert.Equal(t, []string{"main"}, commit.Bookmarks)
	assert.Equal(t, []string{"parent111"}, commit.Parents)
}

func TestIntegrationJjBackendGetCommitWithConflicts(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	commit, err := backend.GetCommit("conflicted123")
	require.NoError(t, err)
	require.NotNil(t, commit)

	assert.True(t, commit.HasConflict)
	assert.Equal(t, []string{"file1.txt", "file2.txt"}, commit.ConflictFiles)
	assert.Len(t, commit.Parents, 2, "conflicted merge should have two parents")
}

func TestIntegrationJjBackendGetCommitNotFound(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	_, err := backend.GetCommit("nonexistent999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestIntegrationJjBackendGetRefs(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	refs, err := backend.GetRefs()
	require.NoError(t, err)
	require.NotNil(t, refs)

	// Verify bookmarks are mapped to Branches
	assert.Len(t, refs.Branches, 3)
	bookmarkNames := make(map[string]bool)
	for _, b := range refs.Branches {
		bookmarkNames[b.Name] = true
		assert.Equal(t, "bookmark", b.Type)
		assert.NotEmpty(t, b.CommitID)
	}
	assert.True(t, bookmarkNames["main"], "expected 'main' bookmark")
	assert.True(t, bookmarkNames["feature-x"], "expected 'feature-x' bookmark")
	assert.True(t, bookmarkNames["develop"], "expected 'develop' bookmark")

	// Verify tags
	assert.Len(t, refs.Tags, 2)
	tagNames := make(map[string]bool)
	for _, tg := range refs.Tags {
		tagNames[tg.Name] = true
		assert.Equal(t, "tag", tg.Type)
		assert.NotEmpty(t, tg.CommitID)
	}
	assert.True(t, tagNames["v1.0.0"], "expected 'v1.0.0' tag")
	assert.True(t, tagNames["v0.9.0"], "expected 'v0.9.0' tag")

	// Verify anonymous heads
	assert.Len(t, refs.Heads, 2)

	// Verify operation head
	assert.Equal(t, "op-abc-123", refs.OpHead)
}

func TestIntegrationJjBackendGetDefaultRefReturnsMain(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	ref, err := backend.GetDefaultRef()
	require.NoError(t, err)
	assert.Equal(t, "main", ref)
}

func TestIntegrationJjBackendGetDefaultRefFallsBackToFirstBookmark(t *testing.T) {
	// Set up a server that returns bookmarks but no "main"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"bookmarks": map[string]string{
				"develop": "abc123",
				"trunk":   "def456",
			},
			"tags":           map[string]string{},
			"heads":          []any{},
			"operation_head": "op999",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	origURL := setting.JJForge.SidecarURL
	origToken := setting.JJForge.InternalToken
	setting.JJForge.SidecarURL = server.URL
	setting.JJForge.InternalToken = "test-token"
	defer func() {
		setting.JJForge.SidecarURL = origURL
		setting.JJForge.InternalToken = origToken
	}()

	backend := NewJjBackend("alice", "myrepo")
	ref, err := backend.GetDefaultRef()
	require.NoError(t, err)
	// Should return some bookmark (map iteration order is non-deterministic,
	// but it should be one of the existing ones)
	assert.True(t, ref == "develop" || ref == "trunk",
		"expected first bookmark fallback, got %q", ref)
}

func TestIntegrationJjBackendGetDefaultRefReturnsAtWhenNoBookmarks(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"bookmarks":      map[string]string{},
			"tags":           map[string]string{},
			"heads":          []any{},
			"operation_head": "op999",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	origURL := setting.JJForge.SidecarURL
	origToken := setting.JJForge.InternalToken
	setting.JJForge.SidecarURL = server.URL
	setting.JJForge.InternalToken = "test-token"
	defer func() {
		setting.JJForge.SidecarURL = origURL
		setting.JJForge.InternalToken = origToken
	}()

	backend := NewJjBackend("alice", "myrepo")
	ref, err := backend.GetDefaultRef()
	require.NoError(t, err)
	assert.Equal(t, "@", ref)
}

// --- Auth header verification ---

func TestIntegrationJjBackendAuthHeaderSentOnAllEndpoints(t *testing.T) {
	var capturedAuth []string
	mux := http.NewServeMux()

	captureAndRespond := func(respJSON map[string]any) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = append(capturedAuth, r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(respJSON)
		}
	}

	emptyTree := map[string]any{"path": "", "ref": "@", "entries": []any{}, "total_entries": 0}
	emptyBlob := map[string]any{"path": "", "size": 0, "is_binary": false, "is_large": false, "content": ""}
	emptyChanges := map[string]any{"changes": []any{}, "total": 0, "page": 1, "per_page": 20}
	emptyRefs := map[string]any{"bookmarks": map[string]string{}, "tags": map[string]string{}, "heads": []any{}, "operation_head": ""}

	mux.HandleFunc("/api/jj/alice/myrepo/tree/", captureAndRespond(emptyTree))
	mux.HandleFunc("/api/jj/alice/myrepo/blob/", captureAndRespond(emptyBlob))
	mux.HandleFunc("/api/jj/alice/myrepo/changes", captureAndRespond(emptyChanges))
	mux.HandleFunc("/api/jj/alice/myrepo/changes/", captureAndRespond(map[string]any{
		"commit_id": "x", "short_id": "x", "author": map[string]any{
			"name": "X", "email": "x@x", "timestamp": "2026-01-01T00:00:00Z",
		}, "subject": "x", "message": "x", "parents": []string{},
	}))
	mux.HandleFunc("/api/jj/alice/myrepo/refs", captureAndRespond(emptyRefs))

	server := httptest.NewServer(mux)
	defer server.Close()

	origURL := setting.JJForge.SidecarURL
	origToken := setting.JJForge.InternalToken
	setting.JJForge.SidecarURL = server.URL
	setting.JJForge.InternalToken = "secret-auth-token"
	defer func() {
		setting.JJForge.SidecarURL = origURL
		setting.JJForge.InternalToken = origToken
	}()

	backend := NewJjBackend("alice", "myrepo")

	_, _ = backend.ListTree("@", "")
	_, _ = backend.GetBlob("@", "x")
	_, _ = backend.GetCommits("@", "", 1, 20)
	_, _ = backend.GetCommit("x")
	_, _ = backend.GetRefs()

	require.Len(t, capturedAuth, 5, "expected 5 requests (one per endpoint)")
	for i, auth := range capturedAuth {
		assert.Equal(t, "Bearer secret-auth-token", auth,
			"request %d should have the correct auth header", i)
	}
}

func TestIntegrationJjBackendNoAuthHeaderWhenTokenEmpty(t *testing.T) {
	var capturedAuth string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"path": "", "ref": "@", "entries": []any{}, "total_entries": 0,
		})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	origURL := setting.JJForge.SidecarURL
	origToken := setting.JJForge.InternalToken
	setting.JJForge.SidecarURL = server.URL
	setting.JJForge.InternalToken = ""
	defer func() {
		setting.JJForge.SidecarURL = origURL
		setting.JJForge.InternalToken = origToken
	}()

	backend := NewJjBackend("alice", "myrepo")
	_, _ = backend.ListTree("@", "")

	assert.Equal(t, "", capturedAuth, "no auth header when token is empty")
}

// --- Error handling tests ---

func TestIntegrationJjBackendSidecarReturns404(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "NOT_FOUND",
			"message": "Repository not found: alice/myrepo",
		})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	origURL := setting.JJForge.SidecarURL
	setting.JJForge.SidecarURL = server.URL
	defer func() { setting.JJForge.SidecarURL = origURL }()

	backend := NewJjBackend("alice", "myrepo")

	tests := []struct {
		name string
		fn   func() error
	}{
		{"ListTree", func() error { _, err := backend.ListTree("@", ""); return err }},
		{"GetBlob", func() error { _, err := backend.GetBlob("@", "x"); return err }},
		{"GetCommits", func() error { _, err := backend.GetCommits("@", "", 1, 20); return err }},
		{"GetCommit", func() error { _, err := backend.GetCommit("x"); return err }},
		{"GetRefs", func() error { _, err := backend.GetRefs(); return err }},
		{"GetDefaultRef", func() error { _, err := backend.GetDefaultRef(); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "404")
		})
	}
}

func TestIntegrationJjBackendSidecarReturns500(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "INTERNAL_ERROR",
			"message": "jj-lib panic: workspace reconstruction failed",
		})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	origURL := setting.JJForge.SidecarURL
	setting.JJForge.SidecarURL = server.URL
	defer func() { setting.JJForge.SidecarURL = origURL }()

	backend := NewJjBackend("alice", "myrepo")

	tests := []struct {
		name string
		fn   func() error
	}{
		{"ListTree", func() error { _, err := backend.ListTree("@", ""); return err }},
		{"GetBlob", func() error { _, err := backend.GetBlob("@", "x"); return err }},
		{"GetCommits", func() error { _, err := backend.GetCommits("@", "", 1, 20); return err }},
		{"GetCommit", func() error { _, err := backend.GetCommit("x"); return err }},
		{"GetRefs", func() error { _, err := backend.GetRefs(); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "500")
		})
	}
}

func TestIntegrationJjBackendSidecarTimeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the client timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	origURL := setting.JJForge.SidecarURL
	origToken := setting.JJForge.InternalToken
	setting.JJForge.SidecarURL = server.URL
	setting.JJForge.InternalToken = ""
	defer func() {
		setting.JJForge.SidecarURL = origURL
		setting.JJForge.InternalToken = origToken
	}()

	// Create a backend with a very short timeout
	backend := &JjBackend{
		client: &http.Client{Timeout: 100 * time.Millisecond},
		owner:  "alice",
		repo:   "myrepo",
	}

	_, err := backend.ListTree("@", "")
	require.Error(t, err)
	// The error should indicate a timeout/connection issue
	assert.Contains(t, err.Error(), "HTTP request failed")
}

func TestIntegrationJjBackendInvalidJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json content`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	origURL := setting.JJForge.SidecarURL
	setting.JJForge.SidecarURL = server.URL
	defer func() { setting.JJForge.SidecarURL = origURL }()

	backend := NewJjBackend("alice", "myrepo")

	tests := []struct {
		name string
		fn   func() error
	}{
		{"ListTree", func() error { _, err := backend.ListTree("@", ""); return err }},
		{"GetBlob", func() error { _, err := backend.GetBlob("@", "x"); return err }},
		{"GetCommits", func() error { _, err := backend.GetCommits("@", "", 1, 20); return err }},
		{"GetCommit", func() error { _, err := backend.GetCommit("x"); return err }},
		{"GetRefs", func() error { _, err := backend.GetRefs(); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unmarshal")
		})
	}
}

func TestIntegrationJjBackendSidecarUnreachable(t *testing.T) {
	origURL := setting.JJForge.SidecarURL
	// Point to a port that nothing is listening on
	setting.JJForge.SidecarURL = "http://127.0.0.1:1"
	defer func() { setting.JJForge.SidecarURL = origURL }()

	backend := &JjBackend{
		client: &http.Client{Timeout: 1 * time.Second},
		owner:  "alice",
		repo:   "myrepo",
	}

	_, err := backend.ListTree("@", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP request failed")
}

func TestIntegrationJjBackendSidecarEmptyBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Empty body
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	origURL := setting.JJForge.SidecarURL
	setting.JJForge.SidecarURL = server.URL
	defer func() { setting.JJForge.SidecarURL = origURL }()

	backend := NewJjBackend("alice", "myrepo")

	_, err := backend.ListTree("@", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

// --- URL construction tests ---

func TestIntegrationJjBackendURLConstruction(t *testing.T) {
	var capturedPaths []string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPaths = append(capturedPaths, r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"path": "", "ref": "@", "entries": []any{}, "total_entries": 0,
		})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	origURL := setting.JJForge.SidecarURL
	origToken := setting.JJForge.InternalToken
	setting.JJForge.SidecarURL = server.URL
	setting.JJForge.InternalToken = ""
	defer func() {
		setting.JJForge.SidecarURL = origURL
		setting.JJForge.InternalToken = origToken
	}()

	backend := NewJjBackend("alice", "myrepo")

	// ListTree with ref
	_, _ = backend.ListTree("main", "src/lib")
	require.Len(t, capturedPaths, 1)
	assert.Equal(t, "/api/jj/alice/myrepo/tree/src%2Flib?ref=main", capturedPaths[0])

	// ListTree without ref
	capturedPaths = nil
	_, _ = backend.ListTree("", "src")
	require.Len(t, capturedPaths, 1)
	assert.Equal(t, "/api/jj/alice/myrepo/tree/src", capturedPaths[0])
}

// --- Full round-trip: ListTree -> GetBlob pipeline ---

func TestIntegrationJjBackendListTreeThenGetBlob(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")

	// First, list the tree
	tree, err := backend.ListTree("main", "src")
	require.NoError(t, err)
	require.NotNil(t, tree)

	// Find main.go in the tree
	var mainGoEntry *TreeEntry
	for i := range tree.Entries {
		if tree.Entries[i].Name == "main.go" {
			mainGoEntry = &tree.Entries[i]
			break
		}
	}
	require.NotNil(t, mainGoEntry, "main.go should be in the tree listing")

	// Then, get the blob content using the path from the tree entry
	blob, err := backend.GetBlob("main", mainGoEntry.Path)
	require.NoError(t, err)
	require.NotNil(t, blob)

	assert.Equal(t, mainGoEntry.Path, blob.Path)
	assert.Contains(t, blob.Content, "package main")
}

// --- Timestamp parsing test ---

func TestIntegrationJjBackendTimestampParsing(t *testing.T) {
	_, cleanup := withMockSidecar(t)
	defer cleanup()

	backend := NewJjBackend("alice", "myrepo")
	commits, err := backend.GetCommits("main", "", 1, 20)
	require.NoError(t, err)
	require.NotEmpty(t, commits.Commits)

	ts := commits.Commits[0].Author.Timestamp
	assert.False(t, ts.IsZero(), "timestamp should not be zero")
	assert.Equal(t, 2026, ts.Year())
	assert.Equal(t, time.March, ts.Month())
	assert.Equal(t, 10, ts.Day())
}
