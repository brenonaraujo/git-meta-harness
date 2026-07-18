# gmh — git-meta-harness CLI

> **Official CLI for the meta-harness framework.**
>
> Single static binary, written in Go, distributed via GitHub Releases.
> Install with one command (no Python, no Node, no Docker required).

## Quick start

```bash
# Install (Linux/macOS)
curl -sSL https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.sh | bash

# Install (Windows PowerShell)
iwr -useb https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.ps1 | iex

# Verify
gmh version

# Install meta-harness into your project
cd my-project
gmh install

# Sync with latest version
gmh sync

# Health check
gmh doctor
```

## Commands

| Command                  | What it does                                            |
|--------------------------|---------------------------------------------------------|
| `gmh install`            | Install meta-harness into current project              |
| `gmh sync`               | Sync the local project with latest version              |
| `gmh update --to vX.Y.Z` | Update to a specific version                            |
| `gmh doctor`             | Health check the local project                          |
| `gmh skills`             | Manage skills (install/list/remove)                     |
| `gmh personas`           | Manage personas (create specialized domain-experts)     |
| `gmh plugins`            | Manage gmh plugins (experimental)                       |
| `gmh version`            | Print version info                                      |

See [docs/CLI.md](../docs/CLI.md) for the full manual.

## Build from source

```bash
cd cli
make build    # local
make all      # cross-compile for all OS/arch
make test
```

## Architecture

```
cli/
├── main.go              # entry point (in cmd/root.go)
├── go.mod
├── Makefile
├── cmd/                 # cobra subcommands
│   ├── root.go
│   ├── install.go
│   ├── sync.go
│   ├── update.go
│   ├── doctor.go
│   ├── skills.go
│   ├── personas.go
│   ├── plugins.go
│   └── version.go
├── internal/
│   └── harness/         # read/write harness/ directory
├── installer/
│   ├── install.sh       # bootstrap (Linux/macOS)
│   └── install.ps1      # bootstrap (Windows)
└── testdata/
```

## Releases

gmh is released via the same `git-meta-harness` GitHub Releases
but with a `cli-vX.Y.Z` tag pattern. The release workflow builds
for 5 platforms (linux/darwin/windows × amd64/arm64) and uploads
binaries to the release.

See [`.github/workflows/cli-release.yml`](../.github/workflows/cli-release.yml).

## License

MIT — same as meta-harness.
