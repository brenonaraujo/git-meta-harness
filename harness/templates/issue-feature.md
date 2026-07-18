# ✨ Feature: <título curto>

## Resumo

(1-2 frases do que esta feature entrega)

## Motivação

(Por que isso importa? Qual problema resolve? Link para issue pai
se houver.)

## User story

**Como** <persona>,
**quero** <ação>,
**para que** <benefício>.

## Critérios de aceite (preenchidos pelo domain-expert)

- [ ] AC1: ...
- [ ] AC2: ...
- [ ] AC3: ...

## Casos de borda

- E1: <cenário> → <comportamento esperado>
- E2: ...

## Dependências

- Bloqueia: #<id>
- Bloqueado por: #<id>

## Esboço técnico (preenchido pelo solutions-architect)

- Componentes: backend, frontend, migrations
- Endpoints/tabelas afetados: ...
- Métricas a adicionar: ...
- Breaking change: sim/não

## Definition of Done (preenchido pelo solutions-architect)

- [ ] OpenAPI atualizado
- [ ] Migration criada
- [ ] Testes unitários (coverage ≥ 80%)
- [ ] Sensores todos verdes (lint, test, vuln, contract, 12-factor)
- [ ] Snapshot local testável (docker compose up)
- [ ] Validação do usuário

## Como testar

```bash
# (preenchido pelo builder)
docker compose -f deploy/docker-compose.yml up -d
curl http://localhost:8080/...
```

## Esforço estimado

- Tempo de implementação: X dias
- Tempo de QA: X dias
- Risco: baixo / médio / alto

---

**Labels:** `triage`, `backend`/`frontend`/`infra`
**Milestone:** vX.Y.Z
**Assignee:** (definido pelo team-manager após DoD)
