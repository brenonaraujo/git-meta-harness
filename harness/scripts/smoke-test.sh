#!/usr/bin/env bash
# Meta-Harness — Smoke Test
# Valida que o projeto tem a versão correta do meta-harness
# e que os anti-padrões conhecidos estão ausentes.
#
# Uso: ./harness/scripts/smoke-test.sh [REPO_OWNER/REPO]
# Exit code: 0 = OK, 1 = falha
#
# Ver: harness/smoke-test.md para detalhes completos.

set -e

REPO="${1:-$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || echo "")}"
FAILS=0
PASSES=0

echo "🔎 Meta-Harness Smoke Test"
echo "Repo: $REPO"
echo

# 1. Versão instalada
echo "1. Versão instalada (esperado: ≥ 60 arquivos)"
FILE_COUNT=$(find harness/ -type f 2>/dev/null | wc -l | tr -d ' ')
if [ "$FILE_COUNT" -ge 60 ]; then
  echo "  ✅ $FILE_COUNT arquivos"
  PASSES=$((PASSES+1))
else
  echo "  ❌ $FILE_COUNT arquivos (esperado ≥ 60)"
  echo "     Fix: rsync -a <meta-harness-m3-code>/harness/ ./harness/"
  FAILS=$((FAILS+1))
fi

# 2. Arquivos críticos
echo
echo "2. Arquivos críticos"
for f in \
  harness/bootstrap.md \
  harness/AGENTS.md \
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
  harness/contrib/design-decisions.md \
  harness/smoke-test.md; do
  if [ -f "$f" ]; then
    echo "  ✅ $f"
    PASSES=$((PASSES+1))
  else
    echo "  ❌ $f AUSENTE"
    FAILS=$((FAILS+1))
  fi
done

# 3. Smart routing documentado
echo
echo "3. Smart routing documentado"
if grep -q 'type/technical' harness/AGENTS.md 2>/dev/null; then
  echo "  ✅ AGENTS.md tem type/technical"
  PASSES=$((PASSES+1))
else
  echo "  ❌ AGENTS.md NÃO tem type/technical (smart routing ausente)"
  FAILS=$((FAILS+1))
fi
if grep -q 'type/infra' harness/bootstrap.md 2>/dev/null; then
  echo "  ✅ bootstrap.md tem type/infra"
  PASSES=$((PASSES+1))
else
  echo "  ❌ bootstrap.md NÃO tem type/infra"
  FAILS=$((FAILS+1))
fi

# 4. Interações matrix
echo
echo "4. Interações matrix"
if [ -f "harness/personas/interactions.md" ]; then
  echo "  ✅ interactions.md presente"
  PASSES=$((PASSES+1))
else
  echo "  ❌ interactions.md AUSENTE (matriz de papéis)"
  FAILS=$((FAILS+1))
fi

# 5. Domain-experts especializados
echo
echo "5. Domain-experts especializados (esperado: ≥ 1)"
DOMAIN_EXPERTS=$(ls harness/personas/domain-expert-*.md harness/personas/examples/domain-expert-*.md 2>/dev/null | grep -v template | wc -l | tr -d ' ')
if [ "$DOMAIN_EXPERTS" -ge 1 ]; then
  echo "  ✅ $DOMAIN_EXPERTS domain-experts especializados"
  PASSES=$((PASSES+1))
else
  echo "  ❌ Nenhum domain-expert-<domínio> (genérico proibido)"
  echo "     Fix: cp harness/personas/domain-expert.template.md harness/personas/domain-expert-<seu-dominio>.md"
  FAILS=$((FAILS+1))
fi

# 6. CRÍTICO — nenhum domain-expert genérico
echo
echo "6. **CRÍTICO** — nenhum domain-expert genérico"
if [ -f "harness/personas/domain-expert.md" ]; then
  echo "  ❌ Bug: harness/personas/domain-expert.md (genérico) EXISTE"
  echo "     Isso viola o invariante 12. Renomeie para domain-expert-<domínio>.md"
  FAILS=$((FAILS+1))
else
  echo "  ✅ Sem domain-expert.md genérico"
  PASSES=$((PASSES+1))
fi

# 7. ADR-0006 aplicado (team-manager cria branch)
echo
echo "7. ADR-0006 aplicado (team-manager cria branch)"
if grep -qE 'team-manager.*cria.*branch|ADR-0006' harness/AGENTS.md 2>/dev/null; then
  echo "  ✅ AGENTS.md menciona team-manager cria branch"
  PASSES=$((PASSES+1))
else
  echo "  ❌ AGENTS.md NÃO menciona team-manager cria branch (ADR-0006 ausente)"
  FAILS=$((FAILS+1))
fi

# 8. 15 invariantes no AGENTS.md
echo
echo "8. Invariantes no AGENTS.md (esperado: ≥ 15)"
INV_COUNT=$(grep -cE '^[0-9]+\. \*\*' harness/AGENTS.md 2>/dev/null || echo 0)
if [ "$INV_COUNT" -ge 15 ]; then
  echo "  ✅ $INV_COUNT invariantes"
  PASSES=$((PASSES+1))
else
  echo "  ❌ $INV_COUNT (esperado ≥ 15)"
  FAILS=$((FAILS+1))
fi

# 9. GitHub labels type/* (se repo configurado)
echo
echo "9. GitHub labels type/* (se repo configurado)"
if [ -n "$REPO" ] && command -v gh >/dev/null 2>&1; then
  for t in feature technical infra bug tech-debt docs spike; do
    if gh label list --repo "$REPO" 2>/dev/null | grep -q "type/$t"; then
      echo "  ✅ type/$t"
      PASSES=$((PASSES+1))
    else
      echo "  ⚠️  type/$t AUSENTE no GitHub (criável: gh label create type/$t --color XXX)"
    fi
  done
else
  echo "  ⚠️  pulado (gh não configurado ou repo não detectado)"
fi

# 10. Hermes profiles sem genérico (se Hermes instalado)
echo
echo "10. Hermes profiles sem genérico (se Hermes instalado)"
if command -v hermes >/dev/null 2>&1; then
  if hermes profile list 2>/dev/null | awk '{print $1}' | grep -qx "domain-expert"; then
    echo "  ❌ Profile 'domain-expert' (genérico) existe no Hermes"
    echo "     Fix: hermes profile delete domain-expert"
    FAILS=$((FAILS+1))
  else
    echo "  ✅ Sem profile domain-expert genérico"
    PASSES=$((PASSES+1))
  fi
else
  echo "  ⚠️  pulado (Hermes não instalado)"
fi

echo
echo "==================="
echo "Passes: $PASSES"
echo "Fails:  $FAILS"
echo "==================="

if [ "$FAILS" -gt 0 ]; then
  echo
  echo "❌ Smoke test FALHOU. Corrija antes de processar issues."
  echo "   Ver harness/smoke-test.md para diagnóstico detalhado."
  exit 1
fi

echo
echo "✅ Smoke test OK. Pode processar issues."
