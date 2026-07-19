# Sensor 00 — Static Analysis (lint + format)

> **Objetivo:** garantir que o código segue o padrão de estilo e
> não tem armadilhas conhecidas antes de chegar no CI.
> **Quando roda:** `pre-commit` (local) + CI (todo push/PR).
> **Falha → ação:** **bloqueia merge**.

---

## Comandos exatos

### Backend (Go)

```bash
golangci-lint run --timeout=5m ./...
```

Com `golangci-lint` v2 (`.golangci.yml`):

```yaml
version: "2"
linters:
  default: none
  enable:
    # core
    - govet
    - errcheck
    - ineffassign
    - unused
    # style
    - gofmt
    - goimports
    - gocritic
    - revive
    - misspell
    - whitespace
    # complexity
    - gocyclo
    - funlen
    - cyclop
    - dupl
    - goconst
    # bug-prevention
    - bodyclose
    - contextcheck
    - copyloopvar
    - errorlint
    - nilerr
    - rowserrcheck
    - sqlclosecheck
    - unparam
    # security
    - gosec
    # tests
    - testifylint
    - thelper
    - paralleltest
  settings:
    funlen:
      # v1.10.0: limit raised from 25 → 35 (recomendado: 25).
      # Decisão documentada em ADR-0020.
      lines: 35
      statements: 30
    gocyclo:
      min-complexity: 15
    gosec:
      severity: medium
    govet:
      enable-all: true
    revive:
      severity: warning
  exclusions:
    rules:
      - path: _test\.go
        linters: [errcheck, gosec]
run:
  timeout: 5m
  tests: true
issues:
  max-issues-per-linter: 0
  max-same-issues: 0
```

### Frontend (Nuxt/TS)

```bash
pnpm lint
pnpm typecheck
pnpm format:check
```

Configuração:
- ESLint (`@nuxt/eslint`)
- Prettier
- TypeScript strict (`"strict": true` em `tsconfig.json`)

---

## Exit codes

- `0` — OK (sem findings).
- `1` — falha (lint ou typecheck encontrou problema).
- `2` — erro de configuração.

---

## Thresholds

| Métrica                | Limite               |
|------------------------|----------------------|
| Linhas por função      | **≤ 35 (max) / ≤ 25 (recomendado)** (`funlen`, v1.10.0) |
| Linhas por arquivo     | **≤ 150** (convenção + script `wc -l` no CI) |
| Complexidade ciclomática | **≤ 15** (`gocyclo`) |
| Duplicação             | **< 5%** (`dupl`)    |
| Erros de tipo TS       | **0**                |

---

## Onde pluga no pipeline

### Local (`pre-commit`)

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    rev: v2.0.0
    hooks:
      - id: golangci-lint
  - repo: local
    hooks:
      - id: frontend-lint
        name: Frontend lint
        entry: pnpm lint
        language: system
        pass_filenames: false
        files: \.(vue|ts|tsx|js)$
```

### CI (`.github/workflows/ci.yml`)

```yaml
lint:
  name: Lint
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with: { go-version: '1.22' }
    - uses: golangci/golangci-lint-action@v7
      with: { version: v2.0.0 }
    - uses: actions/setup-node@v4
      with: { node-version: '20' }
    - run: pnpm install --frozen-lockfile
    - run: pnpm lint
    - run: pnpm typecheck
```

---

## Falha típica & remediação

| Falha                                     | Como corrigir                                          |
|-------------------------------------------|--------------------------------------------------------|
| `Function is too long` (`funlen`)         | v1.10.0: limite 35. Se passou de 25 (recomendado), documente o porquê (skill `pre-implementation-design`). Se passou de 35, refatore OBRIGATORIAMENTE. |
| `Cyclomatic complexity` (`gocyclo`)       | Extrair `if`s em funções nomeadas (early return).      |
| `ineffectual assignment` (`ineffassign`)  | Remover a atribuição ou usar a variável.               |
| `unused`                                  | Deletar ou usar (`_ = x` se for pra silenciar).        |
| `errcheck`                                | Tratar o erro (ou `_, _ = foo()` com justificativa).   |
| `whitespace`                              | Rodar `gofmt` / `prettier --write`.                    |
| `import order` (`goimports`)              | Rodar `goimports -w .`.                                |
| `gosec G404` (math/rand)                  | Usar `crypto/rand` se for segurança.                   |

---

## Quem roda

- **Local:** `backend-engineer`, `frontend-engineer` antes de commitar.
- **CI:** workflow `ci.yml`.
- **Falha no CI:** bloqueia merge (configurar como required check no
  branch protection).
