# ============================================================================
# GMH (git-meta-harness) CLI — Bootstrap Installer (Windows PowerShell)
# ============================================================================
# Installs the gmh binary by downloading the latest release artifact
# for the current OS/arch (Windows amd64).
#
# Usage:
#   iwr -useb https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.ps1 | iex
#
# Environment variables:
#   $env:GMH_INSTALL  — install dir (default: $HOME\.gmh\bin)
#   $env:GMH_VERSION  — specific version to install (default: latest)
# ============================================================================

$ErrorActionPreference = "Stop"

# Defaults
if (-not $env:GMH_INSTALL) { $env:GMH_INSTALL = Join-Path $HOME ".gmh\bin" }
if (-not $env:GMH_VERSION) { $env:GMH_VERSION = "latest" }
$Repo = if ($env:GMH_REPO) { $env:GMH_REPO } else { "brenonaraujo/git-meta-harness" }

# Colors
function Info($msg) { Write-Host "==> $msg" -ForegroundColor Cyan }
function Ok($msg)   { Write-Host "✅ $msg" -ForegroundColor Green }
function Warn($msg) { Write-Host "⚠️  $msg" -ForegroundColor Yellow }
function Fail($msg) { Write-Host "❌ $msg" -ForegroundColor Red; exit 1 }

# ---------------------------------------------------------------------------
# 1. Detect Arch
# ---------------------------------------------------------------------------
Info "Detecting platform..."
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { Fail "32-bit not supported" }
$OS = "windows"
$BinName = "gmh-$OS-$Arch.exe"

Ok "Platform: $OS/$Arch"

# ---------------------------------------------------------------------------
# 2. Resolve version
# ---------------------------------------------------------------------------
if ($env:GMH_VERSION -eq "latest") {
  Info "Resolving latest version..."
  $ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"
  try {
    $Response = Invoke-RestMethod -Uri $ApiUrl -TimeoutSec 10
    $env:GMH_VERSION = $Response.tag_name
  } catch {
    Fail "Could not resolve latest version. Set `$env:GMH_VERSION explicitly."
  }
}
Ok "Version: $env:GMH_VERSION"

# ---------------------------------------------------------------------------
# 3. Create install dir
# ---------------------------------------------------------------------------
Info "Install dir: $env:GMH_INSTALL"
New-Item -ItemType Directory -Force -Path $env:GMH_INSTALL | Out-Null

# ---------------------------------------------------------------------------
# 4. Download binary
# ---------------------------------------------------------------------------
$DownloadUrl = "https://github.com/$Repo/releases/download/$env:GMH_VERSION/$BinName"
Info "Downloading: $DownloadUrl"
$BinPath = Join-Path $env:GMH_INSTALL "gmh.exe"

try {
  Invoke-WebRequest -Uri $DownloadUrl -OutFile $BinPath -UseBasicParsing -TimeoutSec 60
} catch {
  Fail "Download failed. Check: $DownloadUrl"
}

Ok "Installed: $BinPath"

# ---------------------------------------------------------------------------
# 5. Verify
# ---------------------------------------------------------------------------
Info "Verifying install..."
try {
  & $BinPath version | Out-Null
} catch {
  Warn "gmh is installed but 'version' command failed. Try: $BinPath version"
}

# ---------------------------------------------------------------------------
# 6. PATH instructions
# ---------------------------------------------------------------------------
Ok "gmh is installed!"
Write-Host ""
Write-Host "Next steps:" -ForegroundColor White
Write-Host ""
Write-Host "  1. Add gmh to your PATH (one-time):"
Write-Host ""
Write-Host "     `$env:Path += ';$env:GMH_INSTALL'" -ForegroundColor Cyan
Write-Host ""
Write-Host "     (or add it to your PowerShell profile for permanence)"
Write-Host ""
Write-Host "  2. Verify: gmh version" -ForegroundColor Cyan
Write-Host ""
Write-Host "  3. Use: cd your-project; gmh install" -ForegroundColor Cyan
Write-Host ""
Write-Host "Docs: https://github.com/$Repo/blob/main/docs/CLI.md" -ForegroundColor White
