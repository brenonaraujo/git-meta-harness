package source

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extractTarGz extracts a .tar.gz file to destDir and returns the
// path to the top-level directory that was extracted.
//
// GitHub tarballs have a single top-level directory like
// "git-meta-harness-1.6.0/". We detect and return that.
func extractTarGz(tarPath, destDir string) (string, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	// Track the top-level dir name. We collect candidates from
	// regular files/dirs (skipping pax headers) and pick the
	// most common one.
	dirCount := make(map[string]int)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Skip pax extended headers
		if hdr.Typeflag == tar.TypeXGlobalHeader || hdr.Typeflag == tar.TypeXHeader {
			continue
		}

		// Sanitize: prevent zip-slip
		target := filepath.Join(destDir, hdr.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return "", fmt.Errorf("illegal path in tarball: %s", hdr.Name)
		}

		// Track top-level dir from this entry
		if parts := strings.SplitN(hdr.Name, "/", 2); len(parts) > 0 && parts[0] != "" {
			dirCount[parts[0]]++
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			// Ensure parent dir exists
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return "", err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
		default:
			// Skip other types (symlinks, etc.) for safety
			continue
		}
	}

	// Pick the most common top-level dir
	if len(dirCount) == 0 {
		return "", fmt.Errorf("tarball has no top-level directory")
	}
	var topDir string
	maxCount := 0
	for d, c := range dirCount {
		if c > maxCount {
			maxCount = c
			topDir = d
		}
	}
	return filepath.Join(destDir, topDir), nil
}
