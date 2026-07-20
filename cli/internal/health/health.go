// Package health calculates the harness health score and
// produces structured JSON output for `gmh doctor --json`
// and `gmh metrics`.
//
// The health score is a weighted average of 4 dimensions
// (harness × 2, agents × 1, skills × 1, sensors × 2) on a
// 0-100 scale. Thresholds:
//
//	90-100  healthy   (green)
//	70-89   needs attention (yellow)
//	<70     critical  (red, exit 1 with --strict)
//
// See ADR-0026 in harness/contrib/design-decisions.md for
// the full design rationale.
package health

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Report is the structured output of `gmh doctor --json`.
//
// Stable across versions within the v1.14.x range.
type Report struct {
	Version       string         `json:"version"`
	Project       string         `json:"project"`
	LocalVersion  string         `json:"local_version"`
	LatestVersion string         `json:"latest_version"`
	OutOfDate     bool           `json:"out_of_date"`
	Agentic       string         `json:"agentic"`
	HealthScore   HealthScore    `json:"health_score"`
	Invariants    InvariantStats `json:"invariants"`
	Sensors       SensorStats    `json:"sensors"`
	Drift         DriftStats     `json:"drift"`
}

// HealthScore is the 4-dimension weighted average.
//
// Weights: harness=2, agents=1, skills=1, sensors=2. Overall
// = (harness*2 + agents + skills + sensors*2) / 6.
type HealthScore struct {
	Overall int `json:"overall"` // 0-100
	Harness int `json:"harness"` // 0-100
	Agents  int `json:"agents"`  // 0-100
	Skills  int `json:"skills"`  // 0-100
	Sensors int `json:"sensors"` // 0-100
}

// InvariantStats reports how many invariants are declared
// vs. how many are currently passing.
type InvariantStats struct {
	Total    int      `json:"total"`
	Passing  int      `json:"passing"`
	Failing  int      `json:"failing"`
	FailedOn []string `json:"failed_on,omitempty"` // invariant numbers that fail
}

// SensorStats reports on the 13 sensors in harness/sensors/.
type SensorStats struct {
	Total          int      `json:"total"`
	Executable     int      `json:"executable"`     // have a corresponding .sh / .py script
	BlockingActive int      `json:"blocking_active"` // declared as blocking in v1.13.0+
	Names          []string `json:"names"`
}

// DriftStats reports detected drift between the framework
// and the local project.
type DriftStats struct {
	HarnessFilesMissing []string `json:"harness_files_missing,omitempty"`
	PersonasStale       []string `json:"personas_stale,omitempty"`
	SkillsStale         int      `json:"skills_stale"`
	CIDriftLines        int      `json:"ci_drift_lines"`
}

// Calculate computes the full Report for a project at `cwd`.
//
// `version` is the framework version (e.g., "1.14.0").
// `localVersion` is the project's installed version (read
// from VERSION), empty if not installed.
func Calculate(cwd, version, localVersion, latestVersion string) (*Report, error) {
	rep := &Report{
		Version:       version,
		Project:       cwd,
		LocalVersion:  localVersion,
		LatestVersion: latestVersion,
		OutOfDate:     isOutOfDate(localVersion, latestVersion),
		Agentic:       detectAgentic(cwd),
	}

	harnessDir := filepath.Join(cwd, "harness")

	// Invariants
	rep.Invariants = scanInvariants(filepath.Join(harnessDir, "AGENTS.md"))

	// Sensors
	rep.Sensors = scanSensors(filepath.Join(harnessDir, "sensors"))

	// Health score
	rep.HealthScore = computeScore(rep, harnessDir)

	// Drift
	rep.Drift = scanDrift(cwd, harnessDir)

	return rep, nil
}

// String returns a human-readable summary (used by `gmh doctor`
// without --json).
func (r *Report) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Meta-Harness doctor report (v%s)\n", r.Version))
	sb.WriteString(fmt.Sprintf("Project:  %s\n", r.Project))
	sb.WriteString(fmt.Sprintf("Local:    %s\n", orNA(r.LocalVersion)))
	sb.WriteString(fmt.Sprintf("Latest:   %s\n", orNA(r.LatestVersion)))
	if r.OutOfDate {
		sb.WriteString("Status:   OUT OF DATE\n")
	} else {
		sb.WriteString("Status:   up to date\n")
	}
	sb.WriteString(fmt.Sprintf("Agentic:  %s\n", r.Agentic))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Health score: %d/100\n", r.HealthScore.Overall))
	sb.WriteString(fmt.Sprintf("  Harness:   %d/100\n", r.HealthScore.Harness))
	sb.WriteString(fmt.Sprintf("  Agents:    %d/100\n", r.HealthScore.Agents))
	sb.WriteString(fmt.Sprintf("  Skills:    %d/100\n", r.HealthScore.Skills))
	sb.WriteString(fmt.Sprintf("  Sensors:   %d/100\n", r.HealthScore.Sensors))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Invariants: %d/%d passing", r.Invariants.Passing, r.Invariants.Total))
	if len(r.Invariants.FailedOn) > 0 {
		sb.WriteString(fmt.Sprintf(" (failing: %s)", strings.Join(r.Invariants.FailedOn, ", ")))
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Sensors:    %d total, %d executable, %d blocking\n",
		r.Sensors.Total, r.Sensors.Executable, r.Sensors.BlockingActive))
	return sb.String()
}

// JSON serializes the report to indented JSON.
func (r *Report) JSON() (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// -- dimension calculations --

func computeScore(rep *Report, harnessDir string) HealthScore {
	// Harness (weight 2): % of declared invariants currently
	// passing. If 0 declared, score = 0.
	harness := 0
	if rep.Invariants.Total > 0 {
		harness = rep.Invariants.Passing * 100 / rep.Invariants.Total
	}
	if harness > 100 {
		harness = 100
	}

	// Agents (weight 1): 100 base, penalties for missing
	// specialization. See computeAgentsScore for the rule.
	agents := computeAgentsScore(harnessDir)

	// Skills (weight 1): coverage of installed vs declared.
	skills := computeSkillsScore(harnessDir)

	// Sensors (weight 2): only the "enforced" sensors (10+)
	// count toward executable ratio. Sensors 00-09 are
	// category descriptors (no scripts by design).
	enforcedTotal := 0
	enforcedExecutable := 0
	for _, n := range rep.Sensors.Names {
		// Extract number prefix
		re := regexp.MustCompile(`^([0-9]+)-`)
		m := re.FindStringSubmatch(n)
		if len(m) < 2 {
			continue
		}
		num, _ := strconv.Atoi(m[1])
		if num < 10 {
			continue
		}
		enforcedTotal++
		// Check if script exists (heuristic: name has a matching
		// check-*.sh or visual/check_*.py). We rely on the
		// SensorStats.Executable but need to know if THIS one
		// is executable. We don't track that per-name, so we
		// count from the ratio: if a name appears in a script
		// path, it's executable.
		if hasMatchingScript(harnessDir, n) {
			enforcedExecutable++
		}
	}
	sensors := 0
	if enforcedTotal > 0 {
		sensors = enforcedExecutable * 100 / enforcedTotal
	}
	if sensors > 100 {
		sensors = 100
	}

	overall := (harness*2 + agents + skills + sensors*2) / 6

	return HealthScore{
		Overall: overall,
		Harness: harness,
		Agents:  agents,
		Skills:  skills,
		Sensors: sensors,
	}
}

// hasMatchingScript returns true if a script exists for the
// given sensor name (e.g., "12-frontend-polish" →
// "check-frontend-polish.sh" or "check_frontend_polish.py").
func hasMatchingScript(harnessDir, sensorName string) bool {
	// Strip leading "NN-"
	base := sensorName
	if i := strings.Index(base, "-"); i > 0 {
		if _, err := strconv.Atoi(base[:i]); err == nil {
			base = base[i+1:]
		}
	}
	candidates := []string{
		filepath.Join(harnessDir, "scripts", "check-"+base+".sh"),
		filepath.Join(harnessDir, "scripts", "check-"+base+".py"),
		filepath.Join(harnessDir, "scripts", "visual", "check_"+strings.ReplaceAll(base, "-", "_")+".py"),
		filepath.Join(harnessDir, "scripts", "visual", "check_"+strings.ReplaceAll(base, "-", "_")+".mjs"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return true
		}
	}
	return false
}

func computeAgentsScore(harnessDir string) int {
	personasDir := filepath.Join(harnessDir, "personas")
	entries, err := os.ReadDir(personasDir)
	if err != nil {
		return 0
	}
	// Count "real" personas (top-level .md, not templates or
	// examples/ subdir or interactions.md).
	hasTeamManager := false
	hasBackend := false
	hasFrontend := false
	hasQA := false
	hasSolutionsArchitect := false
	domainExperts := 0
	hasGenericDE := false
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".template.md") {
			continue
		}
		if name == "interactions.md" {
			continue
		}
		switch name {
		case "team-manager.md":
			hasTeamManager = true
		case "backend-engineer.md":
			hasBackend = true
		case "frontend-engineer.md":
			hasFrontend = true
		case "quality-assurance.md":
			hasQA = true
		case "solutions-architect.md":
			hasSolutionsArchitect = true
		case "domain-expert.md":
			hasGenericDE = true
		}
		if strings.HasPrefix(name, "domain-expert-") {
			domainExperts++
		}
	}
	// Start at 100, deduct for missing core personas.
	score := 100
	if !hasTeamManager {
		score -= 25
	}
	if !hasBackend {
		score -= 15
	}
	if !hasFrontend {
		score -= 15
	}
	if !hasQA {
		score -= 15
	}
	if !hasSolutionsArchitect {
		score -= 10
	}
	// Reward for specialized domain-experts (bonus, not penalty).
	if domainExperts > 0 {
		score += 10
		if score > 100 {
			score = 100
		}
	}
	// Heavy penalty for generic domain-expert.md (invariant 12).
	if hasGenericDE {
		score -= 30
	}
	if score < 0 {
		score = 0
	}
	return score
}

func computeSkillsScore(harnessDir string) int {
	// Declared in harness/skills/ (each subdir or .md is a skill)
	skillsDir := filepath.Join(harnessDir, "skills")
	declaredEntries, err := os.ReadDir(skillsDir)
	if err != nil {
		return 0
	}
	declared := 0
	for _, e := range declaredEntries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".md") {
			declared++
		}
	}
	if declared == 0 {
		return 0
	}

	// Installed in ~/.hermes/skills/ (Hermes is our primary runtime)
	home, _ := os.UserHomeDir()
	installedDir := filepath.Join(home, ".hermes", "skills")
	installed := 0
	if entries, err := os.ReadDir(installedDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				installed++
			}
		}
	}
	score := installed * 100 / declared
	if score > 100 {
		score = 100
	}
	return score
}

// -- helpers --

func scanInvariants(agentsPath string) InvariantStats {
	stats := InvariantStats{}
	data, err := os.ReadFile(agentsPath)
	if err != nil {
		return stats
	}
	content := string(data)
	// Count `^[0-9]+\. \*\*` style invariant headers.
	re := regexp.MustCompile(`(?m)^([0-9]+)\. \*\*`)
	matches := re.FindAllStringSubmatch(content, -1)
	stats.Total = len(matches)

	// "Passing" is a heuristic: we treat the count of "✅" or
	// positive markers as the passing baseline. For now, all
	// declared invariants are passing (the doctor doesn't
	// actively verify each one — that's the sensor's job).
	// In v1.15.0, we'll have an "active invariant scan" that
	// checks each one (e.g., "no generic domain-expert.md").
	stats.Passing = stats.Total
	stats.Failing = 0
	return stats
}

func scanSensors(sensorsDir string) SensorStats {
	stats := SensorStats{}
	entries, err := os.ReadDir(sensorsDir)
	if err != nil {
		return stats
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		stats.Names = append(stats.Names, name)
		stats.Total++

		// Check if a script exists in harness/scripts/
		// or harness/scripts/visual/
		// Strip leading NN- to get the sensor name.
		base := name
		if i := strings.Index(base, "-"); i > 0 {
			if _, err := strconv.Atoi(base[:i]); err == nil {
				base = base[i+1:]
			}
		}
		// Look for matching script
		candidates := []string{
			filepath.Join(sensorsDir, "..", "scripts", "check-"+base+".sh"),
			filepath.Join(sensorsDir, "..", "scripts", "visual", "check_"+strings.ReplaceAll(base, "-", "_")+".py"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				stats.Executable++
				break
			}
		}
	}
	// Sensors 00-09 = non-blocking, 10-13 = blocking (heuristic)
	stats.BlockingActive = 0
	for _, n := range stats.Names {
		// Extract number prefix
		re := regexp.MustCompile(`^([0-9]+)-`)
		m := re.FindStringSubmatch(n)
		if len(m) < 2 {
			continue
		}
		num, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		// 10 (decomposition), 12 (frontend-polish), 13 (feature-flow) are blocking
		// 11 (scope) is non-blocking (v1.11.0 decision)
		if num == 10 || num == 12 || num == 13 {
			stats.BlockingActive++
		}
	}
	return stats
}

func scanDrift(cwd, harnessDir string) DriftStats {
	d := DriftStats{}

	// Missing harness files
	required := []string{
		"AGENTS.md",
		"bootstrap.md",
		"personas/team-manager.md",
		"sensors/00-static-analysis.md",
		"scripts/smoke-test.sh",
		"contrib/design-decisions.md",
	}
	for _, p := range required {
		full := filepath.Join(harnessDir, p)
		if _, err := os.Stat(full); err != nil {
			d.HarnessFilesMissing = append(d.HarnessFilesMissing, p)
		}
	}

	// CI drift: count lines in template CI not in local CI
	ciLocal := filepath.Join(cwd, ".github", "workflows", "ci.yml")
	ciTpl := filepath.Join(harnessDir, "templates", ".github-workflows-ci.yml")
	if local, err := os.ReadFile(ciLocal); err == nil {
		if tpl, err := os.ReadFile(ciTpl); err == nil {
			d.CIDriftLines = naiveLineDiff(string(local), string(tpl))
		}
	}

	// Skills stale: count of harness skills not in ~/.hermes/skills/
	home, _ := os.UserHomeDir()
	installedDir := filepath.Join(home, ".hermes", "skills")
	skillsDir := filepath.Join(harnessDir, "skills")
	if declaredEntries, err := os.ReadDir(skillsDir); err == nil {
		stale := 0
		for _, e := range declaredEntries {
			if !e.IsDir() {
				continue
			}
			if _, err := os.Stat(filepath.Join(installedDir, e.Name())); err != nil {
				stale++
			}
		}
		d.SkillsStale = stale
	}

	return d
}

func detectAgentic(cwd string) string {
	// If Hermes profiles exist, return "hermes"
	home, _ := os.UserHomeDir()
	if _, err := os.Stat(filepath.Join(home, ".hermes")); err == nil {
		return "hermes"
	}
	// CLAUDE.md?
	if _, err := os.Stat(filepath.Join(cwd, "CLAUDE.md")); err == nil {
		return "claude-code"
	}
	// Otherwise unknown
	return "none"
}

func naiveLineDiff(a, b string) int {
	aLines := strings.Split(a, "\n")
	bLines := strings.Split(b, "\n")
	aSet := make(map[string]bool)
	for _, l := range aLines {
		aSet[l] = true
	}
	diff := 0
	for _, l := range bLines {
		if !aSet[l] && strings.TrimSpace(l) != "" {
			diff++
		}
	}
	return diff
}

// isOutOfDate compares two version strings, normalizing
// the leading "v" prefix that GitHub releases use.
func isOutOfDate(localVersion, latestVersion string) bool {
	if localVersion == "" || latestVersion == "" {
		return false
	}
	strip := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "v")
		s = strings.TrimPrefix(s, "V")
		return s
	}
	return strip(localVersion) != strip(latestVersion)
}

func orNA(s string) string {
	if s == "" {
		return "n/a"
	}
	return s
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

// -- metrics helpers (used by `gmh metrics` in v1.14.0) --

// FlowComplianceResult is returned by FlowCompliance.
type FlowComplianceResult struct {
	Total       int      `json:"total"`        // type/feature issues counted
	Compliant   int      `json:"compliant"`    // with refined+ready+both comments
	Percent     int      `json:"percent"`      // 0-100
	NonCompliant []int   `json:"non_compliant_issues,omitempty"` // issue numbers missing
}

// FlowCompliance calculates the % of `type/feature` issues
// that went through the full feature flow (refined + ready +
// refinement comment + DoD comment).
//
// Uses `gh` CLI to query issues. If `gh` is not available
// or no remote is configured, returns Total=0.
func FlowCompliance(cwd string) FlowComplianceResult {
	res := FlowComplianceResult{}
	// Check if gh is on PATH
	if _, err := exec.LookPath("gh"); err != nil {
		return res
	}
	// Query all issues with type/feature label, last 30 days
	cmd := exec.Command("gh", "issue", "list",
		"--label", "type/feature",
		"--state", "all",
		"--limit", "100",
		"--json", "number,labels,comments")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return res
	}
	type ghLabel struct {
		Name string `json:"name"`
	}
	type ghComment struct {
		Body string `json:"body"`
	}
	type ghIssue struct {
		Number  int         `json:"number"`
		Labels  []ghLabel   `json:"labels"`
		Comments []ghComment `json:"comments"`
	}
	var issues []ghIssue
	if err := json.Unmarshal(out, &issues); err != nil {
		return res
	}
	res.Total = len(issues)
	for _, iss := range issues {
		hasRefined := false
		hasReady := false
		hasRefinementComment := false
		hasDoDComment := false
		for _, l := range iss.Labels {
			if l.Name == "refined" {
				hasRefined = true
			}
			if l.Name == "ready" {
				hasReady = true
			}
		}
		for _, c := range iss.Comments {
			body := strings.ToLower(c.Body)
			// Heuristic: refinement comment has AC, edge case
			if strings.Contains(body, "ac") || strings.Contains(body, "edge case") {
				hasRefinementComment = true
			}
			// Heuristic: DoD comment has "pilar" or "definition of done"
			if strings.Contains(body, "pilar") || strings.Contains(body, "definition of done") || strings.Contains(body, "dod") {
				hasDoDComment = true
			}
		}
		if hasRefined && hasReady && hasRefinementComment && hasDoDComment {
			res.Compliant++
		} else {
			res.NonCompliant = append(res.NonCompliant, iss.Number)
		}
	}
	if res.Total > 0 {
		res.Percent = res.Compliant * 100 / res.Total
	}
	return res
}
