package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/source"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/ui"
)

// InstallCmd creates the `gmh install` command.
//
// `gmh install` copies the `harness/` directory from a specific
// version of the git-meta-harness repo into the current project.
func InstallCmd() *cobra.Command {
	var (
		toVersion string
		force     bool
		skipCheck bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install meta-harness into the current project",
		Long: `Install the meta-harness framework into the current project.

This creates a 'harness/' directory at the project root containing
all the framework files (AGENTS.md, personas, sensors, scripts, etc.).

Examples:
  gmh install                     # Install latest version
  gmh install --to v1.5.0         # Install a specific version
  gmh install --force             # Overwrite existing harness/
  gmh install --skip-check        # Don't run doctor after install`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			harnessDir := filepath.Join(cwd, "harness")

			// Check if harness/ already exists
			if _, err := os.Stat(harnessDir); err == nil && !force {
				ui.Fail("harness/ already exists at %s", harnessDir)
				ui.Info("Use --force to overwrite, or run `gmh sync` to update")
				return fmt.Errorf("harness/ exists")
			}

			src := source.NewClient("")
			version, err := src.ResolveVersion(toVersion)
			if err != nil {
				return fmt.Errorf("resolve version: %w", err)
			}
			ui.Info("Installing meta-harness %s into %s", version, cwd)

			// Download to temp dir
			tmp, err := os.MkdirTemp("", "gmh-install-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmp)

			remoteHarness, err := src.DownloadTarball(version, tmp)
			if err != nil {
				return fmt.Errorf("download: %w", err)
			}
			ui.Step("Downloaded tarball, extracted to %s", remoteHarness)

			// Copy remoteHarness/* → cwd/harness/
			if err := os.MkdirAll(harnessDir, 0o755); err != nil {
				return err
			}
			if err := copyDir(remoteHarness, harnessDir); err != nil {
				return fmt.Errorf("copy: %w", err)
			}

			// Also copy top-level files (VERSION, LICENSE — but keep framework
			// LICENSE separate from project LICENSE; for now skip).
			// We copy VERSION so the project can self-report its framework version.
			extractedRoot := source.ExtractedRoot(remoteHarness)
			versionFile := filepath.Join(extractedRoot, "VERSION")
			if _, err := os.Stat(versionFile); err == nil {
				_ = copyFile(versionFile, filepath.Join(cwd, "VERSION"))
			}

			ui.OK("Installed meta-harness %s", version)
			ui.Info("harness/ created at %s", harnessDir)

			// Optionally run doctor
			if !skipCheck {
				ui.Info("")
				ui.Info("Running gmh doctor to verify...")
				ui.Info("")
				return DoctorCmd().RunE(cmd, args)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&toVersion, "to", "",
		"Specific version to install (e.g., v1.5.0). Default: latest")
	cmd.Flags().BoolVarP(&force, "force", "f", false,
		"Overwrite existing harness/ directory")
	cmd.Flags().BoolVar(&skipCheck, "skip-check", false,
		"Skip running gmh doctor after install")

	return cmd
}

// stripV removes the "v" prefix from a version.
func stripV(v string) string {
	if len(v) > 0 && v[0] == 'v' {
		return v[1:]
	}
	return v
}

// copyDir recursively copies src dir to dst dir.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

// copyFile copies a single file from src to dst, preserving mode.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
