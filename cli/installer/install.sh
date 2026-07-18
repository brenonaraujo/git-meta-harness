#!/usr/bin/env bash
# ============================================================================
# GMH (git-meta-harness) CLI — Bootstrap Installer
# ============================================================================
# Installs the gmh binary by downloading the latest release artifact
# for the current OS/arch.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.sh | bash
#
# Environment variables:
#   GMH_INSTALL    — install dir (default: $HOME/.gmh/bin)
#   GMH_VERSION    — specific version to install (default: latest)
#   GMH_REPO       — source repo (default: brenonaraujo/git-meta-harness)
# ============================================================================
set -euo pipefail

# Defaults
: "${GMH_INSTALL:=$HOME/.gmh/bin}"
: "${GMH_VERSION:=latest}"
: "${GMH_REPO:=brenonaraujo/git-meta-harness}"

# Colors (if TTY)
if [ -t 1 ]; then
  CYAN='\033[0;36m'
  YELLOW='\033[0;33m'
  GREEN='\033[0;32m'
  RED='\033[0;31m'
  BOLD='\033[1m'
  NC='\033[0m'
else
  CYAN=''; YELLOW=''; GREEN=''; RED=''; BOLD=''; NC=''
fi

info()  { echo -e "${CYAN}==>${NC} $1"; }
ok()    { echo -e "${GREEN}✅${NC} $1"; }
warn()  { echo -e "${YELLOW}⚠️${NC}  $1"; }
fail()  { echo -e "${RED}❌${NC} $1"; }

# ---------------------------------------------------------------------------
# 1. Detect OS + Arch
# ---------------------------------------------------------------------------
info "Detecting platform..."

OS_RAW=$(uname -s)
ARCH_RAW=$(uname -m)

case "$OS_RAW" in
  Linux)  OS=linux ;;
  Darwin) OS=darwin ;;
  MINGW*|CYGWIN*|MSYS*) OS=windows ;;
  *)
    fail "Unsupported OS: $OS_RAW"
    fail "gmh supports: linux, darwin, windows (use WSL on Windows)"
    exit 1
    ;;
esac

case "$ARCH_RAW" in
  x86_64)           ARCH=amd64 ;;
  amd64)            ARCH=amd64 ;;
  aarch64|arm64)    ARCH=arm64 ;;
  *)
    fail "Unsupported architecture: $ARCH_RAW"
    fail "gmh supports: amd64 (x86_64), arm64 (aarch64)"
    exit 1
    ;;
esac

BIN_NAME="gmh-$OS-$ARCH"
if [ "$OS" = "windows" ]; then
  BIN_NAME="$BIN_NAME.exe"
fi

ok "Platform: $OS/$ARCH"

# ---------------------------------------------------------------------------
# 2. Resolve version
# ---------------------------------------------------------------------------
if [ "$GMH_VERSION" = "latest" ]; then
  info "Resolving latest version..."
  # GitHub API: latest release
  LATEST_URL="https://api.github.com/repos/$GMH_REPO/releases/latest"
  if command -v gh >/dev/null 2>&1; then
    GMH_VERSION=$(gh release list --repo "$GMH_REPO" --limit 1 \
      --json tagName --jq '.[0].tagName' 2>/dev/null || echo "")
  fi
  if [ -z "$GMH_VERSION" ]; then
    GMH_VERSION=$(curl -fsSL --max-time 10 \
      "$LATEST_URL" 2>/dev/null \
      | grep '"tag_name"' | head -1 \
      | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/')
  fi
  if [ -z "$GMH_VERSION" ]; then
    fail "Could not resolve latest version. Set GMH_VERSION explicitly."
    fail "Example: GMH_VERSION=v1.6.0 curl -sSL ... | bash"
    exit 1
  fi
fi
ok "Version: $GMH_VERSION"

# ---------------------------------------------------------------------------
# 3. Create install dir
# ---------------------------------------------------------------------------
info "Install dir: $GMH_INSTALL"
mkdir -p "$GMH_INSTALL"

# ---------------------------------------------------------------------------
# 4. Download binary
# ---------------------------------------------------------------------------
DOWNLOAD_URL="https://github.com/$GMH_REPO/releases/download/$GMH_VERSION/$BIN_NAME"
info "Downloading: $DOWNLOAD_URL"

if ! curl -fsSL --max-time 60 -o "$GMH_INSTALL/gmh.tmp" "$DOWNLOAD_URL"; then
  fail "Download failed. Check:"
  fail "  - URL: $DOWNLOAD_URL"
  fail "  - Network connectivity"
  fail "  - Version exists: $GMH_VERSION"
  exit 1
fi

chmod +x "$GMH_INSTALL/gmh.tmp"
mv "$GMH_INSTALL/gmh.tmp" "$GMH_INSTALL/gmh"

ok "Installed: $GMH_INSTALL/gmh"

# ---------------------------------------------------------------------------
# 5. Verify install
# ---------------------------------------------------------------------------
info "Verifying install..."
if ! "$GMH_INSTALL/gmh" version >/dev/null 2>&1; then
  warn "gmh is installed but 'version' command failed."
  warn "Try: $GMH_INSTALL/gmh version"
fi

# ---------------------------------------------------------------------------
# 6. Add to PATH (instructions)
# ---------------------------------------------------------------------------
echo
ok "gmh is installed!"
echo
echo -e "${BOLD}Next steps:${NC}"
echo
echo -e "  1. Add gmh to your PATH (one-time):"
echo
SHELL_NAME=$(basename "${SHELL:-/bin/bash}" 2>/dev/null || echo "bash")
case "$SHELL_NAME" in
  fish)
    echo -e "     ${CYAN}fish_add_path $GMH_INSTALL${NC}"
    ;;
  zsh)
    echo -e "     ${CYAN}echo 'export PATH=\"\$PATH:$GMH_INSTALL\"' >> ~/.zshrc${NC}"
    echo -e "     ${CYAN}source ~/.zshrc${NC}"
    ;;
  bash|sh)
    echo -e "     ${CYAN}echo 'export PATH=\"\$PATH:$GMH_INSTALL\"' >> ~/.bashrc${NC}"
    echo -e "     ${CYAN}source ~/.bashrc${NC}"
    ;;
  *)
    echo -e "     ${CYAN}export PATH=\"\$PATH:$GMH_INSTALL\"${NC}"
    ;;
esac
echo
echo -e "  2. Verify: ${CYAN}gmh version${NC}"
echo
echo -e "  3. Use: ${CYAN}cd your-project && gmh install${NC}"
echo
echo -e "${BOLD}Docs:${NC} https://github.com/$GMH_REPO/blob/main/docs/CLI.md"
echo
