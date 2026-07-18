# Sensor 03 — Contract Tests (OpenAPI)

> **Objetivo:** garantir que o serviço implementado **bate** com o
> contrato OpenAPI declarado, e que mudanças no OpenAPI não quebram
> consumidores.
> **Quando roda:** CI (após unit, em todo push/PR).
> **Falha → ação:** **bloqueia merge** se openapi-diff indicar breaking
> change sem waiver.

---

## Estratégia

**Spec-first (padrão meta-harness):**

1. O `solutions-architect` edita `api/openapi.yaml` **antes** de o
   `backend-engineer` começar a codificar.
2. O `backend-engineer` regenera os tipos/servidor com
   `oapi-codegen` (ou `ogen`).
3. O serviço é implementado contra o contrato.
4. O **contract test** valida em runtime que o serviço responde
   conforme o OpenAPI.

---

## Comandos exatos

### Spec-first: gerar código a partir do OpenAPI

```bash
# Instalar
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Regenerar (Make target: make oas)
oapi-codegen --config=codegen.yaml api/openapi.yaml
```

`codegen.yaml`:

```yaml
package: api
output: internal/api/openapi.gen.go
generate:
  models: true
  gin-server: true
  embedded-spec: true
```

### Diff de OpenAPI entre PRs

```bash
# Ferramentas:
# - openapi-diff (https://github.com/OpenAPITools/openapi-diff)
# - oasdiff (https://github.com/oasdiff/oasdiff) — recomendado

oasdiff diff base/main feat/feature-42 --format json
```

### Validação em runtime (consumer-driven)

```bash
# Schemathesis (Python) — descobre rotas e valida contra o schema
pip install schemathesis
schemathesis run http://localhost:8080/api/v1 --schema api/openapi.yaml

# Dredd (Node) — valida cada endpoint
npx dredd api/openapi.yaml http://localhost:8080
```

---

## Thresholds

| Tipo de mudança      | Bloqueia? | Waiver?                            |
|----------------------|-----------|------------------------------------|
| **Breaking** (remove campo, renomeia, muda tipo) | ✅ sim | só com ADR + major version bump   |
| Adição de campo opcional | ❌ não  | OK                                |
| Adição de endpoint   | ❌ não    | OK                                |
| Mudança de exemplo   | ❌ não    | OK                                |

> **Breaking change** = qualquer mudança que invalide um cliente
> existente. Detalhes: <https://docs.oasdiff.com/>

---

## Onde pluga no pipeline

### CI (`.github/workflows/ci.yml`)

```yaml
contract:
  name: OpenAPI contract
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0
    - uses: actions/setup-go@v5
      with: { go-version: '1.22' }
    - name: Install oapi-codegen
      run: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

    - name: Regenerate and check diff
      run: |
        oapi-codegen --config=codegen.yaml api/openapi.yaml
        git diff --exit-code internal/api/openapi.gen.go

    - name: oasdiff
      uses: oasdiff/oasdiff-action@v1
      with:
        base: origin/main
        revision: HEAD
        fail-on: breaking

    - name: schemathesis (against running service)
      run: |
        docker compose -f deploy/docker-compose.yml up -d
        sleep 10
        pip install schemathesis
        schemathesis run http://localhost:8080/api/v1 \
          --schema api/openapi.yaml \
          --checks all
        docker compose -f deploy/docker-compose.yml down
```

---

## Falha típica & remediação

| Falha                                    | Como corrigir                                         |
|------------------------------------------|-------------------------------------------------------|
| `internal/api/openapi.gen.go` mudou      | Commitar a regeneração (intencional).                 |
| Breaking change em endpoint existente    | Adicionar novo endpoint, manter antigo, deprecate.   |
| Schemathesis encontrou 4xx/5xx           | Investigar log; provavelmente bug real.              |
| Response não bate com schema             | Corrigir handler ou ajustar schema (com approval).   |

---

## Quem roda

- **Local:** `backend-engineer` antes de commitar (`make oas`).
- **CI:** workflow `ci.yml` (`contract` job).
- **Falha:** bloqueia merge (a menos que seja mudança intencional e
  regenerada).
