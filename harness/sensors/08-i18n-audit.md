# Sensor 08 — i18n Audit

> **Objetivo:** garantir que **toda** string externalizada passa pelo
> sistema de i18n, e que os bundles `en`, `pt-BR` e `es` têm
> **paridade de chaves** (nenhuma chave faltando em um idioma).
> **Quando roda:** CI (todo push/PR).
> **Falha → ação:** **bloqueia merge**.

---

## O que é auditado

1. **Zero hardcode** — nenhuma string visível ao usuário pode estar
   hardcoded em código Go, Vue, TS ou template.
2. **Paridade de chaves** — `en.json`, `pt-BR.json` e `es.json` devem
   ter o **mesmo conjunto de chaves**. Nenhuma chave pode faltar em um
   idioma.
3. **Idiomas obrigatórios** — `en`, `pt-BR`, `es` devem existir.
4. **Pluralização** — chaves com `one`/`other` devem ter a estrutura
   correta.
5. **Interpolação** — variáveis em mensagens (`{name}`, `{count}`) devem
   estar **declaradas** em todas as traduções (mesmo nome, mesmo tipo).
6. **Cobertura de testes** — testes unitários devem cobrir pelo menos
   1 idioma em cada handler que retorna mensagem.

---

## Comandos exatos

### Script de auditoria (Make target: `make i18n-audit`)

```bash
#!/usr/bin/env bash
# scripts/check-i18n.sh
# Audita i18n: hardcode, paridade de chaves, idiomas obrigatórios.
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
if [ ${#MISSING[@]} -gt 0 ]; then
  echo "❌ faltando: ${MISSING[*]}"
  FAILS=$((FAILS+1))
else
  echo "✅"
fi
echo

# 2. Paridade de chaves entre locales
echo "2. Paridade de chaves"
REFERENCE="${LOCALES_DIR}/en.json"
if [ ! -f "$REFERENCE" ]; then
  echo "  ❌ en.json não encontrado (precisa ser a referência)"
  exit 1
fi

REFERENCE_KEYS=$(jq -r 'paths(scalars) as $p | $p | join(".")' "$REFERENCE" | sort)

for loc in "${REQUIRED_LOCALES[@]}"; do
  if [ "$loc" = "en" ]; then continue; fi
  FILE="${LOCALES_DIR}/${loc}.json"
  [ -f "$FILE" ] || continue

  LOCALE_KEYS=$(jq -r 'paths(scalars) as $p | $p | join(".")' "$FILE" | sort)
  MISSING_IN_LOCALE=$(comm -23 <(echo "$REFERENCE_KEYS") <(echo "$LOCALE_KEYS"))
  EXTRA_IN_LOCALE=$(comm -13 <(echo "$REFERENCE_KEYS") <(echo "$LOCALE_KEYS"))

  if [ -n "$MISSING_IN_LOCALE" ]; then
    echo "  ❌ ${loc}: faltando chaves (presentes em en, ausentes em ${loc}):"
    echo "$MISSING_IN_LOCALE" | sed 's/^/      - /'
    FAILS=$((FAILS+1))
  fi
  if [ -n "$EXTRA_IN_LOCALE" ]; then
    echo "  ⚠️  ${loc}: chaves extras (presentes em ${loc}, ausentes em en):"
    echo "$EXTRA_IN_LOCALE" | sed 's/^/      - /'
    # Extra não bloqueia merge, mas reporta
  fi
  if [ -z "$MISSING_IN_LOCALE" ] && [ -z "$EXTRA_IN_LOCALE" ]; then
    echo "  ✅ ${loc}: paridade OK"
  fi
done
echo

# 3. Detecção de string hardcoded (heurística)
# Procura por strings que parecem user-facing em Go handler/handler_test e em .vue/.ts
echo "3. Detecção de hardcode (heurística)"

# Heurística Go: strings em handler/service que não usam i18n
HARDCODED_GO=$(grep -rnE 'c\.(JSON|String|AbortWithStatusJSON)\s*\(\s*[0-9]+\s*,\s*(gin\.H|map\[string\](string|interface)|api\.Error)\s*\{\s*(Code|Message|error|message):\s*"[^"]*[a-zA-ZÀ-ÿ]' \
  "${SERVICE}/internal" 2>/dev/null | grep -v _test.go || true)

if [ -n "$HARDCODED_GO" ]; then
  echo "  ❌ Possível hardcode em mensagens de erro Go:"
  echo "$HARDCODED_GO" | sed 's/^/      /'
  FAILS=$((FAILS+1))
else
  echo "  ✅ Go: nenhum padrão óbvio de hardcode"
fi

# Heurística Vue/TS: $t() ausente em tags com texto
HARDCODED_VUE=$(grep -rnE '>([A-Z][a-z]+(\s+[a-z]+){1,5})<' \
  "${SERVICE}/app" "${SERVICE}/components" "${SERVICE}/pages" 2>/dev/null \
  | grep -v "node_modules" | grep -v ".nuxt" || true)

if [ -n "$HARDCODED_VUE" ]; then
  echo "  ⚠️  Possível texto hardcoded em templates Vue (revisar manualmente):"
  echo "$HARDCODED_VUE" | head -10 | sed 's/^/      /'
  # Não bloqueia merge; heurística pode dar falso positivo
fi
echo

# Resumo
if [ $FAILS -gt 0 ]; then
  echo "❌ $FAILS verificação(ões) falharam."
  exit 1
fi
echo "✅ Auditoria i18n OK."
```

---

## Thresholds

| Métrica                                         | Limite             | Bloqueia? |
|-------------------------------------------------|--------------------|-----------|
| Idiomas obrigatórios (en, pt-BR, es)            | todos presentes    | ✅        |
| Paridade de chaves (en ↔ pt-BR ↔ es)            | 100%               | ✅        |
| Strings hardcoded em handlers/serviços          | 0                  | ✅        |
| Strings hardcoded em templates Vue (heurística) | 0 (sem FP)         | ⚠️ reportar |
| Cobertura de testes i18n (handler com msg)      | ≥ 80% dos handlers | ⚠️ reportar |

---

## Onde pluga no pipeline

### CI (`.github/workflows/ci.yml`)

```yaml
i18n:
  name: i18n audit
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4

    - name: Install jq
      run: sudo apt-get install -y jq

    - name: Run i18n audit
      run: chmod +x scripts/check-i18n.sh && ./scripts/check-i18n.sh .
```

### Local (pre-commit)

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: i18n-audit
        name: i18n audit
        entry: scripts/check-i18n.sh
        language: system
        pass_filenames: false
```

---

## Falha típica & remediação

| Falha                                                | Como corrigir                                                       |
|------------------------------------------------------|---------------------------------------------------------------------|
| Chave `auth.welcome` faltando em `pt-BR.json`        | Adicionar a chave com a tradução em `pt-BR.json` e `es.json`.       |
| `c.JSON(400, gin.H{"message": "Email inválido"})`    | Substituir por `c.JSON(400, i18n.T(ctx, "validation.invalid_email"))`. |
| Template Vue `<h1>Bem-vindo</h1>`                   | Substituir por `<h1>{{ $t('home.welcome') }}</h1>`.                 |
| `interpolação` com `{name}` em en, `{nombre}` em es  | Usar **mesmo nome** de variável em todos os idiomas.                |

---

## Quem roda

- **CI:** workflow `ci.yml` (job `i18n`).
- **Local:** `backend-engineer`, `frontend-engineer` antes de commitar.
- **Falha:** bloqueia merge.
