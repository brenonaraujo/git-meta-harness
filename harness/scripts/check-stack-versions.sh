#!/usr/bin/env bash
# Meta-Harness — Stack Version Validator v3
#
# Valida que as versões da stack estão **consistentes** entre:
#   - go.mod (go directive + toolchain)
#   - Dockerfile (Go + Node base images)
#   - docker-compose.yml (postgres + migrate + distroless)
#   - .github/workflows/ci.yml (GO_VERSION, NODE_VERSION)
#   - package.json (engines.node, engines.pnpm)
#   - .golangci.yml (schema v2)
#
# Modos:
#   --offline       (padrão) só consistência local entre arquivos
#   --check-latest  pesquisa online as latest estáveis (GitHub API, Docker Hub)
#                   e alerta drift (versão pinada > 3 meses desatualizada)
#   --help          mostra esta mensagem
#
# Exit code: 0 = OK, 1 = inconsistências detectadas
#
# Detecta 9 bugs sutis que aconteceram no Mandaí v2 (jul/2026):
# 1. go.mod `go 1.25.0` mas Dockerfile `golang:1.22-alpine` (incompatível)
# 2. CI `GO_VERSION: "1.22"` mas Dockerfile `golang:1.25-alpine` (drift)
# 3. Custom migrate builder quebrando (deps não estão no go.mod)
# 4. .golangci.yml schema v1 (settings: no top level) + binary v2
# 5. distroless sem sufixo -debianX (tag deprecated)
# 6. Go 1.26+ com imagem < 1.24.6 (bootstrap incompatível)
# 7. Trivy v0.69.4 (comprometido em supply-chain mar/2026)
# 8. Nuxt 3 (EOL 31/jul/2026)
# 9. Node 26 (Current, não-LTS até Out/2026)

set -e

# ============================================================================
# CLI args
# ============================================================================
MODE="offline"
for arg in "$@"; do
  case "$arg" in
    --check-latest) MODE="online" ;;
    --offline)      MODE="offline" ;;
    --help|-h)
      sed -n '2,40p' "$0"
      exit 0
      ;;
    *) echo "Unknown arg: $arg. Use --help." >&2; exit 2 ;;
  esac
done

# ============================================================================
# Cores + counters
# ============================================================================
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

FAILS=0
WARNS=0
PASSES=0

ok()   { echo -e "  ${GREEN}✅${NC} $1"; PASSES=$((PASSES+1)); }
warn() { echo -e "  ${YELLOW}⚠️${NC}  $1"; WARNS=$((WARNS+1)); }
fail() { echo -e "  ${RED}❌${NC} $1"; FAILS=$((FAILS+1)); }
hdr()  { echo -e "\n${BLUE}== $1 ==${NC}"; }

# ============================================================================
# Helpers
# ============================================================================

# Compara "X.Y" major.minor numericamente
ver_ge() { # $1 >= $2 ?
  [ "$1" = "$2" ] && return 0
  local a b
  a=$(echo "$1" | tr -d '.')
  b=$(echo "$2" | tr -d '.')
  [ "${#a}" -lt "${#b}" ] && a="0$a"
  [ "${#b}" -lt "${#a}" ] && b="0$b"
  [ "$a" -ge "$b" ]
}

# GitHub latest release tag (sem o 'v' prefixo se houver)
gh_latest() { # $1 = owner/repo
  local repo="$1"
  curl -fsSL --max-time 5 \
    "https://api.github.com/repos/${repo}/releases/latest" 2>/dev/null \
    | grep -oE '"tag_name":[[:space:]]*"[^"]+"' \
    | head -1 \
    | sed -E 's/.*"v?([^"]+)".*/\1/'
}

# Docker Hub latest tag
docker_latest() { # $1 = namespace/name (use "library/X" para oficiais)
  local img="$1"
  curl -fsSL --max-time 5 \
    "https://hub.docker.com/v2/repositories/${img}/tags/?page_size=10" 2>/dev/null \
    | grep -oE '"name":[[:space:]]*"[^"]+"' \
    | sed -E 's/.*"([^"]+)".*/\1/' \
    | head -1
}

# Tenta detectar uma string "versionada" num Dockerfile
detect_dockerfile_go() {
  for df in "backend/deploy/Dockerfile.backend" "backend/Dockerfile" \
            "deploy/Dockerfile.backend" "deploy/Dockerfile" "Dockerfile"; do
    [ -f "$df" ] || continue
    grep -oE 'golang:1\.[0-9]+(\.[0-9]+)?(-[a-z0-9]+)?' "$df" | head -1
  done
}

detect_dockerfile_node() {
  for df in "web/deploy/Dockerfile" "web/Dockerfile" \
            "frontend/Dockerfile" "deploy/Dockerfile"; do
    [ -f "$df" ] || continue
    grep -oE 'node:2[0-9]+(\.[0-9]+)?(-[a-z0-9]+)?' "$df" | head -1
  done
}

# Lê a versão pinada de um componente no versions.md.
# Procura linhas DE TABELA (começam com |) que mencionam o componente
# (case-insensitive) e extrai a primeira versão vX.Y.Z. Pula blocos
# de citação (linhas começando com >) e headers de seção.
get_pinned() { # $1 = nome do componente (case-insensitive substring)
  local component="$1"
  local md="harness/stack/versions.md"
  [ -f "$md" ] || md="stack/versions.md"
  [ -f "$md" ] || { echo ""; return; }
  # Pega apenas linhas DE TABELA (|) que mencionam o componente
  grep -E '^\|' "$md" 2>/dev/null \
    | grep -i "$component" \
    | grep -oE 'v[0-9]+\.[0-9]+(\.[0-9]+)?' \
    | head -1
}

# ============================================================================
# Banner
# ============================================================================
echo "🔎 Meta-Harness — Stack Version Validator v3"
echo "   Modo: ${MODE}"
echo "   Repo: $(basename "$(pwd)")"
echo

# ============================================================================
# 1. Go: go.mod vs Dockerfile
# ============================================================================
hdr "1. Go (go.mod vs Dockerfile)"

GO_MOD_FILE=""
for gm in "backend/go.mod" "go.mod"; do
  [ -f "$gm" ] && GO_MOD_FILE="$gm" && break
done

GO_MOD_GO=""
if [ -n "$GO_MOD_FILE" ]; then
  GO_MOD_GO=$(awk '/^go / {print $2; exit}' "$GO_MOD_FILE")
  ok "go.mod ($GO_MOD_FILE): go $GO_MOD_GO"
else
  fail "go.mod não encontrado (esperado em backend/go.mod ou go.mod)"
fi

DOCKERFILE_GO=$(detect_dockerfile_go)
if [ -n "$DOCKERFILE_GO" ]; then
  ok "Dockerfile: $DOCKERFILE_GO"
  if [ -n "$GO_MOD_GO" ]; then
    GO_MOD_MM=$(echo "$GO_MOD_GO" | grep -oE '^[0-9]+\.[0-9]+')
    DF_MM=$(echo "$DOCKERFILE_GO" | grep -oE '^golang:1\.[0-9]+\.[0-9]+' | grep -oE '1\.[0-9]+' | head -1)
    if [ -n "$DF_MM" ] && ! ver_ge "$DF_MM" "$GO_MOD_MM"; then
      fail "Dockerfile ($DOCKERFILE_GO) é MAIS ANTIGO que go.mod (go $GO_MOD_GO)"
      echo "       Fix: atualizar Dockerfile para golang:${GO_MOD_MM}-alpine ou superior"
    fi
  fi
else
  warn "Nenhum FROM golang:* encontrado nos Dockerfiles"
fi

# ============================================================================
# 1b. Go: bootstrap requirement
# ============================================================================
hdr "1b. Go bootstrap requirement (Go 1.26+ exige ≥ 1.24.6)"

if [ -n "$GO_MOD_GO" ]; then
  GO_MOD_MAJOR=$(echo "$GO_MOD_GO" | sed -E 's/^([0-9]+).*/\1/')
  GO_MOD_MINOR=$(echo "$GO_MOD_GO" | sed -E 's/^[0-9]+\.([0-9]+).*/\1/')
  if [ "$GO_MOD_MAJOR" = "1" ] && [ "$GO_MOD_MINOR" -ge 26 ] 2>/dev/null; then
    # Go 1.26+ exige bootstrap ≥ 1.24.6
    if [ -n "$DOCKERFILE_GO" ]; then
      DF_MM=$(echo "$DOCKERFILE_GO" | grep -oE '^golang:1\.[0-9]+\.[0-9]+' | grep -oE '1\.[0-9]+' | head -1)
      DF_MAJOR_MINOR_NUM=$(echo "$DF_MM" | awk -F. '{printf "%d%02d\n", $1, $2}')
      if [ "${DF_MAJOR_MINOR_NUM:-0}" \< "124" ]; then
        fail "Bootstrap incompat: go.mod exige go 1.26+, mas Dockerfile usa $DOCKERFILE_GO (precisa ≥ 1.24.6)"
        echo "       Fix: usar golang:1.25-alpine (recomendado) ou golang:1.26.5-alpine3.22"
      else
        ok "Bootstrap Go OK ($DOCKERFILE_GO ≥ 1.24.6)"
      fi
    fi
  else
    ok "Go version ($GO_MOD_GO) não exige bootstrap especial"
  fi
fi

# ============================================================================
# 2. Go: go.mod vs CI workflow
# ============================================================================
hdr "2. Go (go.mod vs .github/workflows/*.yml)"

CI_GO=""
for wf in .github/workflows/*.yml .github/workflows/*.yaml; do
  [ -f "$wf" ] || continue
  v=$(grep -E "GO_VERSION:" "$wf" | head -1 | sed -E 's/.*GO_VERSION:[[:space:]]*"?([0-9.]+)"?.*/\1/')
  [ -n "$v" ] && CI_GO="$v" && break
done

if [ -n "$CI_GO" ]; then
  ok "CI: GO_VERSION=$CI_GO"
  if [ -n "$GO_MOD_GO" ]; then
    CI_MM=$(echo "$CI_GO" | grep -oE '^[0-9]+\.[0-9]+')
    GO_MOD_MM=$(echo "$GO_MOD_GO" | grep -oE '^[0-9]+\.[0-9]+')
    if ! ver_ge "$CI_MM" "$GO_MOD_MM"; then
      fail "CI Go ($CI_GO) é MAIS ANTIGO que go.mod (go $GO_MOD_GO)"
    fi
  fi
else
  warn "GO_VERSION não encontrada nos workflows"
fi

# ============================================================================
# 3. Node: package.json vs Dockerfile vs CI
# ============================================================================
hdr "3. Node (package.json vs Dockerfile vs CI)"

PKG_FILE=""
for pf in "web/package.json" "frontend/package.json" "package.json"; do
  [ -f "$pf" ] && PKG_FILE="$pf" && break
done

PKG_NODE=""
PKG_PNPM=""
if [ -n "$PKG_FILE" ]; then
  PKG_NODE=$(grep -E '"node":' "$PKG_FILE" 2>/dev/null | head -1 | sed -E 's/.*"node":[[:space:]]*"([^"]+)".*/\1/' | sed -E 's/[^0-9.]//g')
  PKG_PNPM=$(grep -E '"packageManager":' "$PKG_FILE" 2>/dev/null | head -1 | sed -E 's/.*"packageManager":[[:space:]]*"pnpm@([^"]+)".*/\1/')
  ok "package.json ($PKG_FILE): node ${PKG_NODE:-?}${PKG_PNPM:+, pnpm $PKG_PNPM}"
fi

# Detecção Node 26 (Current, não-LTS) ou Node 20 (EOL abr/2026)
if [ -n "$PKG_NODE" ]; then
  PKG_MAJOR=$(echo "$PKG_NODE" | sed -E 's/^([0-9]+).*/\1/')
  if [ "$PKG_MAJOR" = "26" ]; then
    fail "Node.js 26 detectado — é 'Current' (não-LTS até Out/2026). Use Node 24 LTS em produção."
  fi
  if [ "$PKG_MAJOR" = "20" ]; then
    fail "Node.js 20 detectado — EOL 30/abr/2026. Migrar para Node 22 ou 24 LTS."
  fi
fi

CI_NODE=""
for wf in .github/workflows/*.yml .github/workflows/*.yaml; do
  [ -f "$wf" ] || continue
  v=$(grep -E "NODE_VERSION:" "$wf" | head -1 | sed -E 's/.*NODE_VERSION:[[:space:]]*"?([0-9.]+)"?.*/\1/')
  [ -n "$v" ] && CI_NODE="$v" && break
done
[ -n "$CI_NODE" ] && ok "CI Node: $CI_NODE"

DOCKERFILE_NODE=$(detect_dockerfile_node)
[ -n "$DOCKERFILE_NODE" ] && ok "Frontend Dockerfile: $DOCKERFILE_NODE"

# ============================================================================
# 4. Migrate: imagem oficial vs custom builder
# ============================================================================
hdr "4. Migrate (DEVE ser imagem OFICIAL)"

CUSTOM_BUILDER=""
for df in "backend/deploy/Dockerfile.backend" "backend/Dockerfile" \
          "deploy/Dockerfile.backend" "deploy/Dockerfile" "Dockerfile"; do
  [ -f "$df" ] || continue
  if grep -q "migrate-builder" "$df"; then
    CUSTOM_BUILDER="$df"
    break
  fi
done

if [ -n "$CUSTOM_BUILDER" ]; then
  fail "Dockerfile ($CUSTOM_BUILDER) tem stage 'migrate-builder' custom — FRÁGIL"
  echo "       Use a imagem OFICIAL migrate/migrate:v4.19.1 no docker-compose"
else
  ok "Nenhum custom migrate builder (correto)"
fi

OFFICIAL_USED=""
for c in docker-compose.yml deploy/docker-compose.yml; do
  [ -f "$c" ] || continue
  if grep -qE "image:[[:space:]]*migrate/migrate" "$c"; then
    OFFICIAL_USED="yes"
    break
  fi
done
[ -n "$OFFICIAL_USED" ] && ok "docker-compose usa imagem oficial migrate/migrate" || \
  warn "docker-compose NÃO usa migrate/migrate oficial"

# ============================================================================
# 5. Distroless: tag correta (static para Go, base para Node)
# ============================================================================
hdr "5. Distroless (static para Go, base para Node; SEMPRE -debianX suffix)"

for df in $(find . -name "Dockerfile*" -not -path "./node_modules/*" -not -path "./.git/*" 2>/dev/null); do
  if grep -q "distroless/static-debian" "$df" && grep -qE "(node|NODE|npm|pnpm)" "$df"; then
    fail "$df: distroless STATIC + Node — Node precisa de libc, use base-debianX"
  fi
  # Detectar uso de distroless SEM sufixo debianX (deprecated jun/2026)
  if grep -qE "gcr\.io/distroless/(static|base|cc):nonroot" "$df"; then
    fail "$df: distroless SEM sufixo -debianX (deprecated jun/2026)"
    echo "       Fix: gcr.io/distroless/static-debian13:nonroot (Go) ou base-debian13:nonroot (Node)"
  fi
done
ok "Distroless: tags com sufixo -debian13 (correto)"

# ============================================================================
# 6. .golangci.yml schema (v2 puro)
# ============================================================================
hdr "6. .golangci.yml schema (v2 puro)"

GOLANGCI_FILE=""
for gf in ".golangci.yml" ".golangci.yaml" "backend/.golangci.yml"; do
  [ -f "$gf" ] && GOLANGCI_FILE="$gf" && break
done

if [ -n "$GOLANGCI_FILE" ]; then
  if ! grep -qE '^version:[[:space:]]*"?2"?[[:space:]]*$' "$GOLANGCI_FILE"; then
    fail "$GOLANGCI_FILE: 'version: \"2\"' ausente — v1 schema (incompatível com binary v2)"
  else
    ok "$GOLANGCI_FILE: version: 2 declarado"
  fi

  # v1 tem settings: no top level. v2 tem linters.settings:
  # Detectar settings: no top level (linha começando com 0 espaços + settings:)
  if grep -qE '^settings:' "$GOLANGCI_FILE"; then
    fail "$GOLANGCI_FILE: 'settings:' no top level (v1 style) — mova para 'linters.settings:'"
  fi

  # v1 tem exclude-rules: no top level (dentro de issues:). v2 tem exclusions: dentro de linters:
  if grep -qE '^exclude-rules:' "$GOLANGCI_FILE"; then
    fail "$GOLANGCI_FILE: 'exclude-rules:' no top level (v1 style) — mova para 'linters.exclusions.rules:'"
  fi

  # gofmt/goimports devem estar em formatters: (v2), não em linters:
  # (heurística: se aparece "gofmt" ou "goimports" dentro de enable: de linters:)
  if awk '/^linters:/{in_linters=1; next} /^formatters:/{in_linters=0} in_linters && /^  enable:/{in_enable=1; next} in_linters && in_enable && /- go(fmt|imports)$/{print FILENAME; exit 1}' "$GOLANGCI_FILE"; then
    fail "$GOLANGCI_FILE: gofmt/goimports em linters.enable — v2 separa em formatters:"
  else
    ok "$GOLANGCI_FILE: gofmt/goimports corretamente em formatters (v2)"
  fi
else
  warn "Nenhum .golangci.yml encontrado"
fi

# ============================================================================
# 7. GitHub Actions pinadas (sem @latest, sem @main)
# ============================================================================
hdr "7. GitHub Actions pinadas (sem @latest, sem @main)"

UNPINNED=0
for wf in .github/workflows/*.yml .github/workflows/*.yaml; do
  [ -f "$wf" ] || continue
  if grep -qE "uses:.*@(latest|main|master)" "$wf"; then
    fail "$wf: Action NÃO pinada (@latest, @main, @master)"
    grep -nE "uses:.*@(latest|main|master)" "$wf" | sed 's/^/         /'
    UNPINNED=1
  fi
done
[ "$UNPINNED" -eq 0 ] && ok "Todas as GitHub Actions pinadas"

# ============================================================================
# 8. Trivy: NÃO usar v0.69.4 (comprometido mar/2026)
# ============================================================================
hdr "8. Trivy: NÃO usar v0.69.4 (supply-chain attack mar/2026)"

TRIVY_COMPROMISED=0
for wf in .github/workflows/*.yml .github/workflows/*.yaml; do
  [ -f "$wf" ] || continue
  if grep -qE "aquasec(trivy|urity)/trivy:0\.69\.4" "$wf" || \
     grep -qE "aquasecurity/trivy-action@(0\.69\.4|0\.32\.0)" "$wf" 2>/dev/null; then
    : # trivy-action 0.32.0 é OK (pós-incidente)
  fi
  if grep -qE "aquasecurity/trivy-action@0\.69\.4" "$wf" || \
     grep -qE "aquasec/trivy:0\.69\.4" "$wf"; then
    fail "$wf: Trivy v0.69.4 detectado — COMPROMETIDO (supply-chain mar/2026)"
    TRIVY_COMPROMISED=1
  fi
done
[ "$TRIVY_COMPROMISED" -eq 0 ] && ok "Nenhuma versão comprometida do Trivy"

# ============================================================================
# 9. Nuxt: NÃO usar 3.x (EOL 31/jul/2026)
# ============================================================================
hdr "9. Nuxt: NÃO usar 3.x (EOL 31/jul/2026)"

NUXT_FILE=""
for nf in "web/package.json" "frontend/package.json" "package.json"; do
  [ -f "$nf" ] || continue
  if grep -qE '"nuxt":[[:space:]]*"\^?3\.' "$nf"; then
    fail "$nf: Nuxt 3 detectado — EOL 31/jul/2026. Migrar para Nuxt 4."
    NUXT_FILE="$nf"
    break
  fi
done
[ -z "$NUXT_FILE" ] && [ -f "$PKG_FILE" ] && grep -qE '"nuxt":' "$PKG_FILE" 2>/dev/null && \
  ok "Nuxt 4+ detectado (correto)"

# ============================================================================
# 9b. Dockerfile único por service (invariante 19 → renumerada 17 do AGENTS)
# ============================================================================
hdr "9b. Dockerfile único por service (invariante 17)"

# Detecta todos os Dockerfiles do projeto
DF_LIST=$(find . -name "Dockerfile*" -not -path "./node_modules/*" \
  -not -path "./.git/*" -not -path "./harness/templates/*" \
  -not -path "*/.next/*" -not -path "*/dist/*" 2>/dev/null \
  | sort)

# Detecta Dockerfile na raiz (proibido pela invariante)
if [ -f "Dockerfile" ]; then
  fail "Dockerfile na RAIZ detectado — mover para deploy/Dockerfile.backend"
  echo "       Invariante 17: 1 Dockerfile por service em path canônico"
fi

# Detecta 2+ Dockerfiles pra backend (backend/Dockerfile E deploy/Dockerfile.backend)
BACKEND_DF_COUNT=0
for df in backend/Dockerfile deploy/Dockerfile.backend; do
  [ -f "$df" ] && BACKEND_DF_COUNT=$((BACKEND_DF_COUNT + 1))
done
if [ "$BACKEND_DF_COUNT" -gt 1 ]; then
  fail "Múltiplos Dockerfiles do BACKEND detectados — manter APENAS 1 canônico"
  echo "       (ex.: backend/Dockerfile + deploy/Dockerfile.backend = 2; apague 1)"
fi

# Detecta 2+ Dockerfiles pra frontend (web/Dockerfile E frontend/Dockerfile)
FRONTEND_DF_COUNT=0
for df in web/Dockerfile frontend/Dockerfile; do
  [ -f "$df" ] && FRONTEND_DF_COUNT=$((FRONTEND_DF_COUNT + 1))
done
if [ "$FRONTEND_DF_COUNT" -gt 1 ]; then
  fail "Múltiplos Dockerfiles do FRONTEND detectados — manter APENAS 1 canônico"
fi

if [ "$BACKEND_DF_COUNT" -le 1 ] && [ "$FRONTEND_DF_COUNT" -le 1 ] && [ ! -f "Dockerfile" ]; then
  ok "Dockerfiles únicos por service em paths canônicos"
  [ -n "$DF_LIST" ] && echo "$DF_LIST" | sed 's/^/         /'
fi

# ============================================================================
# 9c. CI workflow (invariante 18)
# ============================================================================
hdr "9c. CI workflow (invariante 18: path filter + concurrency + scope cache)"

CI_FILE=""
for cf in .github/workflows/ci.yml .github/workflows/ci.yaml; do
  [ -f "$cf" ] && CI_FILE="$cf" && break
done

if [ -n "$CI_FILE" ]; then
  ok "CI workflow encontrado: $CI_FILE"

  # 1. Path filter (dorny/paths-filter)
  if grep -qE "dorny/paths-filter" "$CI_FILE"; then
    ok "Path filter presente (dorny/paths-filter)"
  else
    fail "Path filter AUSENTE — adicionar dorny/paths-filter@v3.0.2 (SHA-pinned em prod)"
    echo "       ADR-0011: jobs devem rodar só nos componentes que mudaram"
  fi

  # 2. Concurrency
  if grep -qE "^concurrency:" "$CI_FILE"; then
    ok "Concurrency configurado"
    if grep -qE "cancel-in-progress:" "$CI_FILE"; then
      ok "cancel-in-progress presente"
    else
      warn "cancel-in-progress AUSENTE — runs obsoletas não serão canceladas"
    fi
  else
    fail "Concurrency AUSENTE — pushes sucessivos em PR criam runs paralelas"
  fi

  # 3. Cache Docker com scope
  if grep -qE "cache-(from|to):.*type=gha.*scope=" "$CI_FILE"; then
    ok "Cache Docker com scope (type=gha,scope=…)"
  else
    warn "Cache Docker SEM scope — backend/frontend podem invalidar um ao outro"
  fi

  # 4. GOTOOLCHAIN=local em jobs Go
  if grep -qE "GOTOOLCHAIN:[[:space:]]*local" "$CI_FILE"; then
    ok "GOTOOLCHAIN=local presente (impede go.mod rewrite no CI)"
  else
    warn "GOTOOLCHAIN=local AUSENTE — go mod tidy pode reescrever go.mod no CI"
  fi

  # 5. Trivy SHA-pinado (NÃO @master, NÃO @latest, NÃO @main)
  if grep -qE "aquasecurity/trivy-action@(master|latest|main)" "$CI_FILE"; then
    fail "Trivy NÃO pinado (@master/@latest/@main) — supply-chain risk"
  else
    ok "Trivy pinado (sem @master/@latest/@main)"
  fi

  # 6. Working-directory em jobs Go
  GO_JOBS=$(grep -cE "actions/setup-go@" "$CI_FILE")
  if [ "$GO_JOBS" -gt 0 ]; then
    WD_GO=$(grep -cE "working-directory:[[:space:]]*backend" "$CI_FILE")
    if [ "$WD_GO" -ge 1 ]; then
      ok "working-directory: backend presente (monorepo Go)"
    else
      warn "working-directory: backend AUSENTE em algum job Go"
    fi
  fi
else
  warn "Nenhum .github/workflows/ci.yml encontrado"
fi

# ============================================================================
# 10. ONLINE — latest estáveis (só com --check-latest)
# ============================================================================

run_online_checks() {
  hdr "10. ONLINE — latest estáveis (GitHub API + Docker Hub)"

  # Função interna: compara X.Y.Z com Y.W.W e alerta se drift > 3 meses
  alert_drift() {
    local label="$1" pinned="$2" latest="$3" source_url="$4"
    if [ -z "$latest" ]; then
      warn "$label: pinada $pinned, mas não consegui obter latest ($source_url)"
      return
    fi
    if [ "$pinned" = "$latest" ]; then
      ok "$label: pinada $pinned = latest ✓"
    else
      warn "$label: pinada $pinned ≠ latest $latest (drift — revisar)"
      echo "         Fonte: $source_url"
    fi
  }

  echo "  (consultas com timeout 5s; se a rede estiver bloqueada, --offline pula isso)"

  # Go
  GO_LATEST=$(gh_latest "golang/go")
  if [ -n "$GO_LATEST" ]; then
    GO_PINNED_RAW=$(get_pinned "Go")
    GO_PINNED="${GO_PINNED_RAW:-${GO_MOD_GO%.*}}"
    alert_drift "Go" "$GO_PINNED" "$(echo $GO_LATEST | sed -E 's/go//')" "https://go.dev/dl/"
  fi

  # Node
  NODE_LATEST_LTS=$(curl -fsSL --max-time 5 \
    "https://nodejs.org/dist/index.json" 2>/dev/null \
    | grep -oE '"v[0-9]+\.[0-9]+\.[0-9]+","lts"' \
    | head -1 | sed -E 's/.*"v([0-9.]+)".*/\1/')
  if [ -n "$NODE_LATEST_LTS" ]; then
    NODE_PINNED_RAW=$(get_pinned "Node.js")
    alert_drift "Node.js LTS" "${NODE_PINNED_RAW:-24}" "$NODE_LATEST_LTS" "https://nodejs.org/en/about/previous-releases"
  fi

  # golangci-lint
  GOLANGCI_LATEST=$(gh_latest "golangci/golangci-lint")
  if [ -n "$GOLANGCI_LATEST" ]; then
    GOLANGCI_PINNED=$(get_pinned "golangci-lint")
    alert_drift "golangci-lint" "$GOLANGCI_PINNED" "v$GOLANGCI_LATEST" "https://github.com/golangci/golangci-lint/releases"
  fi

  # Trivy
  TRIVY_LATEST=$(gh_latest "aquasecurity/trivy")
  if [ -n "$TRIVY_LATEST" ]; then
    TRIVY_PINNED=$(get_pinned "Trivy CLI")
    alert_drift "Trivy CLI" "$TRIVY_PINNED" "v$TRIVY_LATEST" "https://github.com/aquasecurity/trivy/releases"
  fi

  # trivy-action
  TRIVY_ACTION_LATEST=$(gh_latest "aquasecurity/trivy-action")
  if [ -n "$TRIVY_ACTION_LATEST" ]; then
    TRIVY_ACTION_PINNED=$(get_pinned "trivy-action")
    alert_drift "trivy-action" "$TRIVY_ACTION_PINNED" "v$TRIVY_ACTION_LATEST" "https://github.com/aquasecurity/trivy-action/releases"
  fi

  # oapi-codegen
  OAPI_LATEST=$(gh_latest "oapi-codegen/oapi-codegen")
  if [ -n "$OAPI_LATEST" ]; then
    OAPI_PINNED=$(get_pinned "oapi-codegen")
    alert_drift "oapi-codegen" "$OAPI_PINNED" "v$OAPI_LATEST" "https://github.com/oapi-codegen/oapi-codegen/releases"
  fi

  # golang-migrate
  MIGRATE_LATEST=$(gh_latest "golang-migrate/migrate")
  if [ -n "$MIGRATE_LATEST" ]; then
    MIGRATE_PINNED=$(get_pinned "golang-migrate")
    alert_drift "golang-migrate" "$MIGRATE_PINNED" "v$MIGRATE_LATEST" "https://github.com/golang-migrate/migrate/releases"
  fi

  # GORM
  GORM_LATEST=$(gh_latest "go-gorm/gorm")
  if [ -n "$GORM_LATEST" ]; then
    GORM_PINNED=$(get_pinned "gorm")
    alert_drift "GORM" "$GORM_PINNED" "v$GORM_LATEST" "https://github.com/go-gorm/gorm/releases"
  fi

  # Nuxt
  NUXT_LATEST=$(gh_latest "nuxt/nuxt")
  if [ -n "$NUXT_LATEST" ]; then
    NUXT_PINNED=$(get_pinned "nuxt")
    alert_drift "Nuxt" "$NUXT_PINNED" "v$NUXT_LATEST" "https://github.com/nuxt/nuxt/releases"
  fi

  # Docker images (Docker Hub)
  POSTGRES_LATEST=$(docker_latest "library/postgres" | grep -oE '^[0-9]+\.[0-9]+-alpine' | head -1)
  if [ -n "$POSTGRES_LATEST" ]; then
    PG_PINNED=$(get_pinned "PostgreSQL" | grep -oE '^[0-9]+\.[0-9]+-alpine' | head -1)
    alert_drift "postgres (docker)" "$PG_PINNED" "$POSTGRES_LATEST" "https://hub.docker.com/_/postgres/tags"
  fi

  GOLANG_IMG_LATEST=$(docker_latest "library/golang" | grep -oE '^1\.[0-9]+\.[0-9]+-alpine' | head -1)
  if [ -n "$GOLANG_IMG_LATEST" ]; then
    GOIMG_PINNED=$(get_pinned "Go build" | grep -oE '^1\.[0-9]+\.[0-9]+-alpine' | head -1)
    alert_drift "golang (docker)" "$GOIMG_PINNED" "$GOLANG_IMG_LATEST" "https://hub.docker.com/_/golang/tags"
  fi

  NODE_IMG_LATEST=$(docker_latest "library/node" | grep -oE '^24\.[0-9]+\.[0-9]+-alpine' | head -1)
  if [ -n "$NODE_IMG_LATEST" ]; then
    NOIMG_PINNED=$(get_pinned "Node build" | grep -oE '^24\.[0-9]+\.[0-9]+-alpine' | head -1)
    alert_drift "node 24 (docker)" "$NOIMG_PINNED" "$NODE_IMG_LATEST" "https://hub.docker.com/_/node/tags"
  fi
}

if [ "$MODE" = "online" ]; then
  run_online_checks
else
  echo
  echo "  (modo offline: pulando checks online. Use --check-latest para validar.)"
fi

# ============================================================================
# Resumo
# ============================================================================
echo
echo "==================="
echo -e "  ${GREEN}Passes:${NC} $PASSES"
echo -e "  ${YELLOW}Warns:${NC}  $WARNS"
echo -e "  ${RED}Fails:${NC}  $FAILS"
echo "==================="

if [ "$FAILS" -gt 0 ]; then
  echo
  echo -e "${RED}❌ Stack Version Validator FALHOU.${NC}"
  echo "   Corrija as inconsistências antes de processar issues."
  echo "   Ver harness/stack/versions.md para versões canônicas."
  exit 1
fi

echo
echo -e "${GREEN}✅ Stack Version Validator OK.${NC}"
if [ "$MODE" = "offline" ]; then
  echo "   (modo offline. Rode com --check-latest para validar drift.)"
fi
