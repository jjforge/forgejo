// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"forgejo.org/modules/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJjBackendImplementsVCSBackend(t *testing.T) {
	var _ VCSBackend = NewJjBackend("alice", "repo")
}

func TestJjBackendListTree(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/api/jj/alice/myrepo/tree/src")
		assert.Equal(t, "main", r.URL.Query().Get("ref"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		resp := map[string]any{
			"path": "src",
			"ref":  "abc123",
			"entries": []map[string]any{
				{"name": "main.go", "path": "src/main.go", "type": "file", "size": 1024},
				{"name": "lib", "path": "src/lib", "type": "directory"},
			},
			"total_entries": 2,
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
	tree, err := backend.ListTree("main", "src")
	require.NoError(t, err)
	require.NotNil(t, tree)
	assert.Equal(t, "src", tree.Path)
	assert.Equal(t, "abc123", tree.Ref)
	assert.Len(t, tree.Entries, 2)
	assert.Equal(t, "main.go", tree.Entries[0].Name)
	assert.Equal(t, "file", tree.Entries[0].Type)
	assert.Equal(t, "lib", tree.Entries[1].Name)
	assert.Equal(t, "directory", tree.Entries[1].Type)
}

func TestJjBackendGetBlob(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/api/jj/alice/myrepo/blob/src/main.go")
		assert.Equal(t, "main", r.URL.Query().Get("ref"))

		resp := map[string]any{
			"path":      "src/main.go",
			"size":      42,
			"is_binary": false,
			"is_large":  false,
			"content":   "package main\n",
			"language":  "go",
			"encoding":  "utf-8",
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
	blob, err := backend.GetBlob("main", "src/main.go")
	require.NoError(t, err)
	require.NotNil(t, blob)
	assert.Equal(t, "src/main.go", blob.Path)
	assert.Equal(t, int64(42), blob.Size)
	assert.False(t, blob.IsBinary)
	assert.Equal(t, "package main\n", blob.Content)
	assert.Equal(t, "go", blob.Language)
}

func TestJjBackendGetCommits(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/api/jj/alice/myrepo/commits")
		assert.Equal(t, "main", r.URL.Query().Get("ref"))
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("per_page"))

		resp := map[string]any{
			"commits": []map[string]any{
				{
					"commit_id":       "abc123def456",
					"change_id":       "xyz789",
					"short_id":        "abc123de",
					"short_change_id": "xyz789ab",
					"author": map[string]any{
						"name":      "Alice",
						"email":     "alice@example.com",
						"timestamp": "2026-03-10T14:30:00Z",
					},
					"subject":      "Add main entry point",
					"has_conflict": false,
					"parents":      []string{"parent123"},
					"bookmarks":    []string{"main"},
				},
			},
			"total":    1,
			"page":     1,
			"per_page": 20,
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
	commits, err := backend.GetCommits("main", "", 1, 20)
	require.NoError(t, err)
	require.NotNil(t, commits)
	assert.Equal(t, 1, commits.Total)
	assert.Len(t, commits.Commits, 1)
	assert.Equal(t, "abc123def456", commits.Commits[0].CommitID)
	assert.Equal(t, "xyz789", commits.Commits[0].ChangeID)
	assert.Equal(t, "Add main entry point", commits.Commits[0].Subject)
}

func TestJjBackendGetCommit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/api/jj/alice/myrepo/commit/abc123")

		resp := map[string]any{
			"commit_id":       "abc123def456",
			"change_id":       "xyz789",
			"short_id":        "abc123de",
			"short_change_id": "xyz789ab",
			"author": map[string]any{
				"name":      "Alice",
				"email":     "alice@example.com",
				"timestamp": "2026-03-10T14:30:00Z",
			},
			"subject":      "Add main entry point",
			"message":      "Add main entry point\n\nDetailed description.",
			"has_conflict": false,
			"parents":      []string{"parent123"},
			"bookmarks":    []string{"main"},
			"stats": map[string]any{
				"files_changed": 1,
				"insertions":    5,
				"deletions":     0,
			},
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
	commit, err := backend.GetCommit("abc123")
	require.NoError(t, err)
	require.NotNil(t, commit)
	assert.Equal(t, "abc123def456", commit.CommitID)
	assert.Equal(t, "Add main entry point\n\nDetailed description.", commit.Message)
	require.NotNil(t, commit.Stats)
	assert.Equal(t, 1, commit.Stats.FilesChanged)
}

func TestJjBackendGetRefs(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/api/jj/alice/myrepo/refs")

		resp := map[string]any{
			"heads": []map[string]any{
				{"commit_id": "abc123", "change_id": "xyz789", "name": nil, "type": "anonymous"},
			},
			"bookmarks": map[string]string{
				"main":      "abc123",
				"feature-x": "def456",
			},
			"tags": map[string]string{
				"v1.0.0": "ghi789",
			},
			"operation_head": "op123",
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
	refs, err := backend.GetRefs()
	require.NoError(t, err)
	require.NotNil(t, refs)
	assert.Len(t, refs.Branches, 2) // bookmarks mapped to Branches
	assert.Len(t, refs.Tags, 1)
	assert.Len(t, refs.Heads, 1)
	assert.Equal(t, "op123", refs.OpHead)
}

func TestJjBackendGetDefaultRef(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/api/jj/alice/myrepo/refs")

		resp := map[string]any{
			"bookmarks": map[string]string{
				"main": "abc123",
			},
			"tags":           map[string]string{},
			"heads":          []any{},
			"operation_head": "op123",
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
	assert.Equal(t, "main", ref)
}

func TestJjBackendAuthHeaders(t *testing.T) {
	var capturedHeaders http.Header
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header
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
	setting.JJForge.InternalToken = "my-internal-token"
	defer func() {
		setting.JJForge.SidecarURL = origURL
		setting.JJForge.InternalToken = origToken
	}()

	backend := NewJjBackend("alice", "myrepo")
	_, _ = backend.ListTree("@", "")

	assert.Equal(t, "Bearer my-internal-token", capturedHeaders.Get("Authorization"))
}

func TestJjBackendSidecarError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
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
	_, err := backend.ListTree("@", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
