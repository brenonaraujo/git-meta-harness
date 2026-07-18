// Package gitutil wraps the `git` and `gh` CLIs for common operations.
package gitutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// CmdResult holds stdout/stderr of a command.
type CmdResult struct {
	Stdout string
	Stderr string
	Code   int
}

// Run runs a command, returning combined output and exit code.
func Run(name string, args ...string) CmdResult {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := CmdResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.Code = ee.ExitCode()
		} else {
			res.Code = -1
		}
	}
	return res
}

// IsGitRepo returns true if cwd is inside a git repo.
func IsGitRepo() bool {
	r := Run("git", "rev-parse", "--git-dir")
	return r.Code == 0
}

// IsClean returns true if there are no uncommitted changes.
func IsClean() bool {
	r := Run("git", "status", "--porcelain")
	return r.Code == 0 && strings.TrimSpace(r.Stdout) == ""
}

// CurrentBranch returns the current branch name.
func CurrentBranch() string {
	r := Run("git", "branch", "--show-current")
	return strings.TrimSpace(r.Stdout)
}

// CreateBranch creates and checks out a new branch from current HEAD.
func CreateBranch(name string) error {
	r := Run("git", "checkout", "-b", name)
	if r.Code != 0 {
		return fmt.Errorf("git checkout -b %s: %s", name, r.Stderr)
	}
	return nil
}

// AddAll stages all changes.
func AddAll() error {
	r := Run("git", "add", "-A")
	if r.Code != 0 {
		return fmt.Errorf("git add -A: %s", r.Stderr)
	}
	return nil
}

// Commit creates a commit with the given message.
func Commit(msg string) error {
	r := Run("git", "commit", "-m", msg)
	if r.Code != 0 {
		return fmt.Errorf("git commit: %s", r.Stderr)
	}
	return nil
}

// Push pushes the current branch to origin.
func Push(remote, branch string) error {
	r := Run("git", "push", "-u", remote, branch)
	if r.Code != 0 {
		return fmt.Errorf("git push: %s", r.Stderr)
	}
	return nil
}

// HasRemote returns true if `origin` remote is configured.
func HasRemote() bool {
	r := Run("git", "remote", "get-url", "origin")
	return r.Code == 0 && strings.TrimSpace(r.Stdout) != ""
}

// GhAvailable returns true if `gh` CLI is installed and authed.
func GhAvailable() bool {
	r := Run("gh", "auth", "status")
	return r.Code == 0
}

// CreatePR creates a GitHub PR using `gh pr create`.
func CreatePR(title, body, base, head string) (string, error) {
	args := []string{"pr", "create",
		"--title", title,
		"--body", body,
	}
	if base != "" {
		args = append(args, "--base", base)
	}
	if head != "" {
		args = append(args, "--head", head)
	}
	r := Run("gh", args...)
	if r.Code != 0 {
		return "", fmt.Errorf("gh pr create: %s", r.Stderr)
	}
	// gh pr create prints the PR URL on stdout
	url := strings.TrimSpace(r.Stdout)
	return url, nil
}
