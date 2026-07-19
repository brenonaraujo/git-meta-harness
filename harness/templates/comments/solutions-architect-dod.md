<!-- solutions-architect-dod.md (v1.13.0)
     Comentário canônico de DoD do solutions-architect.
     SEMPRE use este template (copie e preencha) — sensor 13
     `feature-flow` valida que o comentário tem pelo menos
     1 pilar e 1 item DoD (BLOQUEANTE).
     Ver skill `solution-scoping` + invariante 24. -->

## 🏛️ Definition of Done — `solutions-architect`

> **Issue:** #<id>
> **Refinamento de:** @<domain-expert> em #<comment-id>
> **Domínio:** `domain/<nome>`

### Pilares (alto nível, 3-5)

> Princípio **PILARES, não BLUEPRINTS** (skill `solution-scoping`,
> invariante 22). Cada pilar é o que + por quê, **NUNCA**
> nomes de funções, paths, SQL, ORM.

1. **Pilar 1**: ...
2. **Pilar 2**: ...
3. **Pilar 3**: ...
(recomendado 3-5; máximo 5)

### Definition of Done (DoD)

- [ ] Backend: <macro do que o backend precisa entregar>
- [ ] Frontend: <macro do que o frontend precisa entregar>
- [ ] Tests: cobertura ≥ 80% no path-scope; e2e do fluxo crítico
- [ ] i18n: paridade en/pt-BR/es (sensor 08)
- [ ] 12-factor: F1..F12 OK
- [ ] CI: lint + test + audit + openapi-diff verdes
- [ ] Local pre-flight: `make lint && make test && make vuln` (ou equivalente)
- [ ] Path-scope declarado: <glob do(s) arquivo(s) alterado(s)>
- [ ] Dependências: `depends-on: #X` se aplicável
- [ ] Documentação: SPEC.md / ADR atualizado se mudou contrato

### Decisões arquiteturais (ADR-lite inline)

**Decisão 1**: <título>
- **Contexto**: ...
- **Opções consideradas**: A (pro), B (contra), C (escolhida)
- **Decisão**: ...
- **Consequências**: ...

(mínimo 1 se a issue tem impacto arquitetural; pode ser
"sem decisões arquiteturais novas" para issues simples)

### Riscos & rollback

- **Risco 1**: ... → mitigação: ...
- **Rollback**: como reverter se der errado em produção

### 12-factor audit (resumo)

| Fator | OK? | Notas |
|---|---|---|
| F1 Codebase | ✅ | 1 repo, 1 deploy |
| F2 Dependencies | ✅ | declared in go.mod / package.json |
| F3 Config | ✅ | env vars only (verificável em CI) |
| F4 Backing services | ✅ | treat as attached resources |
| F5 Build/release/run | ✅ | separate stages, CI modular |
| F6 Processes | ✅ | stateless, shared-nothing |
| F7 Port binding | ✅ | export via HTTP_PORT |
| F8 Concurrency | ✅ | scale out via process model |
| F9 Disposability | ✅ | fast startup, graceful shutdown |
| F10 Dev/prod parity | ✅ | same backing services, same code |
| F11 Logs | ✅ | JSON to stdout (slog) |
| F12 Admin processes | ✅ | run as one-off in prod-like env |

---

> **Próximo passo:** quando terminar, adicionar a label
> `ready` à issue. O `team-manager` então aciona o builder
> (backend ou frontend, conforme o DoD).
