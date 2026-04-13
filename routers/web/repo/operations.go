// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"net/http"

	"forgejo.org/modules/setting"
	"forgejo.org/modules/vcsbackend"
	"forgejo.org/services/context"
)

const tplOperations = "repo/operations"

// Operations renders the jj operation log page.
func Operations(ctx *context.Context) {
	ctx.Data["Title"] = "Operation Log"
	ctx.Data["PageIsOperations"] = true
	vcsbackend.InjectContextData(ctx.Data, ctx.Repo.Repository)

	if !ctx.Repo.Repository.IsJJ() {
		ctx.NotFound("Operations", nil)
		return
	}

	backend := vcsbackend.GetBackend(ctx.Repo.Repository)

	page := ctx.FormInt("page")
	if page <= 1 {
		page = 1
	}
	pageSize := ctx.FormInt("limit")
	if pageSize <= 0 {
		pageSize = setting.Git.CommitsRangeSize
	}

	resp, err := backend.GetOperations(page, pageSize)
	if err != nil {
		ctx.ServerError("GetOperations", err)
		return
	}

	ctx.Data["Operations"] = resp.Operations

	total := resp.Total
	if total == 0 {
		total = len(resp.Operations)
	}
	pager := context.NewPagination(total, pageSize, page, 5)
	pager.SetDefaultParams(ctx)
	ctx.Data["Page"] = pager

	ctx.HTML(http.StatusOK, tplOperations)
}
