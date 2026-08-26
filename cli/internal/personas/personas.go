// Package personas manages domain-expert specializations under
// harness/personas/. Domain-experts are always specialized
// (never a generic domain-expert.md).
package personas

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const (
	reservedGeneric      = "generic"
	reservedDomainExpert = "domain-expert"
	adopterName          = "domain-expert-adopter"
	templateFile         = "domain-expert.template.md"
	skeleton             = `# Persona — Domain Expert ` + "`<domínio>`" + `

Specialized domain-expert for ` + "`<seu-dominio>`" + `.
`
)

// Slug converts a domain name to kebab-case [a-z0-9-].
// Accents are folded when possible. Empty slugs and the reserved
// values "generic" and "domain-expert" are rejected.
func Slug(domain string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(domain))
	s = stripAccents(s)

	var b strings.Builder
	b.Grow(len(s))
	prevHyphen := false
	for _, r := range s {
		switch {
		case r == ' ' || r == '_' || r == '\t' || r == '\n' || r == '-':
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevHyphen = false
		default:
			return "", fmt.Errorf("invalid domain %q: slug must match [a-z0-9-]", domain)
		}
	}

	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "", fmt.Errorf("invalid domain %q: empty slug", domain)
	}
	if out == reservedGeneric || out == reservedDomainExpert {
		return "", fmt.Errorf("invalid domain %q: reserved slug %q", domain, out)
	}
	return out, nil
}

func stripAccents(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case 'á', 'à', 'ã', 'â', 'ä', 'å':
			r = 'a'
		case 'é', 'è', 'ê', 'ë':
			r = 'e'
		case 'í', 'ì', 'î', 'ï':
			r = 'i'
		case 'ó', 'ò', 'õ', 'ô', 'ö':
			r = 'o'
		case 'ú', 'ù', 'û', 'ü':
			r = 'u'
		case 'ç':
			r = 'c'
		case 'ñ':
			r = 'n'
		case 'ý', 'ÿ':
			r = 'y'
		}
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Create writes harnessDir/personas/domain-expert-<slug>.md from the
// local template (or a short built-in skeleton). It refuses to
// overwrite an existing file. When projectContext is non-empty it
// appends a "Project context" section.
func Create(harnessDir, domain, projectContext string) (string, error) {
	slug, err := Slug(domain)
	if err != nil {
		return "", err
	}

	personasDir := filepath.Join(harnessDir, "personas")
	dest := filepath.Join(personasDir, "domain-expert-"+slug+".md")
	if _, err := os.Stat(dest); err == nil {
		return "", fmt.Errorf("persona already exists: %s", dest)
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if err := os.MkdirAll(personasDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir personas: %w", err)
	}

	body, err := readTemplate(personasDir)
	if err != nil {
		return "", err
	}
	body = strings.ReplaceAll(body, "<domínio>", slug)
	body = strings.ReplaceAll(body, "<seu-dominio>", slug)

	if projectContext != "" {
		body = strings.TrimRight(body, "\n") + "\n\n## Project context\n\n" + projectContext
		if !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
	}

	if err := os.WriteFile(dest, []byte(body), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", dest, err)
	}
	return dest, nil
}

func readTemplate(personasDir string) (string, error) {
	p := filepath.Join(personasDir, templateFile)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return skeleton, nil
		}
		return "", fmt.Errorf("read template: %w", err)
	}
	return string(data), nil
}

// List returns basenames of *.md files in harnessDir/personas.
// A missing directory yields an empty slice.
func List(harnessDir string) ([]string, error) {
	dir := filepath.Join(harnessDir, "personas")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".md") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	if out == nil {
		return []string{}, nil
	}
	return out, nil
}

// Remove deletes a domain-expert-* persona. Template and
// domain-expert-adopter files are protected.
func Remove(harnessDir, name string) error {
	base := filepath.Base(strings.TrimSpace(name))
	base = strings.TrimSuffix(base, ".md")
	if !removable(base) {
		return fmt.Errorf("refusing to remove %q: only removable domain-expert-* personas are allowed", name)
	}

	path := filepath.Join(harnessDir, "personas", base+".md")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}

func removable(base string) bool {
	if !strings.HasPrefix(base, "domain-expert-") {
		return false
	}
	if strings.Contains(base, "template") {
		return false
	}
	if base == adopterName {
		return false
	}
	return true
}
