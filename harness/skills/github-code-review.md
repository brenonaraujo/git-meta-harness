# Skill — github-code-review

> Skill para revisar PRs de forma sistemática.
> Usada por **builders** (peer review) e pelo `team-manager` (final
> review).

---

## Quando carregar

- Foi atribuído como reviewer de um PR.
- Precisa aprovar ou pedir mudanças.

---

## Checklist de revisão

### 1. Descrição do PR

- [ ] Issue referenciada (Closes #X ou Refs #X).
- [ ] "Como testar localmente" presente e funcional.
- [ ] Sensores marcados (todos verdes).

### 2. Mudança de código

- [ ] Funções ≤ 25 linhas.
- [ ] Arquivos ≤ 150 linhas (backend) / ≤ 200 (frontend).
- [ ] Sem comentários redundantes.
- [ ] Sem código duplicado (DRY).
- [ ] Nomes descritivos.
- [ ] Erros wrappados com contexto (`fmt.Errorf("...: %w", err)`).
- [ ] Logs em JSON, com `request_id` quando aplicável.
- [ ] Sem `fmt.Println` / `panic` em produção.

### 3. Testes

- [ ] Coverage ≥ 80% nos pacotes alterados.
- [ ] Casos de borda cobertos (não só happy path).
- [ ] Table-driven quando múltiplos casos.
- [ ] `require` em precondições, `assert` em validações.
- [ ] Sem testes flaky (`-shuffle=on` passou).

### 4. Backend (se Go)

- [ ] OpenAPI atualizado primeiro (spec-first).
- [ ] `internal/api/openapi.gen.go` commitado.
- [ ] Migrations em `migrations/<seq>_<nome>.{up,down}.sql`.
- [ ] Sem query N+1 não tratada.
- [ ] Transações onde necessário.
- [ ] Timeouts em chamadas externas.
- [ ] Métricas adicionadas.
- [ ] `/healthz` e `/readyz` expostos.
- [ ] Graceful shutdown (SIGTERM).

### 5. Frontend (se Nuxt/TS)

- [ ] `<script setup>` usado.
- [ ] Setup Store do Pinia (não Options).
- [ ] Composables para lógica.
- [ ] `storeToRefs` ao desestruturar store.
- [ ] Sem `any` sem justificativa.
- [ ] Sem mutação de props.
- [ ] Acessibilidade básica (labels em inputs, alt em imagens).
- [ ] i18n (se aplicável).

### 6. Infra

- [ ] Dockerfile multi-stage.
- [ ] Imagem final ≤ 25 MB (ideal).
- [ ] Non-root user.
- [ ] Healthcheck definido.
- [ ] Sem secrets hardcoded.
- [ ] `docker-compose.yml` testado localmente.
- [ ] `.env.example` atualizado.

### 7. Segurança

- [ ] Inputs validados (validator / Zod).
- [ ] Auth/authz revisado.
- [ ] Sem dados sensíveis em log.
- [ ] Sem SQL injection (usar GORM ou query parametrizada).
- [ ] Sem path traversal.
- [ ] Dependências sem CVE HIGH/CRITICAL.

### 8. Observability

- [ ] Métricas relevantes adicionadas.
- [ ] Logs estruturados com campos úteis.
- [ ] Health checks expostos.
- [ ] Tracing (se OTLP configurado): spans em fluxos críticos.

---

## Como responder

### Aprovar

```bash
gh pr review <num> --approve --body "✅ LGTM. Aprovado."
```

### Pedir mudanças

```bash
gh pr review <num> --request-changes --body "🔧 Mudanças solicitadas:

1. **handler/Login.go:42** — função com 32 linhas. Quebrar.
2. **service/auth.go:15** — comentário redundante. Remover.
3. **Sem teste** para o caso de senha inválida. Adicionar.

Bloqueando merge até corrigir."
```

### Comentar (sem bloquear)

```bash
gh pr review <num> --comment --body "💡 Sugestões (não-bloqueante):

- Considere extrair validação de email para helper
- Métrica `auth_login_duration_seconds` ficaria melhor como histogram
- Documentação de OpenAPI poderia ter exemplo de erro 401"
```

---

## Reviewer matrix

| Persona              | Revisa                                             |
|----------------------|----------------------------------------------------|
| `backend-engineer`   | outros backends, partes de Go                      |
| `frontend-engineer`  | outros frontends, partes Vue/TS                    |
| `quality-assurance`  | qualquer PR (sempre, como gate final)              |
| `solutions-architect`| PRs com mudança arquitetural (OpenAPI, schema, lib)|
| `devops-engineer`    | PRs com mudança em Dockerfile, workflow, infra     |
| `team-manager`       | qualquer PR (sempre, no fim)                       |

---

## Tom

- **Construtivo**: "considere X" em vez de "isso está errado".
- **Específico**: aponte arquivo:linha, explique o motivo, sugira a
  correção.
- **Bloqueio explícito**: "Bloqueando merge até corrigir" (com checklist).
- **Sem comentário "+1"**: o review deve ter conteúdo.

---

## Quem carrega

- `backend-engineer`, `frontend-engineer`, `quality-assurance`,
  `solutions-architect`, `devops-engineer`, `team-manager`.
