#!/usr/bin/env bash
# check-i18n.sh — Audita i18n: hardcode, paridade de chaves, idiomas obrigatórios.
# Extraído de harness/sensors/08-i18n-audit.md (v1.6.7).
# Uso: ./scripts/check-i18n.sh [service-dir]
set -e

SERVICE="${1:-.}"
LOCALES_DIR="${SERVICE}/internal/i18n/locales"
[ -d "$LOCALES_DIR" ] || LOCALES_DIR="${SERVICE}/i18n/locales"
[ -d "$LOCALES_DIR" ] || LOCALES_DIR="${SERVICE}/locales"

REQUIRED_LOCALES=("en" "pt-BR" "es")
FAILS=0

echo "🔎 i18n audit: ${SERVICE}"
echo "  Locales dir: ${LOCALES_DIR}"
echo

# 1. Idiomas obrigatórios existem
echo -n "1. Idiomas obrigatórios (en, pt-BR, es) ... "
MISSING=()
for loc in "${REQUIRED_LOCALES[@]}"; do
  if [ ! -f "${LOCALES_DIR}/${loc}.json" ]; then
    MISSING+=("$loc")
  fi
done
if [ ${#MISSING[@]} -eq 0 ]; then
  echo "✅"
else
  echo "❌ faltando: ${MISSING[*]}"
  FAILS=$((FAILS+1))
fi

# 2. Paridade de chaves entre locales
echo "2. Paridade de chaves"
REFERENCE=""
for loc in "${REQUIRED_LOCALES[@]}"; do
  if [ -f "${LOCALES_DIR}/${loc}.json" ]; then
    REFERENCE="$loc"
    break
  fi
done

if [ -z "$REFERENCE" ]; then
  echo "  ❌ Nenhum locale de referência encontrado"
  FAILS=$((FAILS+1))
else
  echo "  📌 Referência: ${REFERENCE}.json"
  REF_KEYS=$(jq -r 'keys[]' "${LOCALES_DIR}/${REFERENCE}.json" 2>/dev/null | sort)
  if [ -z "$REF_KEYS" ]; then
    echo "  ❌ ${REFERENCE}.json vazio ou inválido"
    FAILS=$((FAILS+1))
  else
    for loc in "${REQUIRED_LOCALES[@]}"; do
      if [ "$loc" = "$REFERENCE" ]; then continue; fi
      if [ ! -f "${LOCALES_DIR}/${loc}.json" ]; then continue; fi
      LOC_KEYS=$(jq -r 'keys[]' "${LOCALES_DIR}/${loc}.json" 2>/dev/null | sort)
      DIFF=$(diff <(echo "$REF_KEYS") <(echo "$LOC_KEYS") || true)
      if [ -n "$DIFF" ]; then
        echo "  ❌ ${loc}.json divergente de ${REFERENCE}.json:"
        echo "$DIFF" | head -10 | sed 's/^/    /'
        FAILS=$((FAILS+1))
      else
        echo "  ✅ ${loc}.json em paridade"
      fi
    done
  fi
fi

# 3. Hardcode de strings em PT/EN/ES no código (heurística simples)
echo
echo "3. Hardcode (heurística: strings em código sem i18n)"
if [ -d "${SERVICE}/internal" ]; then
  HARDCODES=$(grep -rE '"[A-Z][a-z]+( [A-Z][a-z]+)+"' "${SERVICE}/internal" 2>/dev/null \
    | grep -v "_test.go\|//\|nolint" | head -5 || true)
  if [ -n "$HARDCODES" ]; then
    echo "  ⚠️  Possíveis strings hardcoded (revisar manualmente):"
    echo "$HARDCODES" | sed 's/^/    /'
  else
    echo "  ✅ Nenhum hardcode óbvio"
  fi
fi

echo
if [ $FAILS -gt 0 ]; then
  echo "❌ $FAILS verificação(ões) falharam."
  exit 1
fi
echo "✅ Auditoria i18n OK."
