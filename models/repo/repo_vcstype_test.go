// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVCSTypeConstants(t *testing.T) {
	assert.Equal(t, VCSType("git"), VCSTypeGit)
	assert.Equal(t, VCSType("jj"), VCSTypeJJ)
}

func TestRepositoryIsJJ(t *testing.T) {
	t.Run("default is git", func(t *testing.T) {
		repo := &Repository{}
		assert.False(t, repo.IsJJ())
	})

	t.Run("explicit git", func(t *testing.T) {
		repo := &Repository{VCSBackendType: VCSTypeGit}
		assert.False(t, repo.IsJJ())
	})

	t.Run("jj repo", func(t *testing.T) {
		repo := &Repository{VCSBackendType: VCSTypeJJ}
		assert.True(t, repo.IsJJ())
	})
}

func TestVCSTypeString(t *testing.T) {
	assert.Equal(t, "git", string(VCSTypeGit))
	assert.Equal(t, "jj", string(VCSTypeJJ))
}
