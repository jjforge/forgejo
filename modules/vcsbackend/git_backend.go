// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vcsbackend

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	repo_model "forgejo.org/models/repo"
	"forgejo.org/modules/git"
)

// GitBackend wraps Forgejo's existing modules/git/ package to implement
// the VCSBackend interface. Git repos behave identically to stock Forgejo.
// Each method opens the git repository, delegates to modules/git/ functions,
// and converts the result to vcsbackend response types.
type GitBackend struct {
	repo *repo_model.Repository
}

// NewGitBackend creates a new GitBackend for the given repository model.
func NewGitBackend(repo *repo_model.Repository) *GitBackend {
	return &GitBackend{repo: repo}
}

// gitEntryType converts git entry booleans to the vcsbackend type string.
func gitEntryType(isDir, isLink bool) string {
	switch {
	case isDir:
		return "directory"
	case isLink:
		return "symlink"
	default:
		return "file"
	}
}

// openGitRepo opens the git repository on disk.
func (b *GitBackend) openGitRepo() (*git.Repository, error) {
	return git.OpenRepository(context.Background(), b.repo.RepoPath())
}

func (b *GitBackend) ListTree(ref, treePath string) (*TreeResponse, error) {
	gitRepo, err := b.openGitRepo()
	if err != nil {
		return nil, fmt.Errorf("GitBackend.ListTree: open repo: %w", err)
	}
	defer gitRepo.Close()

	commit, err := gitRepo.GetCommit(ref)
	if err != nil {
		return nil, fmt.Errorf("GitBackend.ListTree: get commit for ref %q: %w", ref, err)
	}

	var tree *git.Tree
	if treePath == "" || treePath == "/" {
		tree = &commit.Tree
	} else {
		entry, err := commit.GetTreeEntryByPath(treePath)
		if err != nil {
			return nil, fmt.Errorf("GitBackend.ListTree: get tree entry at path %q: %w", treePath, err)
		}
		tree, err = entry.Tree().SubTree(treePath)
		if err != nil {
			// Fall back to getting the subtree from the commit tree
			tree, err = commit.SubTree(treePath)
			if err != nil {
				return nil, fmt.Errorf("GitBackend.ListTree: get subtree at path %q: %w", treePath, err)
			}
		}
	}

	entries, err := tree.ListEntries()
	if err != nil {
		return nil, fmt.Errorf("GitBackend.ListTree: list entries: %w", err)
	}

	resp := &TreeResponse{
		Path:         treePath,
		Ref:          commit.ID.String(),
		TotalEntries: len(entries),
		Entries:      make([]TreeEntry, 0, len(entries)),
	}

	for _, entry := range entries {
		entryPath := entry.Name()
		if treePath != "" && treePath != "/" {
			entryPath = path.Join(treePath, entry.Name())
		}
		te := TreeEntry{
			Name:         entry.Name(),
			Path:         entryPath,
			Type:         gitEntryType(entry.IsDir(), entry.IsLink()),
			IsExecutable: entry.IsExecutable(),
		}
		if !entry.IsDir() {
			te.Size = entry.Size()
		}
		resp.Entries = append(resp.Entries, te)
	}

	return resp, nil
}

func (b *GitBackend) GetBlob(ref, filePath string) (*BlobResponse, error) {
	gitRepo, err := b.openGitRepo()
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetBlob: open repo: %w", err)
	}
	defer gitRepo.Close()

	commit, err := gitRepo.GetCommit(ref)
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetBlob: get commit for ref %q: %w", ref, err)
	}

	entry, err := commit.GetTreeEntryByPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetBlob: get entry at path %q: %w", filePath, err)
	}

	blob := entry.Blob()
	size := blob.Size()

	resp := &BlobResponse{
		Path:     filePath,
		Size:     size,
		Encoding: "utf-8",
	}

	// Large file threshold: 1MB
	const maxInlineSize = 1024 * 1024
	if size > maxInlineSize {
		resp.IsLarge = true
		return resp, nil
	}

	reader, err := blob.DataAsync()
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetBlob: read blob: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetBlob: read data: %w", err)
	}

	// Binary detection: check for null bytes in first 8KB
	checkLen := len(data)
	if checkLen > 8192 {
		checkLen = 8192
	}
	for i := 0; i < checkLen; i++ {
		if data[i] == 0 {
			resp.IsBinary = true
			return resp, nil
		}
	}

	resp.Content = string(data)
	resp.Language = detectLanguage(filePath)
	return resp, nil
}

// detectLanguage returns a language identifier based on file extension.
func detectLanguage(filePath string) string {
	ext := strings.ToLower(path.Ext(filePath))
	if ext == "" {
		return ""
	}
	ext = ext[1:] // remove leading dot
	switch ext {
	case "go":
		return "go"
	case "rs":
		return "rust"
	case "py", "pyw":
		return "python"
	case "js", "jsx", "mjs", "cjs":
		return "javascript"
	case "ts", "tsx":
		return "typescript"
	case "java":
		return "java"
	case "rb":
		return "ruby"
	case "sh", "bash", "zsh":
		return "bash"
	case "yml", "yaml":
		return "yaml"
	case "json", "jsonc":
		return "json"
	case "toml":
		return "toml"
	case "md", "markdown":
		return "markdown"
	case "html", "htm":
		return "html"
	case "css":
		return "css"
	case "c", "h":
		return "c"
	case "cpp", "cc", "cxx", "hpp":
		return "cpp"
	default:
		return ""
	}
}

func (b *GitBackend) GetCommits(ref string, filePath string, page, perPage int) (*CommitsResponse, error) {
	gitRepo, err := b.openGitRepo()
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetCommits: open repo: %w", err)
	}
	defer gitRepo.Close()

	commit, err := gitRepo.GetCommit(ref)
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetCommits: get commit for ref %q: %w", ref, err)
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	var commits []*git.Commit
	var total int

	if filePath != "" {
		commits, err = commit.CommitsByRange(page, perPage, "")
	} else {
		commits, err = commit.CommitsByRange(page, perPage, "")
	}
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetCommits: list commits: %w", err)
	}

	allCount, err := commit.CommitsCount()
	if err != nil {
		total = len(commits) // fallback
	} else {
		total = int(allCount)
	}

	resp := &CommitsResponse{
		Commits: make([]CommitInfo, 0, len(commits)),
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}

	for _, c := range commits {
		ci := convertGitCommit(c)
		resp.Commits = append(resp.Commits, ci)
	}

	return resp, nil
}

// convertGitCommit converts a modules/git Commit to a vcsbackend CommitInfo.
func convertGitCommit(c *git.Commit) CommitInfo {
	ci := CommitInfo{
		CommitID: c.ID.String(),
		ShortID:  c.ID.String()[:8],
		Subject:  c.Summary(),
		Parents:  make([]string, 0, len(c.Parents)),
	}
	if c.Author != nil {
		ci.Author = Signature{
			Name:      c.Author.Name,
			Email:     c.Author.Email,
			Timestamp: c.Author.When,
		}
	}
	for _, p := range c.Parents {
		ci.Parents = append(ci.Parents, p.String())
	}
	return ci
}

func (b *GitBackend) GetCommit(id string) (*CommitDetailResponse, error) {
	gitRepo, err := b.openGitRepo()
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetCommit: open repo: %w", err)
	}
	defer gitRepo.Close()

	commit, err := gitRepo.GetCommit(id)
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetCommit: get commit %q: %w", id, err)
	}

	ci := convertGitCommit(commit)
	resp := &CommitDetailResponse{
		CommitInfo: ci,
		Message:    commit.Message(),
	}

	if commit.Committer != nil {
		resp.Committer = &Signature{
			Name:      commit.Committer.Name,
			Email:     commit.Committer.Email,
			Timestamp: commit.Committer.When,
		}
	}

	return resp, nil
}

func (b *GitBackend) GetRefs() (*RefsResponse, error) {
	gitRepo, err := b.openGitRepo()
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetRefs: open repo: %w", err)
	}
	defer gitRepo.Close()

	branches, _, err := gitRepo.GetBranches(0, 0)
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetRefs: get branches: %w", err)
	}

	resp := &RefsResponse{
		Branches: make([]RefInfo, 0, len(branches)),
	}

	for _, branch := range branches {
		ref := RefInfo{
			Name: branch.Name,
			Type: "branch",
		}
		resp.Branches = append(resp.Branches, ref)
	}

	tags, err := gitRepo.GetTags(0, 0)
	if err != nil {
		return nil, fmt.Errorf("GitBackend.GetRefs: get tags: %w", err)
	}

	resp.Tags = make([]RefInfo, 0, len(tags))
	for _, tag := range tags {
		resp.Tags = append(resp.Tags, RefInfo{
			Name: tag,
			Type: "tag",
		})
	}

	return resp, nil
}

func (b *GitBackend) GetDefaultRef() (string, error) {
	return b.repo.DefaultBranch, nil
}

func (b *GitBackend) GetDiff(commitID string) (*DiffResponse, error) {
	return nil, fmt.Errorf("GitBackend.GetDiff: not yet implemented")
}

func (b *GitBackend) CompareDiff(base, head string) (*DiffResponse, error) {
	return nil, fmt.Errorf("GitBackend.CompareDiff: not yet implemented")
}

func (b *GitBackend) BlameFile(ref, filePath string) (*BlameResponse, error) {
	return nil, fmt.Errorf("GitBackend.BlameFile: not yet implemented")
}

func (b *GitBackend) Merge(targetRef string, sourceCommits []string, strategy MergeStrategy) (*MergeResult, error) {
	return nil, fmt.Errorf("GitBackend.Merge: not yet implemented")
}

func (b *GitBackend) Rebase(commits []string, onto string) (*RebaseResult, error) {
	return nil, fmt.Errorf("GitBackend.Rebase: not yet implemented")
}

func (b *GitBackend) CreateRef(name, commitID string) error {
	return fmt.Errorf("GitBackend.CreateRef: not yet implemented")
}

func (b *GitBackend) DeleteRef(name string) error {
	return fmt.Errorf("GitBackend.DeleteRef: not yet implemented")
}

func (b *GitBackend) GetOperations(page, perPage int) (*OperationsResponse, error) {
	return nil, fmt.Errorf("GitBackend.GetOperations: not applicable for git repos")
}
