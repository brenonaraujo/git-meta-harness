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
func (c *Client) ReadSkill(name string) (string, error) {
	// Try SKILL.md first
	p := filepath.Join(c.Home, "skills", name, "SKILL.md")
	if data, err := os.ReadFile(p); err == nil {
		return string(data), nil
	}
	// Fall back to first .md in dir
	dir := filepath.Join(c.Home, "skills", name)
	entries, err := os.ReadDir(dir)
	if err != nil {
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
