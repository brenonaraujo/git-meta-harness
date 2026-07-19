// Package hermes provides a filesystem client for the Hermes Agent
// install directory (typically ~/.hermes/).
//
// Hermes stores per-persona state in ~/.hermes/profiles/<name>/ with:
//
//	SOUL.md       — persona definition (generated from harness/personas/<name>.md)
//	config.yaml   — model, skills, etc.
//	memory/      — per-profile state (can have user customizations)
//
// And global skills in ~/.hermes/skills/<name>/SKILL.md.
package hermes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Client is a filesystem client for the Hermes install directory.
type Client struct {
	// Home is the path to ~/.hermes (or equivalent).
	Home string
}

// NewClient creates a new client. If home is empty, uses ~/.hermes.
func NewClient(home string) (*Client, error) {
	if home == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		home = filepath.Join(h, ".hermes")
	}
	return &Client{Home: home}, nil
}

// IsInstalled returns true if the Hermes directory exists.
func (c *Client) IsInstalled() bool {
	info, err := os.Stat(c.Home)
	return err == nil && info.IsDir()
}

// Profile represents a Hermes persona profile.
type Profile struct {
	Name      string
	Path      string // ~/.hermes/profiles/<name>/
	SoulPath  string // ~/.hermes/profiles/<name>/SOUL.md
	HasConfig bool   // has config.yaml
	HasMemory bool   // has memory/ subdir
}

// ListProfiles returns all installed profiles.
func (c *Client) ListProfiles() ([]Profile, error) {
	profilesDir := filepath.Join(c.Home, "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var profiles []Profile
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := Profile{
			Name:     e.Name(),
			Path:     filepath.Join(profilesDir, e.Name()),
			SoulPath: filepath.Join(profilesDir, e.Name(), "SOUL.md"),
		}
		if _, err := os.Stat(filepath.Join(p.Path, "config.yaml")); err == nil {
			p.HasConfig = true
		}
		if info, err := os.Stat(filepath.Join(p.Path, "memory")); err == nil && info.IsDir() {
			p.HasMemory = true
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

// ReadSoul reads the SOUL.md for a profile. Returns "" if not present.
func (c *Client) ReadSoul(profileName string) (string, error) {
	path := filepath.Join(c.Home, "profiles", profileName, "SOUL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// WriteSoul writes the SOUL.md for a profile, creating the dir if needed.
func (c *Client) WriteSoul(profileName, content string) error {
	dir := filepath.Join(c.Home, "profiles", profileName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "SOUL.md"), []byte(content), 0o644)
}

// Skill represents an installed skill.
type Skill struct {
	Name string
	Path string
}

// ListSkills returns all skills in ~/.hermes/skills/.
func (c *Client) ListSkills() ([]Skill, error) {
	skillsDir := filepath.Join(c.Home, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var skills []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skills = append(skills, Skill{
			Name: e.Name(),
			Path: filepath.Join(skillsDir, e.Name()),
		})
	}
	return skills, nil
}

// ReadSkill reads a skill's content (SKILL.md or main file).
//
// Returns ("", nil) if the skill is not installed (no directory).
// Returns an error only if the directory exists but has no readable
// .md file.
func (c *Client) ReadSkill(name string) (string, error) {
	// Try SKILL.md first
	p := filepath.Join(c.Home, "skills", name, "SKILL.md")
	if data, err := os.ReadFile(p); err == nil {
		return string(data), nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	// Fall back to first .md in dir
	dir := filepath.Join(c.Home, "skills", name)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Skill not installed — return empty (not an error)
			return "", nil
		}
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("no .md file in %s", dir)
}

// WriteSkill writes a skill's content to ~/.hermes/skills/<name>/SKILL.md.
func (c *Client) WriteSkill(name, content string) error {
	dir := filepath.Join(c.Home, "skills", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)
}

// WriteProfileSkill writes a skill's content to a profile's local
// skills directory: ~/.hermes/profiles/<profileName>/skills/<skillName>/SKILL.md.
//
// v1.10.3: this is required because the Hermes desktop UI counts
// only physical skills in each profile's `skills/` directory; it
// does NOT count skills from `external_dirs` (which still work at
// runtime, but show 0 in the UI). To make the UI show the right
// count, we copy the 12 harness skills into each profile's
// `skills/` directory.
//
// The 73+ Hermes global catalog skills are NOT copied (they live
// only in ~/.hermes/skills/ and are reachable via external_dirs).
func (c *Client) WriteProfileSkill(profileName, skillName, content string) error {
	dir := filepath.Join(c.Home, "profiles", profileName, "skills", skillName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)
}

// ProfileConfig represents a Hermes profile's config.yaml.
//
// IMPORTANT (v1.12.1 hotfix): ProfileConfig only describes the
// fields this package READS. WriteConfig must NEVER marshal a
// ProfileConfig back to disk directly — that would erase
// `model`, `agent`, and any other field the user configured.
// Instead, WriteConfig uses a generic map[string]interface{} so
// all fields (including unknown ones) are preserved.
//
// This struct is retained for backward compatibility with existing
// callers (ReadConfig) but should NOT be extended. New code should
// use the map-based path.
type ProfileConfig struct {
	Skills *ProfileSkills `yaml:"skills,omitempty"`
}

// ProfileSkills represents the `skills:` section of a profile config.
type ProfileSkills struct {
	ExternalDirs []string `yaml:"external_dirs,omitempty"`
}

// ReadConfig reads the config.yaml for a profile. Returns an empty
// ProfileConfig (not an error) if the file does not exist.
func (c *Client) ReadConfig(profileName string) (*ProfileConfig, error) {
	path := filepath.Join(c.Home, "profiles", profileName, "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProfileConfig{}, nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg := &ProfileConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

// writeConfigPreserveAll writes the config.yaml for a profile while
// preserving ALL fields (model, agent, custom keys, etc).
//
// The hotfix (v1.12.1) replaced the previous WriteConfig, which
// marshaled a typed ProfileConfig struct and silently erased any
// field not in the struct. That bug wiped `model.default`,
// `model.provider`, and `agent.reasoning_effort` on every
// `gmh agents sync`, breaking profiles that the user had
// configured by hand.
//
// This implementation uses a generic map[string]interface{}, so
// unknown fields are read, modified (only `skills.external_dirs`
// if needed), and rewritten. Field order may differ from the
// original file (yaml.v3 sorts map keys alphabetically) but
// the content is preserved.
func (c *Client) writeConfigPreserveAll(profileName string, externalDirs []string) error {
	path := filepath.Join(c.Home, "profiles", profileName, "config.yaml")

	// Read existing config (if any) as a generic map.
	root := map[string]interface{}{}
	data, err := os.ReadFile(path)
	if err == nil && len(data) > 0 {
		if err := yaml.Unmarshal(data, &root); err != nil {
			// Don't silently lose data on parse error — back up and
			// return an error so the user can intervene.
			backup := path + ".bak-" + fmt.Sprintf("%d", os.Getpid())
			_ = os.WriteFile(backup, data, 0o644)
			return fmt.Errorf("parse config %s: %w (backup at %s)", path, err, backup)
		}
	}

	// Update only the `skills.external_dirs` field, preserving
	// everything else (model, agent, custom keys).
	skills, _ := root["skills"].(map[string]interface{})
	if skills == nil {
		skills = map[string]interface{}{}
	}
	// Build the new external_dirs list (deduped, preserving order)
	// while leaving the user's order of OTHER keys untouched.
	existingDirs, _ := skills["external_dirs"].([]interface{})
	merged := mergeUniqueGeneric(existingDirs, externalDirs)
	skills["external_dirs"] = merged
	root["skills"] = skills

	out, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// mergeUniqueGeneric returns the union of an existing []interface{}
// (loaded from yaml) and a new []string, preserving order (existing
// first, then new entries appended if not present).
func mergeUniqueGeneric(existing []interface{}, add []string) []interface{} {
	seen := make(map[string]bool)
	out := make([]interface{}, 0, len(existing)+len(add))
	for _, v := range existing {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, s := range add {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// EnsureExternalDirs ensures the profile's config.yaml has
// `skills.external_dirs` set to the given dirs. If the profile
// already has external_dirs, the dirs are added (deduped, preserving
// order). If the file doesn't exist, it's created.
//
// v1.12.1 hotfix: now uses writeConfigPreserveAll so model/agent
// fields are preserved across `gmh agents sync` runs.
//
// Returns the final list of external_dirs after the operation.

// WriteConfig writes the config.yaml for a profile while preserving
// ALL fields the user has set (model.default, model.provider,
// agent.reasoning_effort, custom keys, etc).
//
// v1.12.1 hotfix: the previous implementation unmarshaled into a
// typed ProfileConfig struct that only knew about `skills`, then
// marshaled it back, silently erasing any other field. This broke
// profiles whenever `gmh agents sync` ran. The fix uses a generic
// map[string]interface{} (writeConfigPreserveAll) so all fields
// round-trip safely.
func (c *Client) WriteConfig(profileName string, externalDirs []string) error {
	return c.writeConfigPreserveAll(profileName, externalDirs)
}

// Returns the final list of external_dirs after the operation.
func (c *Client) EnsureExternalDirs(profileName string, dirs []string) ([]string, error) {
	// Read the existing config as a generic map to learn what's
	// already in external_dirs (we need to return the merged list).
	path := filepath.Join(c.Home, "profiles", profileName, "config.yaml")
	root := map[string]interface{}{}
	data, err := os.ReadFile(path)
	if err == nil && len(data) > 0 {
		_ = yaml.Unmarshal(data, &root)
	}
	var existing []string
	if skills, ok := root["skills"].(map[string]interface{}); ok {
		if d, ok := skills["external_dirs"].([]interface{}); ok {
			for _, v := range d {
				if s, ok := v.(string); ok {
					existing = append(existing, s)
				}
			}
		}
	}
	merged := mergeUnique(existing, dirs)
	if err := c.WriteConfig(profileName, merged); err != nil {
		return nil, err
	}
	return merged, nil
}

// mergeUnique returns the union of two string slices, preserving
// order (existing first, then new entries appended if not present).
func mergeUnique(a, b []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(a)+len(b))
	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
