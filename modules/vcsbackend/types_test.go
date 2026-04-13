// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefsResponse_JSON_BookmarksKey(t *testing.T) {
	resp := RefsResponse{
		Branches: []RefInfo{
			{Name: "main", CommitID: "abc123", Type: "bookmark"},
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Contains(t, raw, "bookmarks", "expected JSON key 'bookmarks'")
	assert.NotContains(t, raw, "branches", "should NOT have JSON key 'branches'")
}

func TestRefsResponse_JSON_Roundtrip(t *testing.T) {
	input := `{"bookmarks":[{"name":"main","commit_id":"abc","type":"bookmark"}],"tags":[{"name":"v1","commit_id":"def","type":"tag"}]}`

	var resp RefsResponse
	require.NoError(t, json.Unmarshal([]byte(input), &resp))

	require.Len(t, resp.Branches, 1)
	assert.Equal(t, "main", resp.Branches[0].Name)
}
