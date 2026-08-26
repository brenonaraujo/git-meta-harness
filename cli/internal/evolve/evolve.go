// Package evolve implements a Stanford-style outer loop over
// delivery personas and skills using issue/comment traces.
//
// It is the governance analog of Stanford IRIS Meta-Harness
// (arXiv:2603.28052): a filesystem of prior candidates+traces.
// This package writes that filesystem and a deterministic
// proposal. It never calls a live LLM.
package evolve

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	commentsFile = "comments.json"
	maxSnippet   = 200
	utcCompact   = "20060102T150405Z"
)

// Comment is one GitHub issue comment harvested into a trace.
type Comment struct {
	Issue     int    `json:"issue"`
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

// signalRE matches comment bodies that should drive a persona/skill
// patch suggestion. Short tokens (AC, DoD) are whole-word only so
// "acceptance" does not also fire "AC".
var (
	signalRE = regexp.MustCompile(`(?i)\b(?:acceptance|AC|DoD|skills?|personas?|domain-expert|missing|gaps?)\b`)
	acRE     = regexp.MustCompile(`(?i)\b(?:acceptance|AC|DoD)\b`)
)

// LoadComments reads comment traces from dir.
//
// If dir/comments.json exists it is decoded as a JSON array.
// Otherwise each *.json file is one Comment object (comments.json
// is skipped — already handled by the array path).
func LoadComments(dir string) ([]Comment, error) {
	arrayPath := filepath.Join(dir, commentsFile)
	data, err := os.ReadFile(arrayPath)
	if err == nil {
		var comments []Comment
		if err := json.Unmarshal(data, &comments); err != nil {
			return nil, fmt.Errorf("decode %s: %w", arrayPath, err)
		}
		return comments, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read %s: %w", arrayPath, err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	var comments []Comment
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.EqualFold(filepath.Ext(name), ".json") {
			continue
		}
		if name == commentsFile {
			continue
		}
		path := filepath.Join(dir, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var c Comment
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, fmt.Errorf("decode %s: %w", path, err)
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// Propose builds a markdown evolve proposal from comments.
// Matching is deterministic (keyword scan); no live LLM.
func Propose(comments []Comment, personaNames []string) string {
	var b strings.Builder
	b.WriteString("# Persona/skill evolve proposal\n\n")

	b.WriteString("## Personas in play\n\n")
	if len(personaNames) == 0 {
		b.WriteString("- (none listed)\n")
	} else {
		for _, name := range personaNames {
			b.WriteString(fmt.Sprintf("- `%s`\n", name))
		}
	}
	b.WriteString("\n")

	b.WriteString("## Signals from issue comments\n\n")
	n := 0
	for _, c := range comments {
		if !signalRE.MatchString(c.Body) {
			continue
		}
		n++
		snip := snippet(c.Body, maxSnippet)
		b.WriteString(fmt.Sprintf("### Issue #%d (@%s)\n\n", c.Issue, c.Author))
		b.WriteString(fmt.Sprintf("> %s\n\n", snip))
		b.WriteString(fmt.Sprintf("**Suggest:** patch %s.\n\n", suggestPatch(c.Body, personaNames)))
	}
	if n == 0 {
		b.WriteString("No comments mentioned acceptance, AC, DoD, skill, persona, domain-expert, missing, or gap.\n\n")
	}

	b.WriteString("## Hermes (OS/tool harness)\n\n")
	b.WriteString("The team-manager should run this proposal using Hermes tools (`gh`, filesystem).\n")
	b.WriteString("git-meta-harness only stores the delivery context (traces under `harness/memory/traces/`).\n")
	b.WriteString("It does not call a live LLM and never overwrites persona markdown in v1.\n\n")
	b.WriteString("Run this prompt in Hermes team-manager (OS/tool harness).\n\n")

	b.WriteString("## PROMPT\n\n")
	b.WriteString("```\n")
	b.WriteString(buildPrompt(comments, personaNames))
	b.WriteString("```\n")
	return b.String()
}

// WriteTrace persists comments + proposal under
// memoryDir/traces/<utc compact, e.g. 20060102T150405Z>/.
// Files: comments.json, proposal.md, PROMPT.md (the prompt section).
func WriteTrace(memoryDir string, comments []Comment, proposal string) (string, error) {
	stamp := time.Now().UTC().Format(utcCompact)
	traceDir := filepath.Join(memoryDir, "traces", stamp)
	if err := os.MkdirAll(traceDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", traceDir, err)
	}

	raw, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode comments: %w", err)
	}
	if err := os.WriteFile(filepath.Join(traceDir, commentsFile), append(raw, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write comments.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(traceDir, "proposal.md"), []byte(proposal), 0o644); err != nil {
		return "", fmt.Errorf("write proposal.md: %w", err)
	}
	prompt := extractPrompt(proposal)
	if err := os.WriteFile(filepath.Join(traceDir, "PROMPT.md"), []byte(prompt), 0o644); err != nil {
		return "", fmt.Errorf("write PROMPT.md: %w", err)
	}
	return traceDir, nil
}

func buildPrompt(comments []Comment, personaNames []string) string {
	var b strings.Builder
	b.WriteString("You are the team-manager of this meta-harness project.\n\n")
	b.WriteString("Apply the persona/skill evolve proposal using Hermes tools (gh, filesystem).\n")
	b.WriteString("git-meta-harness only stores the delivery context — do not invent traces.\n")
	b.WriteString("Do NOT overwrite persona markdown unless a human explicitly approves.\n\n")
	if len(personaNames) > 0 {
		b.WriteString("Personas in play: ")
		b.WriteString(strings.Join(personaNames, ", "))
		b.WriteString("\n\n")
	}
	b.WriteString("Quoted signals:\n")
	any := false
	for _, c := range comments {
		if !signalRE.MatchString(c.Body) {
			continue
		}
		any = true
		b.WriteString(fmt.Sprintf("- #%d @%s: %s\n", c.Issue, c.Author, snippet(c.Body, maxSnippet)))
	}
	if !any {
		b.WriteString("- (none)\n")
	}
	return b.String()
}

func extractPrompt(proposal string) string {
	const marker = "## PROMPT"
	i := strings.Index(proposal, marker)
	if i < 0 {
		return proposal
	}
	return strings.TrimSpace(proposal[i:]) + "\n"
}

func snippet(body string, max int) string {
	s := strings.Join(strings.Fields(body), " ")
	if len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}

func suggestPatch(body string, personaNames []string) string {
	named := namedPersona(body, personaNames)
	lower := strings.ToLower(body)
	hasSkill := strings.Contains(lower, "skill")
	hasDE := strings.Contains(lower, "domain-expert")
	hasAC := acRE.MatchString(body)

	switch {
	case named != "" && hasSkill:
		return fmt.Sprintf("persona `%s` and the mentioned skill", named)
	case named != "":
		return fmt.Sprintf("persona `%s`", named)
	case hasDE && hasSkill:
		return "persona `domain-expert` and the mentioned skill"
	case hasDE:
		return "persona `domain-expert`"
	case hasSkill:
		return "the mentioned skill"
	case hasAC:
		return "persona `domain-expert` (acceptance / DoD)"
	default:
		return "the relevant persona or skill"
	}
}

func namedPersona(body string, personaNames []string) string {
	lower := strings.ToLower(body)
	for _, name := range personaNames {
		if name == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(name)) {
			return name
		}
	}
	return ""
}
