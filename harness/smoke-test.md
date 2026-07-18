# Smoke Test — Validação do bootstrap do meta-harness

> **O QUÊ:** um script de **pré-flight check** que o `team-manager`
> **DEVE** rodar antes de processar qualquer issue. Se qualquer item
> falhar, o `team-manager` **NÃO** segue e pede ao usuário para
> corrigir.
>
> **Versão do meta-harness:** 1.0.0
>
> **POR QUÊ:** aprendemos com o projeto-piloto **Mandaí v2** (jul/2026)
> que 3 bugs sutis passaram batido:
> 1. team-manager usou `domain-expert` genérico (viola invariante 12).
> 2. team-manager roteou `type/technical` para domain-expert (deveria
>    pular — smart routing).
> 3. Projeto ficou com versão antiga do meta-harness (54 arquivos em
>    vez de 62) — smart routing não funcionou.
>
> O smoke test abaixo **previne** esses 3 bugs.

---

## Como rodar

```bash
# O team-manager roda no início de qualquer materialização:
./harness/scripts/smoke-test.sh
```

> Se qualquer item falhar, **NÃO processe issues** até corrigir.

---

## Checklist completo (12 itens)

### 1. Versão do meta-harness instalada

```bash
# Esperado: 62 arquivos (com todas as correções)
FILE_COUNT=$(find harness/ -type f | wc -l | tr -d ' ')
if [ "$FILE_COUNT" -lt 60 ]; then
  echo "❌ Versão antiga do meta-harness detectada ($FILE_COUNT arquivos)"
  echo "   Esperado: 62 arquivos. Rode: rsync -a meta-harness-m3-code/harness/ ./harness/"
  exit 1
fi
```

### 2. Arquivos críticos presentes

```bash
# Lista de arquivos que DEVEM existir
REQUIRED=(
  "harness/bootstrap.md"
  "harness/AGENTS.md"
  "harness/seed/meta-harness-seed.md"
  "harness/personas/interactions.md"            # matriz de papéis
  "harness/personas/team-manager.md"
  "harness/personas/domain-expert.template.md"
  "harness/personas/solutions-architect.md"
  "harness/personas/backend-engineer.md"
  "harness/personas/frontend-engineer.md"
  "harness/personas/quality-assurance.md"
  "harness/personas/devops-engineer.md"
  "harness/sensors/00-static-analysis.md"
  "harness/sensors/01-vulnerability-scan.md"
  "harness/sensors/02-unit-tests.md"
  "harness/sensors/03-contract-tests.md"
  "harness/sensors/04-image-scan.md"
  "harness/sensors/05-smoke-tests.md"
  "harness/sensors/06-load-tests.md"
  "harness/sensors/07-twelve-factor-audit.md"
  "harness/sensors/08-i18n-audit.md"
  "harness/skills/i18n.md"
  "harness/templates/locales.template.json"
  "harness/contrib/design-decisions.md"
)
```

### 3. Smart routing documentado

```bash
# AGENTS.md deve ter routing rules com `type/`
if ! grep -q "type/technical\|type/infra\|type/feature" harness/AGENTS.md; then
  echo "❌ Smart routing não está documentado em AGENTS.md"
  exit 1
fi
```

### 4. Interações matrix presente

```bash
if [ ! -f "harness/personas/interactions.md" ]; then
  echo "❌ Matriz de interações ausente (harness/personas/interactions.md)"
  exit 1
fi
```

### 5. Labels de tipo criadas no GitHub

```bash
# Esperado: 7 labels type/*
for t in feature technical infra bug tech-debt docs spike; do
  if ! gh label list --repo "$REPO" | grep -q "type/$t"; then
    echo "❌ Label type/$t não existe no repo"
    exit 1
  fi
done
```

### 6. Domain-expert(s) especializado(s) existente(s)

```bash
# Esperado: ≥ 1 domain-expert-<domínio> (com sufixo)
if ! ls harness/personas/domain-expert-*.md 2>/dev/null | grep -v template | grep -q .; then
  echo "❌ Nenhum domain-expert-<domínio> especializado encontrado"
  echo "   Esperado: harness/personas/domain-expert-<seu-dominio>.md"
  echo "   Use harness/personas/domain-expert.template.md como base"
  exit 1
fi
```

### 7. **CRÍTICO: nenhum `domain-expert.md` genérico**

```bash
# Se existir `personas/domain-expert.md` (sem sufixo), é BUG
if [ -f "harness/personas/domain-expert.md" ]; then
  echo "❌ Bug detectado: harness/personas/domain-expert.md (genérico) existe"
  echo "   Isso viola o invariante 12. Use apenas domain-expert-<domínio>.md"
  exit 1
fi
```

### 8. Profile `domain-expert` (genérico) NÃO criado no Hermes

```bash
# Se o tool for Hermes, profiles devem ter sufixo
if command -v hermes >/dev/null; then
  for profile in $(hermes profile list 2>/dev/null | awk '{print $1}'); do
    if [ "$profile" = "domain-expert" ]; then
      echo "❌ Profile 'domain-expert' (genérico) existe no Hermes"
      echo "   Use apenas 'domain-expert-<domínio>' (ex.: domain-expert-mandai)"
      echo "   Delete: hermes profile delete domain-expert"
      exit 1
    fi
  done
fi
```

### 9. ADR-0006 (team-manager cria branch) aplicado

```bash
if ! grep -q "ADR-0006\|team-manager.*cria.*branch" harness/AGENTS.md; then
  echo "❌ ADR-0006 não está aplicado no AGENTS.md"
  echo "   team-manager DEVE criar a branch (não o builder)"
  exit 1
fi
```

### 10. 12 invariantes do AGENTS.md §8

```bash
# Confere que o AGENTS.md tem os 15 invariantes (atualizado)
INVARIANT_COUNT=$(grep -cE "^[0-9]+\. \*\*" harness/AGENTS.md || echo 0)
if [ "$INVARIANT_COUNT" -lt 15 ]; then
  echo "⚠️ AGENTS.md tem $INVARIANT_COUNT itens numerados (esperado ≥ 15 invariantes)"
  exit 1
fi
```

### 11. Version manifest está no meta-harness

```bash
# O meta-harness-m3-code/VERSION ou manifest
if [ -f "harness/VERSION" ]; then
  VERSION=$(cat harness/VERSION)
  echo "✅ Versão do harness: $VERSION"
else
  echo "⚠️ harness/VERSION ausente (opcional, mas recomendado)"
fi
```

### 12. Bootstrap tem data

```bash
# bootstrap.md deve ter a data de hoje (evita usar versão antiga)
if ! grep -q "$(date +%Y)" harness/bootstrap.md; then
  echo "⚠️ bootstrap.md não tem o ano atual — pode ser versão antiga"
fi
```

---

## Script executável (`scripts/smoke-test.sh`)

```bash
#!/usr/bin/env bash
# Smoke test do meta-harness.
# Uso: ./harness/scripts/smoke-test.sh [REPO_OWNER/REPO]
set -e

REPO="${1:-$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || echo "")}"
FAILS=0
PASSES=0

check() {
  local name="$1"
  local cmd="$2"
  if eval "$cmd" >/dev/null 2>&1; then
    echo "  ✅ $name"
    PASSES=$((PASSES+1))
  else
    echo "  ❌ $name"
    FAILS=$((FAILS+1))
  fi
}

echo "🔎 Meta-Harness Smoke Test"
echo "Repo: $REPO"
echo

echo "1. Versão instalada (esperado: ≥ 60 arquivos)"
FILE_COUNT=$(find harness/ -type f 2>/dev/null | wc -l | tr -d ' ')
[ "$FILE_COUNT" -ge 60 ] && echo "  ✅ $FILE_COUNT arquivos" \
  || { echo "  ❌ $FILE_COUNT arquivos (esperado ≥ 60)"; FAILS=$((FAILS+1)); }
PASSES=$((PASSES+1))

echo "2. Arquivos críticos"
for f in \
  harness/bootstrap.md harness/AGENTS.md \
  harness/seed/meta-harness-seed.md \
  harness/personas/interactions.md \
  harness/personas/team-manager.md \
  harness/personas/domain-expert.template.md \
  harness/personas/solutions-architect.md \
  harness/personas/backend-engineer.md \
  harness/personas/frontend-engineer.md \
  harness/personas/quality-assurance.md \
  harness/personas/devops-engineer.md \
  harness/sensors/08-i18n-audit.md \
  harness/templates/locales.template.json \
  harness/contrib/design-decisions.md; do
  if [ -f "$f" ]; then
    echo "  ✅ $f"
  else
    echo "  ❌ $f AUSENTE"
    FAILS=$((FAILS+1))
  fi
done

echo "3. Smart routing documentado"
check "AGENTS.md tem type/technical" "grep -q 'type/technical' harness/AGENTS.md"
check "bootstrap.md tem type/infra"   "grep -q 'type/infra' harness/bootstrap.md"

echo "4. Interações matrix"
check "interactions.md presente" "[ -f harness/personas/interactions.md ]"

echo "5. Domain-experts especializados"
DOMAIN_EXPERTS=$(ls harness/personas/domain-expert-*.md 2>/dev/null | grep -v template | wc -l | tr -d ' ')
[ "$DOMAIN_EXPERTS" -ge 1 ] && echo "  ✅ $DOMAIN_EXPERTS domain-experts especializados" \
  || { echo "  ❌ Nenhum domain-expert-<domínio> (genérico proibido)"; FAILS=$((FAILS+1)); }
PASSES=$((PASSES+1))

echo "6. **CRÍTICO** — nenhum domain-expert genérico"
if [ -f "harness/personas/domain-expert.md" ]; then
  echo "  ❌ Bug: harness/personas/domain-expert.md (genérico) EXISTE"
  echo "     Renomeie para domain-expert-<domínio>.md"
  FAILS=$((FAILS+1))
else
  echo "  ✅ Sem domain-expert.md genérico"
  PASSES=$((PASSES+1))
fi

echo "7. ADR-0006 aplicado (team-manager cria branch)"
check "AGENTS.md menciona team-manager cria branch" \
  "grep -qE 'team-manager.*cria.*branch|ADR-0006' harness/AGENTS.md"

echo "8. 15 invariantes no AGENTS.md"
INV_COUNT=$(grep -cE '^[0-9]+\. \*\*' harness/AGENTS.md 2>/dev/null || echo 0)
[ "$INV_COUNT" -ge 15 ] && echo "  ✅ $INV_COUNT invariantes" \
  || { echo "  ❌ $INV_COUNT (esperado ≥ 15)"; FAILS=$((FAILS+1)); }
PASSES=$((PASSES+1))

echo "9. GitHub labels type/* (se repo configurado)"
if [ -n "$REPO" ] && command -v gh >/dev/null; then
  for t in feature technical infra bug tech-debt docs spike; do
    if gh label list --repo "$REPO" 2>/dev/null | grep -q "type/$t"; then
      echo "  ✅ type/$t"
    else
      echo "  ❌ type/$t AUSENTE no GitHub"
      FAILS=$((FAILS+1))
    fi
  done
else
  echo "  ⚠️  pulado (gh não configurado ou repo não detectado)"
fi

echo "10. Hermes profiles sem genérico (se Hermes instalado)"
if command -v hermes >/dev/null; then
  GENERIC_DOMAIN=$(hermes profile list 2>/dev/null | awk '{print $1}' | grep -x "domain-expert" || true)
  if [ -n "$GENERIC_DOMAIN" ]; then
    echo "  ❌ Profile 'domain-expert' (genérico) existe no Hermes"
    echo "     Delete: hermes profile delete domain-expert"
    FAILS=$((FAILS+1))
  else
    echo "  ✅ Sem profile domain-expert genérico"
  fi
  PASSES=$((PASSES+1))
else
  echo "  ⚠️  pulado (Hermes não instalado)"
fi

echo
echo "==================="
echo "Passes: $PASSES"
echo "Fails:  $FAILS"
echo "==================="

if [ "$FAILS" -gt 0 ]; then
  echo "❌ Smoke test FALHOU. Corrija antes de processar issues."
  exit 1
fi
echo "✅ Smoke test OK. Pode processar issues."
```

---

## Saída esperada (sucesso)

```
🔎 Meta-Harness Smoke Test
Repo: brenonaraujo/mandai-v2

1. Versão instalada (esperado: ≥ 60 arquivos)
  ✅ 62 arquivos
2. Arquivos críticos
  ✅ harness/bootstrap.md
  ✅ harness/AGENTS.md
  ...
3. Smart routing documentado
  ✅ AGENTS.md tem type/technical
  ✅ bootstrap.md tem type/infra
4. Interações matrix
  ✅ interactions.md presente
5. Domain-experts especializados
  ✅ 1 domain-experts especializados
6. **CRÍTICO** — nenhum domain-expert genérico
  ✅ Sem domain-expert.md genérico
7. ADR-0006 aplicado (team-manager cria branch)
  ✅ AGENTS.md menciona team-manager cria branch
8. 15 invariantes no AGENTS.md
  ✅ 15 invariantes
9. GitHub labels type/* (se repo configurado)
  ✅ type/feature
  ...
10. Hermes profiles sem genérico (se Hermes instalado)
  ✅ Sem profile domain-expert genérico

===================
Passes: 18
Fails:  0
===================
✅ Smoke test OK. Pode processar issues.
```

---

## Saída esperada (falha — exemplo do Mandaí v2)

```
🔎 Meta-Harness Smoke Test
Repo: brenonaraujo/mandai-v2

1. Versão instalada (esperado: ≥ 60 arquivos)
  ❌ 54 arquivos (esperado ≥ 60)            ← BUG: versão antiga
...
5. Domain-experts especializados
  ❌ Nenhum domain-expert-<domínio>          ← BUG: usou genérico
6. **CRÍTICO** — nenhum domain-expert genérico
  ❌ Bug: harness/personas/domain-expert.md (genérico) EXISTE
...
===================
Passes: 12
Fails:  3
===================
❌ Smoke test FALHOU. Corrija antes de processar issues.
```

> **Bugs do Mandaí v2 detectados pelo smoke test:**
> 1. Versão antiga (54 arquivos) — sync com `meta-harness-m3-code/`.
> 2. Nenhum `domain-expert-<x>` (só o genérico).
> 3. Smart routing não está aplicado.

---

## Quando rodar

- **Sempre** antes de processar a primeira issue de um projeto.
- **Sempre** depois de sincronizar uma nova versão do
  `meta-harness-m3-code/`.
- **Opcional:** em CI, para validar que o harness não regrediu.

---

## Quem é o dono

- **Cria e mantém:** o time de plataforma (responsável pelo
  `meta-harness-m3-code`).
- **Roda:** o `team-manager` no início de qualquer materialização.
- **Falhas param o fluxo:** se o smoke test falha, **NÃO
  processe issues** até corrigir.
