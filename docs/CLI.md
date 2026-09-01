# gmh — git-meta-harness CLI

> **O QUÊ:** documentação da CLI `gmh` (v1.16.0+).
> **Single static binary**, escrita em Go.
>
> **Comece aqui:** `gmh seed` (v1.16.0) — harness + personas +
> CI + issue 0 + Hermes cron loop. Ver [`docs/SEED.md`](./SEED.md).
>
> **POR QUÊ:** o `gmh` é a porta de entrada para adotar o
> meta-harness. Sem ele, o usuário precisa clonar o repo
> manualmente, copiar `harness/`, e rezar para que esteja
> atualizado. Com `gmh install`, o scaffold chega em 1 comando.
>
> **QUEM:** developers que querem adotar o meta-harness num
> projeto novo, ou sincronizar um projeto existente.

---

## 1. Instalação

### 1.1. Linux / macOS

```bash
curl -sSL https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.sh | bash
```

### 1.2. Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.ps1 | iex
```

### 1.3. Versão específica

```bash
# Linux/macOS
GMH_VERSION=v1.6.0 curl -sSL https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.sh | bash

# Windows
$env:GMH_VERSION = "v1.6.0"
iwr -useb https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.ps1 | iex
```

### 1.4. Verificação

```bash
$ gmh version
gmh 1.6.0
  commit: abc1234
  built: 2026-07-18T18:00:00Z
```

---

## 2. Comandos

### 2.1. `gmh install`

Instala o meta-harness num projeto (cria `harness/` na raiz).

```bash
# Em um projeto novo
cd my-new-project
gmh install
# Cria: harness/AGENTS.md, harness/personas/, harness/sensors/, etc.

# Com versão específica
gmh install --to v1.5.0

# Sobrescrevendo harness/ existente
gmh install --force

# Sem rodar doctor após install
gmh install --skip-check
```

**O que faz:**

1. Resolve a versão (latest se `--to` não for passado).
2. Faz download do tarball do meta-harness nessa versão.
3. Extrai `harness/` para a raiz do projeto.
4. Roda `gmh doctor` (a menos que `--skip-check`).
5. Imprime resumo do que foi instalado.

### 2.2. `gmh sync`

Sincroniza o `harness/` local com a última versão (preserva
customizações locais).

```bash
cd my-existing-project
gmh sync
# Atualiza harness/ mas mantém arquivos modificados localmente

# Ver o que mudaria sem aplicar
gmh sync --dry-run
```

**O que faz:**

1. Lê `VERSION` em `harness/..` (raiz do projeto).
2. Resolve a última versão do meta-harness.
3. Faz diff entre `harness/` local e remoto.
4. Aplica mudanças, mas:
   - Arquivos não modificados localmente: sobrescreve.
   - Arquivos modificados localmente: preserva (com warning).
5. Roda `gmh doctor`.

### 2.3. `gmh update`

Atualiza para uma versão específica (ou a latest).

```bash
# Última versão
gmh update

# Versão específica
gmh update --to v1.5.0

# Downgrade (destrutivo)
gmh update --to v1.4.0 --force
```

É um alias de `gmh sync --to <versão>`, com `--force` para
permitir downgrades.

### 2.4. `gmh doctor`

Roda health checks no projeto local.

```bash
# Modo normal (só mostra falhas)
gmh doctor

# Verbose (mostra tudo, inclusive passes)
gmh doctor --verbose

# Auto-fix
gmh doctor --fix
```

**Checks (15+):**

- `harness/` directory exists
- 9 arquivos críticos presentes (`AGENTS.md`, `bootstrap.md`, etc.)
- 19 invariantes no `AGENTS.md` (v1.5.0+)
- 10 sensors (00-09)
- ADR-0014 (verify-after-build) presente
- Domain-experts especializados (≥ 1)
- Sem `domain-expert.md` genérico (violaria invariante 12)
- Smart routing documentado
- `check-stack-versions.sh` passa
- GitHub labels `type/*` criadas
- Hermes profiles sem `domain-expert` genérico (se Hermes instalado)

**Exit code:** 0 = tudo OK, 1 = pelo menos 1 falha.

### 2.5. `gmh skills`

Gerencia skills (capacidades atômicas) do projeto.

```bash
# Listar skills instaladas
gmh skills list

# Instalar uma skill do registry
gmh skills install code-graph
gmh skills install i18n
gmh skills install tdd-go

# Remover uma skill
gmh skills remove i18n

# Listar skills disponíveis no registry
gmh skills available
```

**Skills built-in (v1.6.0):**

- `code-graph` — uso de code graph em vez de grep+ls+read
- `i18n` — paridade en/pt-BR/es
- `tdd-go` — TDD com table-driven tests em Go
- `twelve-factor` — checklist 12-factor
- `openapi-spec-first` — OpenAPI spec-first workflow
- `github-issues` — uso de `gh issue`
- `github-pr-workflow` — PR template + checks
- `github-code-review` — code review com `gh pr`

### 2.6. `gmh personas`

Gerencia personas (especialmente domain-experts).

```bash
# Listar personas instaladas
gmh personas list

# Criar um domain-expert-<domínio> a partir do template
gmh personas create --domain banking
gmh personas create --domain retail

# Remover um domain-expert
gmh personas remove domain-expert-banking
```

**Por que `personas create` é importante:**

O invariante 12 diz "domain-expert é SEMPRE especializado".
O usuário pode esquecer de criar o specialist ou renomear
errado. `gmh personas create --domain X` faz o trabalho
mecânico (copia template, renomeia, ajusta invariantes).

### 2.7. `gmh plugins`

Gerencia plugins que estendem a própria CLI gmh.

```bash
# Listar plugins instalados
gmh plugins list

# Instalar plugin
gmh plugins install my-plugin

# Remover plugin
gmh plugins remove my-plugin
```

**Nota:** esta feature é **experimental**. A API de plugins
não é estável ainda. Ver [ADR-0016](../harness/contrib/design-decisions.md).

### 2.8. `gmh version`

```bash
$ gmh version
gmh 1.16.0
```

### 2.9. `gmh seed` (v1.16.0)

Materializa o projeto de ponta a ponta. Ver [`SEED.md`](./SEED.md).

```bash
gmh seed booking-saas --describe "SaaS de agendamento online" --github acme/booking-saas
gmh seed booking-saas --spec SPEC.md --from ./harness --no-cron
```

### 2.10. `gmh loop` (v1.16.0)

```bash
gmh loop install          # scripts + hermes cron (no monitor)
gmh loop doctor           # sensor 14
gmh loop status
```

---

## 3. Flags globais

| Flag            | Descrição                                               |
|-----------------|---------------------------------------------------------|
| `-C, --cwd DIR` | Diretório de trabalho (default: `.`)                    |
| `--source REPO` | Repositório fonte (default: `brenonaraujo/git-meta-harness`) |
| `--dry-run`     | Não aplica mudanças; só mostra                          |
| `-v, --verbose` | Output verboso                                          |

Variáveis de ambiente equivalentes: `GMH_CWD`, `GMH_SOURCE`,
`GMH_DRY_RUN`, `GMH_VERBOSE`.

---

## 4. Workflow típico (greenfield)

```bash
# 1. Criar projeto novo
mkdir my-app && cd my-app
git init

# 2. Instalar meta-harness
gmh install

# 3. Validar
gmh doctor

# 4. Editar a spec do projeto
#    (a `harness/seed/meta-harness-seed.md` guia como)
$EDITOR harness/seed/meta-harness-seed.md

# 5. Materializar personas + skills específicos
gmh personas create --domain banking
gmh skills install code-graph
gmh skills install i18n

# 6. Commit
git add .
git commit -m "feat: bootstrap with meta-harness"
```

## 5. Workflow típico (projeto existente)

```bash
cd my-existing-project

# 1. Instalar meta-harness
gmh install

# 2. Validar
gmh doctor

# 3. Sync periódico (mensal)
gmh sync
```

## 6. Workflow típico (CI)

```yaml
# .github/workflows/ci.yml
- name: Install gmh
  run: |
    curl -sSL https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.sh | bash
    echo "$HOME/.gmh/bin" >> $GITHUB_PATH

- name: Health check
  run: gmh doctor
```

## 7. Arquitetura

```
git-meta-harness/
├── cli/
│   ├── main.go              # em cmd/root.go
│   ├── go.mod
│   ├── Makefile
│   ├── cmd/                 # cobra commands
│   │   ├── root.go
│   │   ├── install.go
│   │   ├── sync.go
│   │   ├── update.go
│   │   ├── doctor.go
│   │   ├── skills.go
│   │   ├── personas.go
│   │   ├── plugins.go
│   │   └── version.go
│   ├── internal/
│   │   └── harness/         # read/write harness/
│   ├── installer/
│   │   ├── install.sh       # bootstrap (Linux/macOS)
│   │   └── install.ps1      # bootstrap (Windows)
│   ├── testdata/
│   └── README.md
└── .github/workflows/
    └── cli-release.yml      # build + publish on cli-vX.Y.Z tag
```

## 8. Releases

Binários são publicados em
https://github.com/brenonaraujo/git-meta-harness/releases
com tag `cli-vX.Y.Z`.

**Plataformas suportadas:**

- Linux amd64, arm64
- macOS amd64, arm64 (Apple Silicon)
- Windows amd64

## 9. Troubleshooting

### "command not found: gmh" após install

Adicione `$HOME/.gmh/bin` ao seu PATH:

```bash
# bash
echo 'export PATH="$PATH:$HOME/.gmh/bin"' >> ~/.bashrc
source ~/.bashrc

# zsh
echo 'export PATH="$PATH:$HOME/.gmh/bin"' >> ~/.zshrc
source ~/.zshrc

# fish
fish_add_path $HOME/.gmh/bin
```

### "permission denied" ao rodar `gmh` no Linux/macOS

```bash
chmod +x ~/.gmh/bin/gmh
```

### Versão errada após upgrade

```bash
# Reinstale forçando
GMH_VERSION=v1.6.0 curl -sSL .../install.sh | bash
```

### Windows: "running scripts is disabled"

PowerShell bloqueia scripts por padrão. Rode:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

Depois re-tente o install.

## 10. Quem usa este doc

- **Developers** que querem adotar meta-harness.
- **DevOps** que querem sincronizar projetos.
- **CI** que roda `gmh doctor` em PRs.

## 11. Referências

- [cli/README.md](../cli/README.md) — quick start
- [cli/installer/install.sh](../cli/installer/install.sh) — bootstrap
- [cli/installer/install.ps1](../cli/installer/install.ps1) — Windows bootstrap
- [ADR-0016](../harness/contrib/design-decisions.md) — decisão de usar Go + binário único
- [.github/workflows/cli-release.yml](../.github/workflows/cli-release.yml) — pipeline de release
- Inspirado em: AWS CLI v2, gh CLI, kubectl, gcloud
