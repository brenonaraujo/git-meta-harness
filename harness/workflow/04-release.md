# Workflow 04 — Release

> Como uma entrega vira release versionada, com tag, changelog e
> imagem no GHCR. Disparado pelo `devops-engineer` após validação
> do usuário.

---

## Triggers

- **Manual:** `team-manager` (ou `devops-engineer`) dispara workflow
  `release.yml` via `workflow_dispatch` após merge.
- **Automático (alternativa):** workflow roda em todo merge na `main`
  com [`release-please`](https://github.com/googleapis/release-please)
  ou [`semantic-release`](https://semantic-release.com/).

> **Recomendado:** `release-please` — automatiza versionamento +
  changelog + PR de release baseado em Conventional Commits.

---

## Versionamento (Semver)

- **MAJOR** (X.0.0): breaking change (OpenAPI breaking, schema
  breaking, remoção de endpoint).
- **MINOR** (0.Y.0): nova feature backwards-compatible.
- **PATCH** (0.0.Z): bugfix.

Prefixe com `v`: `v0.4.0`, `v1.0.0`, `v0.4.1`.

---

## Conventional Commits

Todos os commits no main seguem o padrão:

```
<type>(<scope>): <description> [optional body] [optional footer]
```

| Type        | Quando                                | Bump        |
|-------------|---------------------------------------|-------------|
| `feat:`     | Nova feature                          | MINOR       |
| `fix:`      | Bugfix                                | PATCH       |
| `refactor:` | Mudança interna sem behavior change   | -           |
| `perf:`     | Melhoria de performance               | PATCH       |
| `docs:`     | Só docs                               | -           |
| `test:`     | Só testes                             | -           |
| `ci:`       | Só CI                                 | -           |
| `chore:`    | Manutenção                            | -           |
| `feat!:`    | Breaking change                       | MAJOR       |
| `fix!:`     | Breaking change                       | MAJOR       |

> `!` após type/scope indica breaking change. Adicione `BREAKING CHANGE:`
> no body.

---

## Fluxo de release

### Opção A — Automático com release-please

1. `release-please` analisa os commits desde o último release.
2. Cria/atualiza um PR "Release vX.Y.Z" com changelog.
3. Ao mergear o PR de release, cria a tag e roda `release.yml`.
4. `release.yml` faz build da imagem + push para GHCR + cria GitHub
   Release.

### Opção B — Manual

1. `team-manager` (ou `devops-engineer`) roda `gh workflow run release.yml
   -f version=v0.4.0`.
2. Workflow valida que a versão não existe; cria tag; builda imagem;
   publica no GHCR; cria GitHub Release.

---

## Workflow `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      version:
        description: "Version (e.g. v0.4.0)"
        required: true

permissions:
  contents: write
  packages: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
          token: ${{ secrets.GITHUB_TOKEN }}

      - name: Set up Go
        uses: actions/setup-go@v5
        with: { go-version: '1.22' }

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract version
        id: version
        run: |
          if [ "${{ github.event_name }}" = "workflow_dispatch" ]; then
            VERSION="${{ inputs.version }}"
          else
            VERSION=$(git describe --tags --abbrev=0)
          fi
          echo "version=$VERSION" >> $GITHUB_OUTPUT
          echo "sha_short=$(git rev-parse --short HEAD)" >> $GITHUB_OUTPUT

      - name: Build and push image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: |
            ghcr.io/${{ github.repository_owner }}/${{ github.event.repository.name }}:${{ steps.version.outputs.version }}
            ghcr.io/${{ github.repository_owner }}/${{ github.event.repository.name }}:latest
            ghcr.io/${{ github.repository_owner }}/${{ github.event.repository.name }}:sha-${{ steps.version.outputs.sha_short }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ steps.version.outputs.version }}
          generate_release_notes: true
          files: |
            coverage.html
```

---

## GitHub Release notes

Geradas automaticamente pelo GitHub (ou por `release-please`).
Cada release inclui:

- Lista de PRs merged desde a anterior.
- Lista de autores.
- Lista de issues fechadas.
- Mudanças breaking (destacadas).

---

## Smoke pós-release

Antes de marcar release como sucesso, `devops-engineer` (ou
`quality-assurance`) roda smoke em staging (se aplicável).

---

## Rollback

Se um release causar problema:

1. **Reverter merge** (botão "Revert" no GitHub).
2. **Criar tag de patch** com o revert (ex.: `v0.4.1`).
3. **Imagem** da tag anterior continua disponível no GHCR
   (`ghcr.io/.../service:v0.3.0`).
4. **Avisar** via issue + comentário no PR original.

---

## Quando NÃO fazer release

- ❌ Branch ainda aberta.
- ❌ Sensores não passaram.
- ❌ QA não aprovou.
- ❌ Usuário não validou.
- ❌ Sem changelog (release-please resolve; manual precisa escrever).
