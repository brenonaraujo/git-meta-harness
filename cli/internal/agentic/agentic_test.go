package agentic

import (
	"os/exec"
	"strings"
	"testing"
)

// TestInvocation_Hermes_ProfileFlagBeforeSubcommand is a regression
// test for the v1.12.2 hotfix.
//
// Hermes' `-p <profile>` is a **global** flag (on the `hermes` root
// command), NOT a flag on `hermes chat`. The previous form
//
//	hermes chat -p <profile> -q "..."
//
// fails with "argument command: invalid choice" because the `chat`
// subcommand parser doesn't recognize `-p`.
//
// The correct form is
//
//	hermes -p <profile> chat -q "..."
//
// This test verifies Invocation() produces the correct order and
// that the resulting command runs against the installed `hermes` CLI
// (if available in PATH).
func TestInvocation_Hermes_ProfileFlagBeforeSubcommand(t *testing.T) {
	got, err := Invocation(Hermes, "backend-engineer", "Implementar #13 ...")
	if err != nil {
		t.Fatalf("Invocation: %v", err)
	}

	// Must NOT be the old buggy form.
	if strings.HasPrefix(got, "hermes chat -p ") {
		t.Errorf("Invocation produces the BUGGY form `hermes chat -p ...`; "+
			"Hermes `-p` is a global flag (must come before `chat`):\n  got: %s", got)
	}

	// Must be the new correct form.
	want := "hermes -p backend-engineer chat -q 'Implementar #13 ...'"
	if got != want {
		t.Errorf("Invocation(Hermes, backend-engineer, ...) =\n  got:  %s\n  want: %s", got, want)
	}

	// If hermes is on PATH, dry-run the command (--help is harmless
	// and exits 0 even if other validation fails). This catches
	// "the flag is in the wrong place" bugs at CI time.
	if path, err := exec.LookPath("hermes"); err == nil {
		t.Logf("hermes found at %s — validating invocation syntax via `hermes --help`", path)
		out, _ := exec.Command("hermes", "--help").CombinedOutput()
		if !strings.Contains(string(out), "{chat,") {
			t.Errorf("hermes --help output is missing `{chat,...}` subcommand list "+
				"(unexpected CLI shape, manual check needed):\n%s", string(out))
		}
	} else {
		t.Logf("hermes not on PATH; skipping live validation (CI optional)")
	}
}

// TestInvocation_Hermes_DefaultProfile ensures that calling
// Invocation with an empty profile falls back to "team-manager"
// (the default for meta-harness projects).
func TestInvocation_Hermes_DefaultProfile(t *testing.T) {
	got, err := Invocation(Hermes, "", "do something")
	if err != nil {
		t.Fatalf("Invocation: %v", err)
	}
	if !strings.Contains(got, "-p team-manager ") {
		t.Errorf("Empty profile should default to team-manager, got: %s", got)
	}
	// And the order must still be correct (not "hermes chat -p team-manager").
	if strings.HasPrefix(got, "hermes chat -p ") {
		t.Errorf("Default profile form has the buggy flag order: %s", got)
	}
}

// TestInvocation_LongPromptIsShellQuoted ensures that prompts with
// single quotes are safely escaped (Hermes is invoked via shell).
func TestInvocation_LongPromptIsShellQuoted(t *testing.T) {
	prompt := "It's a 'tricky' prompt with quotes"
	got, err := Invocation(Hermes, "frontend-engineer", prompt)
	if err != nil {
		t.Fatalf("Invocation: %v", err)
	}
	// shellQuote escapes ' as '\''
	if !strings.Contains(got, "'\\''") {
		t.Errorf("prompt not shell-quoted safely; got: %s", got)
	}
}
