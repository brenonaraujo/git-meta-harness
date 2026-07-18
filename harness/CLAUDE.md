# CLAUDE.md — atalho para Claude Code

> Este é o **atalho** que o `team-manager` cria em projetos que usam
> Claude Code. É apenas um symlink/conteúdo equivalente ao
> `AGENTS.md`, formatado para ser carregado automaticamente.
>
> **Spec canônica:** [`AGENTS.md`](./AGENTS.md) · **Spec completa:**
> [`bootstrap.md`](./bootstrap.md) · **Versão do meta-harness:** 1.5.0

---

## TL;DR para Claude Code

1. **Source of truth:** GitHub Issues + `harness/*`.
2. **Personas (7):** team-manager, domain-expert-`<domínio>` (sempre
   especializado), solutions-architect, backend-engineer, frontend-engineer,
   quality-assurance, devops-engineer.
3. **Stack pinada** (ver `harness/stack/versions.md`): Go 1.26.5 +
   Gin + GORM + PostgreSQL 18.4 + OpenAPI (backend); Nuxt 4.5 + Nuxt UI +
   Pinia + Node 24 LTS (frontend).
4. **Limites:** ≤ 25 linhas/função, ≤ 150 linhas/arquivo, coverage
   ≥ 80%, 12-factor, KISS/DRY.
5. **Workflow:** issue → triage → refined → ready → in-progress →
   in-review → qa → validação → done.
6. **PR:** 1 issue = 1 PR; "Como testar localmente" obrigatório.
7. **i18n:** en, pt-BR, es — paridade obrigatória.
8. **Sensores (10):** 00 lint · 01 vuln · 02 unit · 03 contract ·
   04 image · 05 smoke · 06 load · 07 12-factor · 08 i18n ·
   09 verify-after-build (team-manager re-checa antes de QA).

---

## Comandos canônicos

```bash
# Backend (Go)
make tidy build test lint vuln oas migrate-up run docker compose-up

# Frontend (Nuxt)
pnpm dev build lint typecheck test test:run audit docker:build

# Pré-flight (gate antes de PR)
./harness/scripts/smoke-test.sh .
./harness/scripts/check-stack-versions.sh --check-latest
```

---

## Invariantes (não-violáveis, 19)

1. Toda issue tem commits referenciando o número.
2. Todo PR cita a issue e tem "Como testar localmente".
3. Todo microsserviço expõe `/healthz`, `/readyz`, `/metrics`.
4. Todo microsserviço loga em JSON via `slog`.
5. Nenhum microsserviço lê config de arquivo (só env).
6. Nenhum microsserviço roda como root no container.
7. Nenhum microsserviço entra em produção sem `govulncheck` verde.
8. Nenhum PR é mergeado sem coverage ≥ 80% nos pacotes alterados.
9. Nenhuma issue é fechada sem validação do usuário.
10. Nenhuma string de usuário é hardcoded (i18n obrigatório).
11. Toda issue é roteada ao `domain-expert-<domínio>` (nunca genérico).
12. Toda issue tem label `type/*` na triagem.
13. Issue-mãe só fecha quando todas as sub-issues estão `done`.
14. Branches de feature são criadas pelo team-manager; builders clonam.
15. Nenhum PR é aberto com CI local vermelho.
16. **1 Dockerfile por service em path canônico** (sem `Dockerfile` na
    raiz; sem 2+ Dockerfiles pro mesmo service).
17. **CI modular com path filters** (dorny/paths-filter + concurrency
    + scope cache + Trivy SHA-pinado + GOTOOLCHAIN=local).
18. Toda decisão de versão passa por `check-stack-versions.sh --check-latest`.
19. **Team-manager verifica, não confia.** Re-executa checks críticos
    após o builder reportar verde, antes de mover para `in-review` ou
    pedir validação humana. Auto-relato de subagente é evidência fraca.

---

## Smart routing (team-manager)

| Label da issue    | Personas no fluxo                                           |
|-------------------|-------------------------------------------------------------|
| `type/feature`    | domain-expert → solutions-architect → builder → QA        |
| `type/technical`  | (sem domain-expert) solutions-architect → builder → QA    |
| `type/infra`      | (sem domain-expert, sem builder) solutions-architect → devops |
| `type/bug`        | (sem domain-expert) solutions-architect → builder → QA    |
| `type/tech-debt`  | (sem domain-expert) builder → QA                            |
| `type/docs`       | team-manager apenas (revisão editorial)                     |
| `type/spike`      | solutions-architect apenas (saída = ADR)                    |

---

Para detalhes, leia **integralmente** [`AGENTS.md`](./AGENTS.md) e
[`bootstrap.md`](./bootstrap.md) antes de agir.
