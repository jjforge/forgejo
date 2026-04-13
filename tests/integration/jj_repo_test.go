// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"net/http"
	"testing"

	"forgejo.org/tests"
)

// TestJJRepoIssuesList verifies the issues page renders for jj repos without panic.
// jj repos have nil GitRepo — issue template functions must handle this gracefully.
func TestJJRepoIssuesList(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/user2/jj-test-repo/issues")
	session.MakeRequest(t, req, http.StatusOK)
}

// TestJJRepoIssueNew verifies the new issue page renders for jj repos.
func TestJJRepoIssueNew(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/user2/jj-test-repo/issues/new")
	session.MakeRequest(t, req, http.StatusOK)
}

// TestJJRepoHome verifies the repo home page renders for jj repos.
func TestJJRepoHome(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/user2/jj-test-repo")
	session.MakeRequest(t, req, http.StatusOK)
}

// TestJJRepoSettings verifies the settings page renders for jj repos.
func TestJJRepoSettings(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/user2/jj-test-repo/settings")
	session.MakeRequest(t, req, http.StatusOK)
}

// TestJJRepoPullsBlocked verifies git-only pages return 404 for jj repos.
func TestJJRepoPullsBlocked(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/user2/jj-test-repo/pulls")
	session.MakeRequest(t, req, http.StatusNotFound)
}

// TestJJRepoBranchesBlocked verifies branches page returns 404 for jj repos.
func TestJJRepoBranchesBlocked(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/user2/jj-test-repo/branches")
	session.MakeRequest(t, req, http.StatusNotFound)
}

// TestJJRepoReleasesBlocked verifies releases page returns 404 for jj repos.
func TestJJRepoReleasesBlocked(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/user2/jj-test-repo/releases")
	session.MakeRequest(t, req, http.StatusNotFound)
}
