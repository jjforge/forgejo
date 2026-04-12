// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repository

import (
	"testing"

	repo_model "forgejo.org/models/repo"

	"github.com/stretchr/testify/assert"
)

func TestCreateRepoOptionsDefaultVCSType(t *testing.T) {
	t.Run("empty VCSType defaults to jj", func(t *testing.T) {
		opts := CreateRepoOptions{}
		if opts.VCSType == "" {
			opts.VCSType = repo_model.VCSTypeJJ
		}
		assert.Equal(t, repo_model.VCSTypeJJ, opts.VCSType)
	})

	t.Run("explicit jj stays jj", func(t *testing.T) {
		opts := CreateRepoOptions{VCSType: repo_model.VCSTypeJJ}
		if opts.VCSType == "" {
			opts.VCSType = repo_model.VCSTypeJJ
		}
		assert.Equal(t, repo_model.VCSTypeJJ, opts.VCSType)
	})
}
