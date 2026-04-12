// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"forgejo.org/modules/setting"
)

// JjBackend calls the jj-sidecar Browse API over HTTP to implement the
// VCSBackend interface for jj repositories. All repository content operations
// are forwarded to the sidecar service.
type JjBackend struct {
	client *http.Client
	owner  string
	repo   string
}

// NewJjBackend creates a new JjBackend for the given owner/repo pair.
// It uses a shared HTTP client with a 30-second timeout.
func NewJjBackend(owner, repo string) *JjBackend {
	return &JjBackend{
		client: &http.Client{Timeout: 30 * time.Second},
		owner:  owner,
		repo:   repo,
	}
}

// baseURL returns the sidecar Browse API base URL for this repository.
func (b *JjBackend) baseURL() string {
	return fmt.Sprintf("%s/api/jj/%s/%s", setting.JJForge.SidecarURL, b.owner, b.repo)
}

// newRequest creates an HTTP GET request with the required auth headers.
func (b *JjBackend) newRequest(method, reqURL string) (*http.Request, error) {
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if setting.JJForge.InternalToken != "" {
		req.Header.Set("Authorization", "Bearer "+setting.JJForge.InternalToken)
	}
	return req, nil
}

// doRequest executes an HTTP request and returns the response body.
// Returns an error for non-2xx status codes.
func (b *JjBackend) doRequest(req *http.Request) ([]byte, error) {
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("sidecar returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (b *JjBackend) ListTree(ref, path string) (*TreeResponse, error) {
	reqURL := fmt.Sprintf("%s/tree/%s", b.baseURL(), url.PathEscape(path))
	if ref != "" {
		reqURL += "?ref=" + url.QueryEscape(ref)
	}

	req, err := b.newRequest("GET", reqURL)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.ListTree: create request: %w", err)
	}

	body, err := b.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.ListTree: %w", err)
	}

	var resp TreeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("JjBackend.ListTree: unmarshal: %w", err)
	}
	return &resp, nil
}

func (b *JjBackend) GetBlob(ref, path string) (*BlobResponse, error) {
	reqURL := fmt.Sprintf("%s/blob/%s", b.baseURL(), url.PathEscape(path))
	if ref != "" {
		reqURL += "?ref=" + url.QueryEscape(ref)
	}

	req, err := b.newRequest("GET", reqURL)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.GetBlob: create request: %w", err)
	}

	body, err := b.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.GetBlob: %w", err)
	}

	var resp BlobResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("JjBackend.GetBlob: unmarshal: %w", err)
	}
	return &resp, nil
}

func (b *JjBackend) GetCommits(ref string, path string, page, perPage int) (*CommitsResponse, error) {
	params := url.Values{}
	if ref != "" {
		params.Set("ref", ref)
	}
	if path != "" {
		params.Set("path", path)
	}
	params.Set("page", strconv.Itoa(page))
	params.Set("per_page", strconv.Itoa(perPage))

	reqURL := fmt.Sprintf("%s/commits?%s", b.baseURL(), params.Encode())

	req, err := b.newRequest("GET", reqURL)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.GetCommits: create request: %w", err)
	}

	body, err := b.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.GetCommits: %w", err)
	}

	var resp CommitsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("JjBackend.GetCommits: unmarshal: %w", err)
	}
	return &resp, nil
}

func (b *JjBackend) GetCommit(id string) (*CommitDetailResponse, error) {
	reqURL := fmt.Sprintf("%s/commit/%s", b.baseURL(), url.PathEscape(id))

	req, err := b.newRequest("GET", reqURL)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.GetCommit: create request: %w", err)
	}

	body, err := b.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.GetCommit: %w", err)
	}

	var resp CommitDetailResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("JjBackend.GetCommit: unmarshal: %w", err)
	}
	return &resp, nil
}

// jjRefsRaw is the raw JSON structure returned by the sidecar's refs endpoint.
// It differs from our RefsResponse, so we need to convert it.
type jjRefsRaw struct {
	Heads     []RefInfo         `json:"heads"`
	Bookmarks map[string]string `json:"bookmarks"`
	Tags      map[string]string `json:"tags"`
	OpHead    string            `json:"operation_head"`
}

func (b *JjBackend) GetRefs() (*RefsResponse, error) {
	reqURL := fmt.Sprintf("%s/refs", b.baseURL())

	req, err := b.newRequest("GET", reqURL)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.GetRefs: create request: %w", err)
	}

	body, err := b.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.GetRefs: %w", err)
	}

	var raw jjRefsRaw
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("JjBackend.GetRefs: unmarshal: %w", err)
	}

	resp := &RefsResponse{
		Branches: make([]RefInfo, 0, len(raw.Bookmarks)),
		Tags:     make([]RefInfo, 0, len(raw.Tags)),
		Heads:    raw.Heads,
		OpHead:   raw.OpHead,
	}

	for name, commitID := range raw.Bookmarks {
		resp.Branches = append(resp.Branches, RefInfo{
			Name:     name,
			CommitID: commitID,
			Type:     "bookmark",
		})
	}

	for name, commitID := range raw.Tags {
		resp.Tags = append(resp.Tags, RefInfo{
			Name:     name,
			CommitID: commitID,
			Type:     "tag",
		})
	}

	return resp, nil
}

func (b *JjBackend) GetDefaultRef() (string, error) {
	// For jj repos, the default ref is the first bookmark.
	// We call GetRefs and pick "main" if it exists, otherwise the first bookmark.
	refs, err := b.GetRefs()
	if err != nil {
		return "", fmt.Errorf("JjBackend.GetDefaultRef: %w", err)
	}

	for _, branch := range refs.Branches {
		if branch.Name == "main" {
			return "main", nil
		}
	}

	// Fall back to first bookmark
	if len(refs.Branches) > 0 {
		return refs.Branches[0].Name, nil
	}

	return "@", nil
}

func (b *JjBackend) GetDiff(commitID string) (*DiffResponse, error) {
	return nil, fmt.Errorf("JjBackend.GetDiff: not yet implemented")
}

func (b *JjBackend) CompareDiff(base, head string) (*DiffResponse, error) {
	return nil, fmt.Errorf("JjBackend.CompareDiff: not yet implemented")
}

func (b *JjBackend) BlameFile(ref, path string) (*BlameResponse, error) {
	return nil, fmt.Errorf("JjBackend.BlameFile: not yet implemented")
}

func (b *JjBackend) Merge(targetRef string, sourceCommits []string, strategy MergeStrategy) (*MergeResult, error) {
	return nil, fmt.Errorf("JjBackend.Merge: not yet implemented")
}

func (b *JjBackend) Rebase(commits []string, onto string) (*RebaseResult, error) {
	return nil, fmt.Errorf("JjBackend.Rebase: not yet implemented")
}

func (b *JjBackend) CreateRef(name, commitID string) error {
	return fmt.Errorf("JjBackend.CreateRef: not yet implemented")
}

func (b *JjBackend) DeleteRef(name string) error {
	return fmt.Errorf("JjBackend.DeleteRef: not yet implemented")
}

func (b *JjBackend) GetOperations(page, perPage int) (*OperationsResponse, error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("per_page", fmt.Sprintf("%d", perPage))

	reqURL := fmt.Sprintf("%s/operations?%s", b.baseURL(), params.Encode())

	req, err := b.newRequest("GET", reqURL)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.GetOperations: %w", err)
	}

	body, err := b.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("JjBackend.GetOperations: %w", err)
	}

	var resp OperationsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("JjBackend.GetOperations: unmarshal: %w", err)
	}
	return &resp, nil
}
