# Sensor 09 — Verify-after-build

> **Objetivo:** impedir que o `team-manager` rotule uma issue como
> `in-review` ou peça validação humana baseado **apenas** no
> auto-relato do builder. **Quem roda:** `team-manager` (NÃO o
> builder, NÃO o QA). **Quando:** entre o build terminar e o
> handoff para QA. **Falha → ação:** reabrir a sub-issue com o
> builder, listando as inconsistências encontradas.

---

## Por que este sensor existe

**Lição do Mandaí v2 (jul/2026, ADR-0014):** durante a construção
do PR #5 do Mandaí v2, **2 subagentes mentiram (ou erraram por
alucinação) ao reportar estado verde**:

1. **Builder backend** reportou "`go.mod` está em `go 1.22.0`"
   quando o arquivo continha `go 1.25.0`. O humano (Brenon) só
   descobriu lendo o arquivo diretamente — `grep go.mod backend/`
   mostrou `1.25.0`.

2. **Builder infra** disse "CMD `${DATABASE_URL}` está expandindo
   corretamente" — mas o compose roda o `command:` em **exec form**
   (sem shell), então `${DATABASE_URL}` é passado **literal** ao
   container, que interpreta como variável de ambiente. Funcionou
   por sorte (o container tinha `DATABASE_URL` setado), mas estava
   frágil.

3. **Builder lint** disse "57 issues restantes" e o humano teve
   que rodar `golangci-lint run` ele mesmo para confirmar.

**Resultado:** o framework ganhou 5 defeitos **passáveis** em
produção porque o team-manager aceitou auto-relato. A correção
levou **+6 horas de debugging manual** que poderiam ter sido
pegas em 5 minutos com este sensor.

**Princípio:** auto-relato de subagente é **evidência fraca**. O
team-manager é o **único** responsável pelo claim de "verde" e
**deve verificar independentemente** antes de propagar.

---

## O que verificar (checklist de 6 itens)

> Rode **todos** estes 6 itens. Marque ✅ ou ❌. Se qualquer um for
> ❌, NÃO mova para `in-review` — devolva ao builder com a lista
> das inconsistências.

### 1. **Re-ler os arquivos de source-of-truth**

```bash
# Go: conferir que a versão bate com o Dockerfile
echo "=== go.mod (go directive) ==="
grep -E "^go " backend/go.mod

echo "=== Dockerfile (Go base image) ==="
grep -E "FROM golang:" deploy/Dockerfile.backend

# Devem bater (major.minor)
```

```bash
# Node: conferir engines vs Dockerfile vs CI
echo "=== package.json (engines.node) ==="
grep '"node":' web/package.json

echo "=== Dockerfile (Node base image) ==="
grep -E "FROM node:" web/Dockerfile

echo "=== CI Node version ==="
grep "NODE_VERSION:" .github/workflows/ci.yml
```

```bash
# Compose: conferir migrate (oficial) e distroless
echo "=== Compose migrate ==="
grep -E "image:.*migrate/migrate" deploy/docker-compose.yml

echo "=== Distroless ==="
grep -E "distroless/(static|base|nodejs)" deploy/Dockerfile.backend web/Dockerfile
```

### 2. **Re-rodar check-stack-versions.sh (offline)**

```bash
./harness/scripts/check-stack-versions.sh
# Deve sair com 0 (ou só WARNs, nunca FAILs)
```

> **Esperado:** 0 fails, 0 critical warns. Se houver fail, é bug
> que o builder deixou passar.

### 3. **Re-rodar localmente os 3 comandos canônicos**

```bash
# Backend (Go)
cd backend
make lint    # golangci-lint run, 0 issues
make test    # go test -race -cover, coverage ≥ 80% no escopo correto
make vuln    # govulncheck, 0 vulnerabilities
cd ..
```

```bash
# Frontend (Node)
cd web
pnpm lint
pnpm typecheck
pnpm test:run
pnpm audit --audit-level=high  # 0 vulnerabilities
cd ..
```

> **Esperado:** todos exit code 0. Se algum falhar, é regressão que
> o builder não pegou.

### 4. **Conferir CI local do PR (não confiar em "CI passou")**

```bash
gh pr checks <PR_NUMBER>
# Esperado: todos os jobs PASS. Se algum FAIL, esperar (não mover
# para in-review antes de verde).
```

### 5. **Conferir o PR template**

```bash
gh pr view <PR_NUMBER> --json body | jq -r '.body' | grep -E "Como testar|## Sensors|## Changes"
```

- [ ] "Como testar localmente" presente e **executável** (comandos
  reais, não placeholders).
- [ ] Bloco "Sensors" com checkboxes marcados.
- [ ] "Changes" lista arquivos reais, não "...".

### 6. **Conferir cobertura no escopo correto**

```bash
# Coverage medida COM -coverpkg (não diluída)
grep -E "COVERPKG.*=" backend/Makefile
# Deve ter algo como: COVERPKG := ./internal/app/...,./internal/handler/...

# Coverage report
cd backend && go tool cover -func=coverage.out | tail -3
cd ..
# Esperado: "total: 90.0%" (não 47% — esse é o sintoma de COVERPKG errado)
```

---

## Saída esperada (template de comentário na issue)

```markdown
🤖 **team-manager — verify-after-build (sensor 09)**

**Sub-issue:** #<id> (PR #<pr>)
**Builder reportou:** "PRONTO — CI verde, 0 issues, 92% coverage"
**Verificação independente:**

- [x] go.mod `go 1.25.0` vs Dockerfile `golang:1.25-alpine` — bate ✅
- [x] package.json node 22 vs Dockerfile node:22-alpine vs CI NODE_VERSION 22 — bate ✅
- [x] distroless static-debian13:nonroot — correto ✅
- [x] migrate/migrate oficial (sem custom builder) — correto ✅
- [x] `make lint` → 0 issues ✅
- [x] `make test` → coverage 92.0% (com -coverpkg correto) ✅
- [x] `make vuln` → 0 vulnerabilities ✅
- [x] `gh pr checks` → 7/7 PASS ✅
- [x] PR template preenchido (Como testar, Sensors, Changes) ✅

**Resultado:** ✅ VERIFICADO. Movendo para `in-review` → @quality-assurance.
```

**Se algum item for ❌:**

```markdown
🤖 **team-manager — verify-after-build (sensor 09)**

**Sub-issue:** #<id> (PR #<pr>)
**Builder reportou:** "PRONTO — CI verde, 0 issues, 92% coverage"
**Verificação independente:**

- [x] go.mod bate com Dockerfile ✅
- [ ] **`make test` coverage 47.8% (NÃO 92%)** ❌
  - Esperado: `total: 90%+` com `-coverpkg=./internal/app/...`
  - Real: `total: 47.8%` (coverage diluída em main, generated)
  - Fix: ajustar `COVERPKG` no `backend/Makefile`
- [x] Resto OK

**Resultado:** ❌ DIVERGÊNCIA ENCONTRADA. **Não** movendo para
`in-review`. Devolvendo ao @backend-engineer com lista acima.
```

---

## Exit codes

| Code | Significado                                            |
|------|--------------------------------------------------------|
| 0    | Todos os 6 itens passaram — pode mover para `in-review`|
| 1    | Pelo menos 1 divergência — devolver ao builder         |

---

## Quem roda / quando

| Quem            | Quando                                          |
|-----------------|-------------------------------------------------|
| `team-manager`  | **Sempre**, entre `in-progress` e `in-review`. **NÃO** delegado a builder ou QA. |
| Não é builder   | Builder não verifica o próprio trabalho (viés). |
| Não é QA        | QA roda os sensores (00-08) depois desta etapa.|

---

## Integração com o workflow

```yaml
build_done:
  -> team-manager: roda SENSOR 09 (verify-after-build)
  # Se passou:
  -> team-manager: aplica label `in-review`
  -> team-manager: comenta briefing para @quality-assurance
  # Se falhou:
  -> team-manager: aplica label `in-progress` (reverte)
  -> team-manager: posta divergências como comentário
  -> team-manager: cutuca @builder
```

> **Quem delega:** o `team-manager` **não delega** este sensor. O
> ele mesmo roda. Builder que reporta "PRONTO" é suspeito até ser
> verificado.

---

## Anti-patterns (NÃO faça)

| ❌ Errado                                                   | ✅ Certo                                                |
|------------------------------------------------------------|----------------------------------------------------------|
| Aceitar "CI verde" do builder sem rodar `gh pr checks`     | Rodar `gh pr checks <id>` você mesmo                    |
| Confiar em "92% coverage" do builder sem ver o report      | Rodar `go tool cover -func=coverage.out` e ler o total   |
| Mover para `in-review` antes de rodar `make test`          | Rodar `make test` você mesmo, exit 0                    |
| Confiar em "0 issues lint" sem ver o output                | Rodar `golangci-lint run` você mesmo                    |
| Pular o sensor 09 "porque CI já passou"                    | Sensor 09 é o complemento do CI — pega o que CI não pega (ex.: -coverpkg errado, compose literal vs escape) |

---

## Métrica de sucesso

Após 1 mês usando este sensor, medir:

- **% de sub-issues que precisaram reabrir após o sensor 09.**
  Meta: < 5%. Se > 20%, o builder está reportando verde
  prematuramente e o team-manager precisa cutucar mais.
- **% de defeitos em produção que o sensor 09 teria pego.**
  Meta: ≥ 80% (defeitos óbvios, não arquiteturais).

---

## Referências

- ADR-0014 (este sensor é a operacionalização da decisão)
- Invariante 19 do `AGENTS.md`
- §11 do `personas/team-manager.md`
- `harness/scripts/check-stack-versions.sh` (seções 1-15)
