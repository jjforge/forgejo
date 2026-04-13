// Copyright 2026 The jjforge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"forgejo.org/modules/base"
)

// BinaryInfo holds metadata about a downloadable binary.
type BinaryInfo struct {
	Name        string
	Description string
	Version     string
	Platforms   []PlatformBinary
}

// PlatformBinary holds metadata about a binary for a specific platform.
type PlatformBinary struct {
	OS          string // Display name: "Linux", "macOS", "Windows"
	Arch        string // "x86_64", "arm64"
	Filename    string // "jj", "jj.exe"
	Size        string // "12.4 MiB"
	URL         string // "/downloads/jj/linux-x86_64/jj"
	platformKey string // "linux-x86_64" — used for sorting, not exported
}

var binaryNames = []string{"jj", "jjf"}

var binaryDescriptions = map[string]string{
	"jj":  "Custom build of Jujutsu VCS — includes jj forge push, jj forge pull, and jj forge clone commands for syncing with jjforge",
	"jjf": "jjforge CLI for repository management, issues, and authentication",
}

// validPlatforms is the set of allowed {os}-{arch} directory names.
var validPlatforms = map[string]bool{
	"linux-x86_64":   true,
	"linux-arm64":    true,
	"darwin-x86_64":  true,
	"darwin-arm64":   true,
	"windows-x86_64": true,
}

// platformDisplayOS maps the os prefix to a display name.
var platformDisplayOS = map[string]string{
	"linux":   "Linux",
	"darwin":  "macOS",
	"windows": "Windows",
}

func scanDownloadsDir(downloadsPath string) []BinaryInfo {
	var binaries []BinaryInfo

	for _, name := range binaryNames {
		binDir := filepath.Join(downloadsPath, name)

		versionBytes, err := os.ReadFile(filepath.Join(binDir, "VERSION"))
		if err != nil {
			continue // No VERSION file = skip this binary
		}
		version := strings.TrimSpace(string(versionBytes))

		entries, err := os.ReadDir(binDir)
		if err != nil {
			continue
		}

		var platforms []PlatformBinary
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			platformKey := entry.Name()
			if !validPlatforms[platformKey] {
				continue
			}

			parts := strings.SplitN(platformKey, "-", 2)
			if len(parts) != 2 {
				continue
			}
			osName, arch := parts[0], parts[1]

			filename := name
			if osName == "windows" {
				filename = name + ".exe"
			}

			binaryPath := filepath.Join(binDir, platformKey, filename)
			fi, err := os.Stat(binaryPath)
			if err != nil {
				continue // Binary file doesn't exist in this platform dir
			}

			platforms = append(platforms, PlatformBinary{
				OS:          platformDisplayOS[osName],
				Arch:        arch,
				Filename:    filename,
				Size:        base.FileSize(fi.Size()),
				URL:         fmt.Sprintf("/downloads/%s/%s/%s", name, platformKey, filename),
				platformKey: platformKey,
			})
		}

		sort.Slice(platforms, func(i, j int) bool {
			return platforms[i].platformKey < platforms[j].platformKey
		})

		if len(platforms) > 0 {
			binaries = append(binaries, BinaryInfo{
				Name:        name,
				Description: binaryDescriptions[name],
				Version:     version,
				Platforms:   platforms,
			})
		}
	}

	return binaries
}
