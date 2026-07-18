// Package skills provides operations on the meta-harness skills
// registry (harness/skills/*.md) and on installed skills.
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Skill represents a skill in the meta-harness framework.
type Skill struct {
	Name    string
	Path    string // absolute path to the SKILL.md
	Content string
}

// ListFromDir returns all skills in a given harness/skills directory.
// Each subdirectory with a SKILL.md is a skill; loose .md files at
// the top level are also detected.
func ListFromDir(dir string) ([]Skill, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Skill
	for _, e := range entries {
		// Skip hidden and special
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		if e.IsDir() {
			// Directory with SKILL.md inside
			skillFile := filepath.Join(full, "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				data, err := os.ReadFile(skillFile)
				if err != nil {
					return nil, err
				}
				out = append(out, Skill{
					Name:    e.Name(),
					Path:    skillFile,
					Content: string(data),
				})
			}
		} else if strings.HasSuffix(e.Name(), ".md") {
			// Loose .md file: treat as skill with same name
			data, err := os.ReadFile(full)
			if err != nil {
				return nil, err
			}
			out = append(out, Skill{
				Name:    strings.TrimSuffix(e.Name(), ".md"),
				Path:    full,
				Content: string(data),
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Manifest is a summary of skills in the framework.
type Manifest struct {
	Skills []Skill
	ByName map[string]Skill
}

// BuildManifest builds a manifest from a skills directory.
func BuildManifest(dir string) (*Manifest, error) {
	skills, err := ListFromDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list skills from %s: %w", dir, err)
	}
	m := &Manifest{
		Skills: skills,
		ByName: make(map[string]Skill),
	}
	for _, s := range skills {
		m.ByName[s.Name] = s
	}
	return m, nil
}
