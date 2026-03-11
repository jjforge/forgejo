// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"xorm.io/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add vcs_backend_type column to repository table",
		Upgrade:     addRepositoryVCSBackendType,
	})
}

func addRepositoryVCSBackendType(x *xorm.Engine) error {
	type Repository struct {
		VCSBackendType string `xorm:"VARCHAR(10) NOT NULL DEFAULT 'git'"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(Repository))
	return err
}
