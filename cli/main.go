// Command gmh is the official CLI for the meta-harness framework.
//
// gmh provides a single static binary for installing, syncing,
// and managing the meta-harness in any project. See
// https://github.com/brenonaraujo/git-meta-harness/blob/main/docs/CLI.md
// for the full manual.
//
// Install:
//
//	curl -sSL https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.sh | bash
//
// Then:
//
//	cd my-project
//	gmh install
//	gmh doctor
package main

import (
	"os"

	"github.com/brenonaraujo/git-meta-harness/cli/cmd"
)

// These vars are set at build time via -ldflags:
//
//	-X main.Version=1.6.0 -X main.Commit=$(git rev-parse --short HEAD) -X main.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	if err := cmd.Execute(Version, Commit, Date); err != nil {
		os.Exit(1)
	}
}
