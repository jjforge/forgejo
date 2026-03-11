// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	repo_model "forgejo.org/models/repo"
	"forgejo.org/modules/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDispatchGetBackendReturnsCorrectTypeForAllVCSTypes(t *testing.T) {
	tests := []struct {
		name         string
		vcsType      repo_model.VCSType
		expectType   string
		expectJj     bool
	}{
		{
			name:       "git repo returns GitBackend",
			vcsType:    repo_model.VCSTypeGit,
			expectType: "*vcsbackend.GitBackend",
			expectJj:   false,
		},
		{
			name:       "jj repo returns JjBackend",
			vcsType:    repo_model.VCSTypeJJ,
			expectType: "*vcsbackend.JjBackend",
			expectJj:   true,
		},
		{
			name:       "empty vcs type defaults to GitBackend",
			vcsType:    "",
			expectType: "*vcsbackend.GitBackend",
			expectJj:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repo_model.Repository{
				OwnerName:      "alice",
				Name:           "myrepo",
				VCSBackendType: tt.vcsType,
			}
			backend := GetBackend(repo)
			require.NotNil(t, backend)

			if tt.expectJj {
				jjb, ok := backend.(*JjBackend)
				assert.True(t, ok, "expected JjBackend")
				assert.Equal(t, "alice", jjb.owner)
				assert.Equal(t, "myrepo", jjb.repo)
			} else {
				gitb, ok := backend.(*GitBackend)
				assert.True(t, ok, "expected GitBackend")
				assert.Equal(t, repo, gitb.repo)
			}
		})
	}
}

func TestDispatchJjBackendCallsMockSidecar(t *testing.T) {
	// Set up a mock sidecar that returns a tree listing
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/api/jj/alice/jjrepo/tree/")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"path":          "",
			"ref":           "abc123",
			"entries":       []any{},
			"total_entries": 0,
		})
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

	// Use GetBackend to dispatch
	repo := &repo_model.Repository{
		OwnerName:      "alice",
		Name:           "jjrepo",
		VCSBackendType: repo_model.VCSTypeJJ,
	}
	backend := GetBackend(repo)

	// Call ListTree through the dispatched backend
	tree, err := backend.ListTree("@", "")
	require.NoError(t, err)
	require.NotNil(t, tree)
	assert.Equal(t, "abc123", tree.Ref)
}

func TestDispatchGitBackendGetDefaultRef(t *testing.T) {
	// GitBackend.GetDefaultRef reads from the model and does not need a real repo
	repo := &repo_model.Repository{
		OwnerName:      "bob",
		Name:           "gitrepo",
		VCSBackendType: repo_model.VCSTypeGit,
		DefaultBranch:  "develop",
	}
	backend := GetBackend(repo)

	ref, err := backend.GetDefaultRef()
	require.NoError(t, err)
	assert.Equal(t, "develop", ref)
}

func TestDispatchPreservesOwnerAndRepoForJjBackend(t *testing.T) {
	// Verify that dispatch correctly passes owner/repo through to JjBackend
	var capturedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"bookmarks":      map[string]string{"main": "abc"},
			"tags":           map[string]string{},
			"heads":          []any{},
			"operation_head": "",
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

	repo := &repo_model.Repository{
		OwnerName:      "myorg",
		Name:           "special-repo",
		VCSBackendType: repo_model.VCSTypeJJ,
	}
	backend := GetBackend(repo)
	_, err := backend.GetRefs()
	require.NoError(t, err)

	assert.Contains(t, capturedPath, "/api/jj/myorg/special-repo/refs")
}

func TestDispatchBothBackendsImplementVCSBackend(t *testing.T) {
	// Ensure both backend types satisfy the interface
	gitRepo := &repo_model.Repository{VCSBackendType: repo_model.VCSTypeGit}
	jjRepo := &repo_model.Repository{
		OwnerName:      "x",
		Name:           "y",
		VCSBackendType: repo_model.VCSTypeJJ,
	}

	var _ VCSBackend = GetBackend(gitRepo)
	var _ VCSBackend = GetBackend(jjRepo)
}

func TestDispatchMultipleReposGetSeparateBackends(t *testing.T) {
	gitRepo := &repo_model.Repository{
		OwnerName:      "alice",
		Name:           "git-project",
		VCSBackendType: repo_model.VCSTypeGit,
		DefaultBranch:  "main",
	}
	jjRepo := &repo_model.Repository{
		OwnerName:      "alice",
		Name:           "jj-project",
		VCSBackendType: repo_model.VCSTypeJJ,
	}

	gitBackend := GetBackend(gitRepo)
	jjBackend := GetBackend(jjRepo)

	_, gitOk := gitBackend.(*GitBackend)
	_, jjOk := jjBackend.(*JjBackend)

	assert.True(t, gitOk, "git repo should get GitBackend")
	assert.True(t, jjOk, "jj repo should get JjBackend")
}
