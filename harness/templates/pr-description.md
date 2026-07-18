# (<issue-id>) <título da issue>

## Summary

(1 parágrafo do que foi feito e por quê)

## Issue

Closes #<id>

Refs #<id> (se for sub-issue ou dependência)

## Changes

- [ ] Mudança 1
- [ ] Mudança 2
- [ ] Mudança 3

## Sensores (todos verdes)

- [ ] `make lint` (golangci-lint v2) — OK
- [ ] `make test` (go test -race -coverprofile) — coverage ≥ 80% nos pacotes alterados
- [ ] `make vuln` (govulncheck) — sem HIGH/CRITICAL
- [ ] `trivy image` — sem CRITICAL (waivers #X se aplicável)
- [ ] `openapi-diff` — sem breaking changes
- [ ] `12-factor audit` — F1..F12 ✅
- [ ] `pnpm lint` (se frontend) — OK
- [ ] `pnpm test` (se frontend) — coverage ≥ 80%
- [ ] `pnpm typecheck` (se frontend) — OK

## Como testar localmente

```bash
# 1. Subir ambiente
docker compose -f deploy/docker-compose.yml up -d --build

# 2. Esperar ficar pronto
docker compose -f deploy/docker-compose.yml exec backend \
  wget -q -O- http://localhost:8080/healthz

# 3. Testar fluxos críticos
# Backend:
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}'
# Frontend:
# Abrir http://localhost:3000 e...

# 4. (Opcional) Inspecionar métricas
curl -s http://localhost:8080/metrics | grep -E "^(http_requests_total|app_info)"
```

## Screenshots / exemplos de resposta

(anexar ou colar)

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2026-07-19T12:00:00Z"
}
```

## Riscos & rollback

- **Risco 1:** <descrição>
  - **Mitigação:** ...
- **Rollback:** reverter merge (`git revert <sha>` e tag `vX.Y.Z-1`),
  ou desabilitar feature flag.

## Checklist de revisão

- [ ] Funções ≤ 25 linhas
- [ ] Arquivos ≤ 150 linhas
- [ ] Sem comentários redundantes
- [ ] Testes cobrem bordas (não só happy path)
- [ ] Métricas Prometheus adicionadas/atualizadas
- [ ] Logs slog JSON com campos relevantes
- [ ] Health/readiness endpoints OK
- [ ] OpenAPI atualizado (se mudou contrato)
- [ ] Migration criada (se mudou schema)
- [ ] Dockerfile atualizado (se mudou deps)
- [ ] docker-compose up -d funciona do zero

## Referências

- Issue: #<id>
- ADR (se houver): docs/adr/XXXX-<titulo>.md
- Docs relacionadas: ...
