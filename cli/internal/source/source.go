// Package source provides version resolution and tarball download
// for the meta-harness framework.
//
// It talks to the GitHub API and the GitHub Releases CDN, never to
// GitHub over a custom protocol. This keeps the CLI auditable and
// debuggable with curl.
package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultRepo is the canonical meta-harness repo on GitHub.
	DefaultRepo = "brenonaraujo/git-meta-harness"
	// DefaultHTTPTimeout is the per-request timeout for HTTP calls.
	DefaultHTTPTimeout = 30 * time.Second
)

// Release describes a single GitHub Release.
type Release struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// Client talks to GitHub.
type Client struct {
	Repo       string
	HTTPClient *http.Client
}

// NewClient creates a new source client.
func NewClient(repo string) *Client {
	if repo == "" {
		repo = DefaultRepo
	}
	return &Client{
		Repo: repo,
		HTTPClient: &http.Client{
			Timeout: DefaultHTTPTimeout,
		},
	}
}

// ResolveVersion resolves "latest" to the actual latest release tag.
// If version is already concrete (e.g., "v1.6.0"), returns it as-is.
func (c *Client) ResolveVersion(version string) (string, error) {
	if version == "" || version == "latest" {
		return c.latestVersion()
	}
	// Sanity check
	if !strings.HasPrefix(version, "v") {
		return "", fmt.Errorf("version must start with 'v' (got %q)", version)
	}
	return version, nil
}

func (c *Client) latestVersion() (string, error) {
	// Try gh CLI first (uses user's auth)
	if tag := ghLatestRelease(c.Repo); tag != "" {
		return tag, nil
	}
	// Fallback: GitHub API (public, unauthenticated, rate-limited)
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", c.Repo)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned %d for %s", resp.StatusCode, url)
	}
	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", fmt.Errorf("decode release: %w", err)
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("no tag_name in release JSON")
	}
	return rel.TagName, nil
}

// ghLatestRelease uses the `gh` CLI to fetch the latest release tag.
// Returns "" if gh is not installed or fails.
func ghLatestRelease(repo string) string {
	// Use gh release list (one shot)
	out, err := runCmd("gh", "release", "list", "--repo", repo,
		"--limit", "1", "--json", "tagName", "--jq", ".[0].tagName")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// DownloadTarball downloads the source tarball for the given version
// into a temp directory. Returns the path to the extracted directory.
func (c *Client) DownloadTarball(version, destDir string) (string, error) {
	if version == "" {
		return "", fmt.Errorf("version is empty")
	}
	// codeload URL is faster than the API for tarball downloads
	url := fmt.Sprintf("https://codeload.github.com/%s/tar.gz/%s", c.Repo, version)

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", destDir, err)
	}

	tarPath := filepath.Join(destDir, fmt.Sprintf("meta-harness-%s.tar.gz", version))
	if err := c.downloadFile(url, tarPath); err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	// Extract
	extracted, err := extractTarGz(tarPath, destDir)
	if err != nil {
		return "", fmt.Errorf("extract: %w", err)
	}

	// The extracted dir is named "<repo>-<version>/" (e.g. "git-meta-harness-1.6.0/").
	// The harnessDir is its inner harness/ subdirectory.
	harnessDir := filepath.Join(extracted, "harness")
	if _, err := os.Stat(harnessDir); err != nil {
		return "", fmt.Errorf("harness/ not found in tarball (extracted to %s): %w", extracted, err)
	}

	// Stash the extracted root path so callers can find VERSION etc.
	// (We rely on the convention that the top-level dir name starts
	// with the repo name; the caller can re-derive it.)
	_ = extracted

	return harnessDir, nil
}

// ExtractedRoot returns the top-level directory of the most recently
// downloaded tarball. This is the "<repo>-<version>" directory that
// contains both harness/ and VERSION at the top level.
func ExtractedRoot(harnessDir string) string {
	// harnessDir is "<root>/harness" — go up one
	return filepath.Dir(harnessDir)
}

func (c *Client) downloadFile(url, dest string) error {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	return nil
}

// runCmd runs a command and returns stdout.
func runCmd(name string, args ...string) ([]byte, error) {
	// Implemented in cmd_unix.go / cmd_windows.go via build tags,
	// or inline here. To keep the package standalone, use os/exec
	// inline (no build tag dance).
	return osExec(name, args...)
}
