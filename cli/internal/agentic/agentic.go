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
//
// All commands use the agentic's actual CLI (validated against
// the installed binary — see TestInvocation_ValidatesHermesCLI):
//   - Hermes: `hermes -p <profile> chat -q "<prompt>"`  (v1.12.2+)
//   - Claude Code: `claude -p "<prompt>"` (TBD — adjust when validated)
//   - Codex: `codex -p "<prompt>"` (TBD)
//   - OpenCode: `opencode -p "<prompt>"` (TBD)
//
// v1.12.2 HOTFIX: Hermes `-p` is a **global** flag (on the
// `hermes` root command), NOT a `hermes chat` subcommand flag.
// The previous form `hermes chat -p <profile> -q ...` fails
// silently — `hermes chat` rejects `-p` as an unknown arg.
//
// For long prompts (>2KB), prefer --query-file or stdin redirect.
func Invocation(a Agentic, profile, prompt string) (string, error) {
	switch a {
	case Hermes:
		if profile == "" {
			profile = "team-manager"
		}
		// Validated against hermes CLI: `-p <profile>` is a root flag
		// (sets the profile for the entire session). `chat -q <prompt>`
		// is the non-interactive subcommand. Order matters.
		return fmt.Sprintf("hermes -p %s chat -q %s", profile, shellQuote(prompt)), nil
	case ClaudeCode:
		// TBD — validate against actual `claude` CLI before shipping
		return fmt.Sprintf("claude -p %s", shellQuote(prompt)), nil
	case Codex:
		// TBD — validate against actual `codex` CLI
		return fmt.Sprintf("codex -p %s", shellQuote(prompt)), nil
	case OpenCode:
		// TBD — validate against actual `opencode` CLI
		return fmt.Sprintf("opencode -p %s", shellQuote(prompt)), nil
	default:
		return "", fmt.Errorf("agentic %q has no CLI invocation (write prompt to stdout instead)", a)
	}
}

// shellQuote wraps s in single quotes for safe shell interpolation.
// Embedded single quotes are escaped as '\”' (close-quote, escaped, reopen-quote).
func shellQuote(s string) string {
	// Replace each ' with '\''
	escaped := ""
	for _, r := range s {
		if r == '\'' {
			escaped += "'\\''"
		} else {
			escaped += string(r)
		}
	}
	return "'" + escaped + "'"
}
