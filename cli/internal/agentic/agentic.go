// Package agentic detects which agentic CLI a user has installed
// and configured, and provides helpers to invoke each.
package agentic

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Agentic is one of the supported agentic CLIs.
type Agentic string

const (
	// Hermes is the open-source Hermes Agent (validated by meta-harness).
	Hermes Agentic = "hermes"
	// ClaudeCode is Anthropic's Claude Code.
	ClaudeCode Agentic = "claude-code"
	// Codex is OpenAI's Codex CLI.
	Codex Agentic = "codex"
	// OpenCode is the opencode CLI.
	OpenCode Agentic = "opencode"
	// Copilot is GitHub Copilot.
	Copilot Agentic = "copilot"
	// Devin is Cognition's Devin CLI.
	Devin Agentic = "devin"
	// Cursor is the Cursor editor.
	Cursor Agentic = "cursor"
	// None means no agentic is detected.
	None Agentic = "none"
)

// String returns the canonical name.
func (a Agentic) String() string {
	return string(a)
}

// Detect returns the agentic the user is most likely using, based
// on what's installed. Order of preference:
//
//  1. Hermes (if installed and has profiles for the meta-harness)
//  2. Claude Code (if .claude/ in cwd or ~/.claude/)
//  3. Codex (if .codex/ in cwd or ~/.codex/)
//  4. OpenCode (if .opencode/ in cwd or ~/.opencode/)
//  5. Copilot (if .github/copilot-instructions.md exists)
//  6. Cursor (if .cursorrules exists)
//  7. None
func Detect() (Agentic, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return None, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return None, err
	}

	// Hermes: command exists AND has team-manager profile
	if _, err := exec.LookPath("hermes"); err == nil {
		// Check if team-manager profile exists
		teamMgr := filepath.Join(home, ".hermes", "profiles", "team-manager")
		if _, err := os.Stat(teamMgr); err == nil {
			return Hermes, nil
		}
		// Hermes installed but no team-manager profile
		// Still return Hermes as detected — user might be using
		// a different profile or no profile
		if dirExists(filepath.Join(home, ".hermes")) {
			return Hermes, nil
		}
	}

	// Claude Code
	if dirExists(filepath.Join(cwd, ".claude")) ||
		dirExists(filepath.Join(home, ".claude")) {
		if _, err := exec.LookPath("claude"); err == nil {
			return ClaudeCode, nil
		}
	}

	// Codex
	if dirExists(filepath.Join(cwd, ".codex")) ||
		dirExists(filepath.Join(home, ".codex")) {
		if _, err := exec.LookPath("codex"); err == nil {
			return Codex, nil
		}
	}

	// OpenCode
	if dirExists(filepath.Join(cwd, ".opencode")) ||
		dirExists(filepath.Join(home, ".opencode")) {
		if _, err := exec.LookPath("opencode"); err == nil {
			return OpenCode, nil
		}
	}

	// Copilot
	if fileExists(filepath.Join(cwd, ".github", "copilot-instructions.md")) {
		return Copilot, nil
	}

	// Cursor
	if fileExists(filepath.Join(cwd, ".cursorrules")) {
		return Cursor, nil
	}

	return None, nil
}

// ListAll returns ALL agentics detected (not just the most likely).
func ListAll() []Agentic {
	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	var found []Agentic

	if _, err := exec.LookPath("hermes"); err == nil {
		if dirExists(filepath.Join(home, ".hermes")) {
			found = append(found, Hermes)
		}
	}
	if _, err := exec.LookPath("claude"); err == nil {
		if dirExists(filepath.Join(cwd, ".claude")) || dirExists(filepath.Join(home, ".claude")) {
			found = append(found, ClaudeCode)
		}
	}
	if _, err := exec.LookPath("codex"); err == nil {
		if dirExists(filepath.Join(cwd, ".codex")) || dirExists(filepath.Join(home, ".codex")) {
			found = append(found, Codex)
		}
	}
	if _, err := exec.LookPath("opencode"); err == nil {
		if dirExists(filepath.Join(cwd, ".opencode")) || dirExists(filepath.Join(home, ".opencode")) {
			found = append(found, OpenCode)
		}
	}
	if fileExists(filepath.Join(cwd, ".github", "copilot-instructions.md")) {
		found = append(found, Copilot)
	}
	if fileExists(filepath.Join(cwd, ".cursorrules")) {
		found = append(found, Cursor)
	}
	return found
}

// IsValid returns true if a is a recognized agentic.
func IsValid(a Agentic) bool {
	switch a {
	case Hermes, ClaudeCode, Codex, OpenCode, Copilot, Devin, Cursor:
		return true
	}
	return false
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// Invocation returns the shell command to invoke this agentic with
// the given prompt. Used by `gmh doctor` to delegate the actual
// harness health check to the agentic.
func Invocation(a Agentic, profile, prompt string) (string, error) {
	switch a {
	case Hermes:
		if profile == "" {
			profile = "team-manager"
		}
		// Hermes invocation (hypothetical — adjust to actual Hermes CLI)
		return fmt.Sprintf("hermes profile %s --prompt %q", profile, prompt), nil
	case ClaudeCode:
		// claude -p "<prompt>"
		return fmt.Sprintf("claude -p %q", prompt), nil
	case Codex:
		return fmt.Sprintf("codex -p %q", prompt), nil
	case OpenCode:
		return fmt.Sprintf("opencode -p %q", prompt), nil
	default:
		return "", fmt.Errorf("agentic %q has no CLI invocation (write prompt to stdout instead)", a)
	}
}
