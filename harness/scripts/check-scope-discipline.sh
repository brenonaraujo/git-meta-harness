#!/usr/bin/env bash
# ============================================================================
# git-meta-harness — Check Scope Discipline (Sensor 11)
# ============================================================================
# Detecta vazamento de camada em outputs de domain-expert e
# solutions-architect. Emite RECOMENDAÇÃO (não bloqueia) quando
# encontra padrões técnicos proibidos.
#
# Uso:
#   ./harness/scripts/check-scope-discipline.sh domain-expert < comment.md
#   cat comment.md | ./harness/scripts/check-scope-discipline.sh solutions-architect
#   echo "$OUTPUT" | ./harness/scripts/check-scope-discipline.sh <persona>
#
# Saída:
#   - exit 0 = sem sinais (output OK)
#   - exit 0 com warnings = sinais detectados (recomendação, não bloqueia)
#   - exit 2 = erro de uso
#
# Diferente do check-parallel-builders.sh: este SENSOR é
# warning-only. Decisão de reformular é do team-manager.
# ============================================================================

set -e

PERSONA="${1:-}"
if [[ -z "$PERSONA" ]]; then
  echo "Usage: $0 <persona> < input.md" >&2
  echo "  persona: domain-expert | solutions-architect" >&2
  exit 2
fi

case "$PERSONA" in
  domain-expert|solutions-architect) ;;
  *)
    echo "❌ Unknown persona: $PERSONA (expected: domain-expert or solutions-architect)" >&2
    exit 2 ;;
esac

# Read from stdin
OUTPUT=$(cat)

# Run detection (per-persona thresholds: domain-expert is stricter)
# (thresholds are set inside the Python script via $PERSONA env var)

# Run detection
SIGNALS=$(echo "$OUTPUT" | python3 -c '
import sys, re, os
text = sys.stdin.read()
persona = os.environ.get("PERSONA", "domain-expert")

# Per-persona thresholds
thresholds = {
    "domain-expert": {
        "sql_keywords": 1, "orm_names": 1, "typeorm_nestjs": 1,
        "go_files": 1, "internal_paths": 1, "migrations": 1,
        "endpoints": 1, "func_names": 2, "prometheus": 1,
        "tokens": 1,
    },
    "solutions-architect": {
        "sql_keywords": 5, "orm_names": 3, "typeorm_nestjs": 2,
        "go_files": 5, "internal_paths": 3, "migrations": 2,
        "endpoints": 5, "func_names": 10, "prometheus": 3,
        "tokens": 1,
    },
}[persona]

checks = [
    ("sql_keywords",    r"\b(SELECT|INSERT|UPDATE|DELETE|WHERE|FROM)\b"),
    ("orm_names",       r"\b(gorm|pgx|sqlx|sqlc|gin|echo|chi|fiber|nestjs|express)\b"),
    ("typeorm_nestjs",  r"\b(TypeORM|GORM|PGx|Sqlx|Gin|Echo|NestJS)\b"),
    ("go_files",        r"\b[a-z_/]+\.go\b"),
    ("internal_paths",  r"\binternal/[a-z_/]+"),
    ("migrations",      r"\b00000[0-9]_.*\.up\.sql"),
    ("endpoints",       r"\b(GET|POST|PUT|PATCH|DELETE) /api"),
    ("func_names",      r"\b[A-Z][a-zA-Z]+\([^)]*\)\s*\{"),
    ("prometheus",      r"\b(prometheus|metrics)\.New(Counter|Histogram|Gauge)"),
    ("tokens",          r".{75000,}"),
]

results = []
for name, pattern in checks:
    count = len(re.findall(pattern, text, re.MULTILINE))
    threshold = thresholds[name]
    if count >= threshold:
        results.append(f"{name}:{count}")

print("\n".join(results) if results else "OK")
' PERSONA="$PERSONA")

# Total output size for display
CHARS=$(echo -n "$OUTPUT" | wc -c | tr -d ' ')
TOKENS_EST=$((CHARS / 3))  # rough estimate: 3 chars per token

echo "==> Scope discipline check (sensor 11, v1.11.0)"
echo "==> Persona: $PERSONA"
echo "==> Output:  $CHARS chars (~${TOKENS_EST} tokens estimated)"
echo

if [[ "$SIGNALS" == "OK" ]]; then
  echo "✅ No scope discipline issues detected."
  exit 0
fi

# Detect issues
echo "⚠️  Scope discipline issues detected (RECOMMENDATION, not blocking):"
echo "$SIGNALS" | while IFS= read -r line; do
  echo "  - $line"
done
echo
echo "Recomendação (NÃO bloqueia — encurte na próxima iteração):"
echo "  - Reformule em PILARES (o que + por quê), não BLUEPRINTS (o como)"
echo "  - Remova nomes de funções, paths, SQL, ORMs específicos"
echo "  - Aplique a skill 'solution-scoping' na próxima iteração"
echo "  - Detalhes: harness/skills/solution-scoping/SKILL.md"
echo
echo "O builder SEGUE o que está escrito. Esta recomendação é só"
echo "para a próxima iteração do refinamento."

# Always exit 0 (warning, not blocking)
exit 0
