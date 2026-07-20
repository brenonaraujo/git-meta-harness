# Skill: spec-decomposition

> **Quando usar:** ao receber uma spec funcional (markdown,
> 1-50 páginas) e precisar quebrá-la em épicos + sub-issues
> estruturadas. Usado por `gmh new --spec` (v1.14.0+, ADR-0028)
> e por `team-manager` ao receber uma spec grande de Brenon.
>
> **Output típico:** TODO list com 5-15 épicos, 30-100
> sub-issues, cada uma com ACs + edge cases + SpecRef
> (link pra seção da spec).

---

## 1. Quando esta skill se aplica

- Spec de 1-50 páginas (markdown ou texto).
- Múltiplas áreas funcionais (auth, workspace, products,
  orders, payments, etc).
- Time de 2-10 engineers.
- Stack: Go + Nuxt (default), ou TS/Node, ou Python, etc.

**NÃO se aplica** se:
- Spec é < 1 página (1-3 issues, decomposição manual).
- Spec é puramente técnica (use `solution-scoping` skill).
- Spec é de bug/tech-debt (1 issue, sem decomposição).

---

## 2. Heurística de decomposição

### 2.1 Épicos

**1 épico = 1 capítulo/área da spec.** Critérios:

- **Coerência funcional**: todas sub-issues do épico
  pertencem ao mesmo bounded context (DDD).
- **Testável independentemente**: cada épico pode ser
  entregue e validado sem depender dos outros (em ordem).
- **Tamanho适中**: 5-15 sub-issues por épico. Mais que 15 =
  decompor mais. Menos que 3 = é sub-issue, não épico.

**Nomes**: `F1: Auth`, `F2: Workspaces + Products`, `F3: Orders + Pix`.
- Prefixo `F<n>` sequencial.
- Título descritivo (1-5 palavras).
- Sem nomes de UI (botões, páginas) nem de tecnologia
  (PostgreSQL, Nuxt).

### 2.2 Sub-issues

**1 sub-issue = 1 entregável testável.** Critérios:

- **1 user story / 1 endpoint / 1 model**: cada sub-issue
  implementa 1 coisa que pode ser testada end-to-end.
- **PR-able**: 1 sub-issue = 1 PR (ou menos).
- **ACs explícitos**: 2-5 critérios de aceite derivados
  da spec (ver §3).
- **Edge cases**: 1-3 casos de borda (ver §4).

**Nomes**: `F1.2: switch-role`, `F2.1: workspace-creation`.
- Prefixo `F<n>.<m>` (épico.sub).
- kebab-case.
- Descreve o que FAZ, não como.

### 2.3 Labels

Toda issue recebe:
- `type/feature` (sempre, para type/feature issues).
- `domain/<domínio>` (ex.: `domain/marketplace`).
- `priority/p0` (crítico) | `p1` (alta) | `p2` (média) | `p3` (baixa).
- `size/S` (≤ 1 dia) | `M` (1-3 dias) | `L` (3-7 dias) | `XL` (> 7 dias).
- `area/auth` | `area/payments` | `area/<x>` (bounded context).

---

## 3. ACs (Acceptance Criteria) — derivação

**Regra**: cada AC é uma frase que começa com verbo (will, must,
should) e descreve comportamento observável.

**Heurística**:
1. Leia a seção da spec.
2. Procure bullet points (`-` ou `*`).
3. Cada bullet vira 1 AC (reformule com verbo + objeto).
4. Procure "if X then Y" → 1 AC pra X (caminho feliz) + 1 AC
   pra Y (caminho alternativo).
5. Se seção não tem bullets, escreva 2-3 ACs derivados.

**Exemplo** (spec: "Multi-tenant auth with role switching"):

```markdown
- [ ] AC1: User can log in with Brazilian mobile number
       (11 digits) and receive SMS OTP within 30 seconds.
- [ ] AC2: User can switch role (Morador ↔ Líder ↔ Fornecedor
       ↔ Admin) via bottom nav, and JWT reflects current role
       within 1 second.
- [ ] AC3: Switching role does not require re-login; current
       session is preserved.
```

**Mínimo**: 1 AC por sub-issue. **Recomendado**: 2-5. **Máximo
prático**: 8 (acima disso, decompor).

---

## 4. Edge cases — extração

**Regra**: cada edge case é uma frase que descreve um cenário
de borda, com ou sem tratamento esperado.

**Heurística**:
1. Procure "if X", "when X", "exception", "nota:", "warning",
   "cuidado", "edge case".
2. Procure valores limite: empty, null, expired, missing,
   invalid, zero, max, min.
3. Procure race conditions: concurrent, async, parallel.
4. Procure i18n / timezone / locale: BR = UTC-3, multi-tenant.
5. Se nada óbvio, escreva 1-2 edge cases derivados do domínio.

**Exemplo** (auth):

```markdown
- [ ] Edge 1: User with expired OTP (5min) gets 401 with
       "OTP_EXPIRED" error and can request a new one.
- [ ] Edge 2: User with 3 failed login attempts in 10min
       gets 429 (rate limited) for 1 hour.
- [ ] Edge 3: User with mobile number not registered gets
       404 with "USER_NOT_FOUND" (no info leak).
```

**Mínimo**: 1 edge case por sub-issue. **Recomendado**: 2-3.

---

## 5. SpecRef — rastreabilidade

**Regra**: cada sub-issue body inclui link pra seção da spec
de origem. Permite rastreabilidade 100%.

**Formato**:
```markdown
**Spec ref:** `spec.md#<anchor>`

Onde `<anchor>` é o slugify do título da seção:
- "Multi-tenant auth" → `multi-tenant-auth`
- "F1.2: Switch Role" → `f12-switch-role`
```

**Por quê**:
- Reviewer pode validar ACs contra spec original.
- Drift de spec é detectável (qual issue ficou órfã?).
- Onboarding de novo engineer: "comece pela spec, depois
  olhe as issues linkadas".

---

## 6. Exemplo completo (v1.14.0)

**Spec input** (5 seções):

```markdown
# Mandaí v2

## F1: Auth + Switch Role
Login com CPF/CNPJ + OTP. Switch role.

### F1.1: Mobile Auth
Login com número BR (11 dígitos) + OTP.
### F1.2: Switch Role
Bottom nav para alternar entre 4 roles.

## F2: Workspaces
Multi-tenant por Fornecedor.
```

**TODO output** (harness/TODO.md, machine-readable via
harness/TODO.json):

```markdown
# TODO — Mandaí v2

> Domain: marketplace.

## F1: Auth + Switch Role (p0)
**Spec ref:** `spec.md#f1-auth-switch-role`
**Labels:** `type/feature`, `domain/marketplace`, `priority/p0`

**ACs:**
- User can log in with Brazilian mobile number (11 digits)
- User receives SMS OTP within 30 seconds
- ...

**Sub-issues:**
- **F1.1** — Mobile Auth (CPF/CNPJ)
- **F1.2** — Switch Role

## F2: Workspaces (p0)
**Spec ref:** `spec.md#f2-workspaces`
**Labels:** `type/feature`, `domain/marketplace`, `priority/p0`

**Sub-issues:**
- **F2.1** — Workspace Creation
```

---

## 7. Validação (antes de commitar TODO)

Rode antes de commitar:

1. **Coverage check**: `harness/SPEC-COVERAGE.md` deve ter
   100% de cobertura (cada seção da spec → ≥1 épico).
2. **Size check**: nenhum sub-issue > 7 dias de esforço
   (decompor se for).
3. **Priority check**: épicos críticos (p0) são os primeiros
   1-3, não os últimos.
4. **Label check**: cada issue tem `type/feature` +
   `domain/<x>` + `priority/<x>`.
5. **SpecRef check**: cada issue tem `Spec ref:` no body.

**Quality gate** (sensor 14 futuro, v1.15.0+):

```bash
./harness/scripts/check-spec-decomposition.sh
```

(v1.14.0: manual; v1.15.0+: automated sensor 14).

---

## 8. Quem usa

- `team-manager` (principal): ao receber spec grande de
  Brenon, quebra em TODO.
- `domain-expert-adopter` (v1.14.0+): em `gmh adopt` /
  `gmh new --spec`.
- LLM agents (Claude, Hermes) com `gmh new --spec` CLI.

---

## 9. Quem mantém

- Brenon + time de plataforma.
- Atualizada quando ADRs 0027, 0028, 0029 evoluem.
