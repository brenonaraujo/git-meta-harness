---
name: pre-implementation-design
version: 1.0.0
type: design-thinking
applies-to: backend-engineer, frontend-engineer
---

# Pre-Implementation Design — Think Before You Code

Skill for the **builder** personas (`backend-engineer`,
`frontend-engineer`). Applies when implementing any function
that **might exceed 25 lines** (or any non-trivial logic).

> **Lição do Mandaí v2 (jul/2026, Épico #12):** o builder backend
> reportou `1 issue — OnboardRole tem 34 linhas (limite 25)` e
> teve que "quebrar em função auxiliar" — extração mecânica
> que adicionou **glue code** sem ganho de clareza. Resultado:
> 2 funções coesas viraram 4 funções coesas + 1 helper que
> só delega. Total: mais código, mais arquivos, mais imports
> — sem ganho de leitura.
>
> O problema NÃO era a função de 34 linhas. O problema era
> que o builder não pensou em **abstração** antes de
> implementar — só tentou encaixar no limite de 25 linhas
> via extração mecânica.

**v1.10.0 (ADR-0020)** aumenta o limite duro de 25 → 35
linhas e adiciona esta skill para forçar o builder a
**pensar antes de codar**, eliminando o anti-pattern de
"split artificial só pra caber em N linhas".

---

## 🚦 Regra de tráfego (quando aplicar)

| Situação | Aplicar skill? |
|---|---|
| Função tem ≤ 25 linhas e é coesa | ❌ Não. Implementar direto. |
| Função tem ≤ 25 linhas mas é "glue" (chamadas sequenciais sem lógica) | ⚠️ Sim. Provavelmente há abstração melhor. |
| Função tem 26-35 linhas | ✅ **Sim, obrigatório.** Justifique ou refatore. |
| Função tem 36+ linhas | 🛑 **Bloqueado pelo `funlen`**. Refatore primeiro. |
| Função com múltiplas responsabilidades (validação + lógica + persistência + audit) | ✅ Sim, mesmo se ≤ 25 linhas. Decomponha por responsabilidade. |
| Função com `for` aninhado, múltiplos `if err != nil` em sequência | ✅ Sim. Extraia pipeline. |

---

## 📋 Protocolo "3 Decomposições" (antes de codar)

Para qualquer função não-trivial (≥ 26 linhas ou múltiplas
responsabilidades), **antes de escrever código**, siga:

### Passo 1 — Listar 2-3 decomposições possíveis

Pense em **pelo menos 2 alternativas** de como quebrar (ou
não quebrar) a função:

```markdown
## Decomposição de `OnboardRole(user, role, tenantID)`

### Opção A — Função única (32 linhas)
- Coesa: pipeline de 1 transação
  (validate → create user → assign role → audit)
- Glue mínimo
- Pró: legível de cima a baixo
- Contra: 1 lugar pra testar, harder to mock intermediate

### Opção B — 4 helpers (1 main + 3 helpers, ~12 linhas cada)
- `validateOnboarding()` / `createUserWithRole()` / `auditOnboarding()`
- Pró: cada função testável isoladamente
- Contra: 4 funções pra navegar, glue explícito

### Opção C — 2 helpers (1 main + 2 helpers, ~16 linhas cada)
- `setupUserAndRole()` (valida + cria) / `commitOnboarding()` (persiste + audit)
- Pró: ainda coeso, testa 2 fluxos principais
- Contra: helpers ainda longos (~16 linhas cada)
```

### Passo 2 — Justificar a escolha

```markdown
### Escolha: Opção A (32 linhas)

**Por quê**: a função é uma **transação atômica**. Os 3
helpers de B fragmentariam a leitura sem ganho real
(qualquer teste de integração cobre os 3 juntos). A
opção C não ganha muito sobre A.

**Quando reverteria**: se a parte de `audit` virar
compliance de outro time (ex.: LGPD, BACEN), aí faz
sentido extrair `auditOnboarding()` para um módulo
separado com ownership claro.
```

### Passo 3 — Documentar no commit message ou no PR body

```markdown
refactor: extract OnboardRole into 32-line atomic function (ADR-0020)

Decomposition considered:
- A: 32-line atomic function (chosen — coesa, transação atômica)
- B: 4 helpers (rejected — glue sem ganho)
- C: 2 helpers (rejected — não ganha muito sobre A)

Limit relaxed from 25 → 35 (v1.10.0) — this is intentional.
Skill: pre-implementation-design
```

### Passo 4 — Validar com sensor

```bash
make lint  # funlen check
# Se passou (≤ 35): ✅
# Se passou > 35: refatore
```

---

## 🧠 Heurísticas para decidir entre "uma função" vs "decompor"

### ✅ Prefira UMA função quando:

1. **Pipeline linear curto** (validate → action → persist)
   onde cada etapa é 1-2 linhas e o todo é legível
   de cima a baixo.
2. **Transação atômica** (operações que devem ser
   revertidas juntas — usuário, role, audit, etc).
3. **Coesão narrativa** (a função conta uma história
   que faz sentido ler linearmente).
4. **Mocks seriam mais complexos** que o código original
   (ex.: se você precisa mockar 3 interfaces pra testar
   1 helper, o helper não está ajudando).

### ❌ Decomponha em helpers quando:

1. **Múltiplas responsabilidades distintas** (validação
   vs. lógica vs. persistência vs. audit).
2. **Lógica reutilizada** em 2+ lugares (helper vale
   a pena).
3. **Teste isolado** faz sentido (ex.: testar
   `validateOnboarding()` sem precisar de DB).
4. **Complexidade ciclomática** > 15 (early return
   aninhado, múltiplos `if err != nil`, switch grande).
5. **Você precisa de mocks diferentes** para cada parte
   (sinal de que cada parte tem边界 próprios).

---

## 🚫 Anti-patterns que a skill ELIMINA

### Anti-pattern 1: "Split for compliance"

```go
// ❌ Antes (limite rígido 25)
func OnboardRole(...) error {
    return doOnboardRole1(...)  // só delega
}

func doOnboardRole1(...) error {
    if err := validate(...); err != nil { return err }  // 1 linha
    if err := create(...); err != nil { return err }     // 1 linha
    if err := assignRole(...); err != nil { return err } // 1 linha
    return audit(...)
}
// 4 funções + glue. Mais código, menos clareza.
```

```go
// ✅ Depois (limite soft 35, skill aplicada)
func OnboardRole(...) error {
    if err := validate(...); err != nil { return err }
    if err := create(...); err != nil { return err }
    if err := assignRole(...); err != nil { return err }
    return audit(...)
}
// 1 função coesa, 32 linhas, transação atômica.
```

### Anti-pattern 2: "Helper vazio"

```go
// ❌ Helper que só repassa args
func (s *Service) getUserByID(ctx context.Context, id string) (*User, error) {
    return s.repo.GetByID(ctx, id)  // 1 linha
}
```

Helper que delega sem lógica própria é overhead sem valor.
Use a chamada direta.

### Anti-pattern 3: "Mega-função sem abstração"

```go
// ❌ 60 linhas, 4 responsabilidades
func ProcessCheckout(...) error {
    // 15 linhas: validate cart
    // 15 linhas: process payment
    // 15 linhas: update inventory
    // 15 linhas: send confirmation
}
```

Aqui a skill **força decompor** (não cabe em 35). O
objetivo é fazer o builder perceber que há 4
responsabilidades distintas.

---

## 📏 Limites atualizados (v1.10.0)

| Faixa | Status | Ação |
|---|---|---|
| 0-25 linhas | ✅ Ideal | Manter |
| 26-35 linhas | ⚠️ Aceitável | **Aplicar skill: justificar ou refatorar** |
| 36+ linhas | ❌ Erro | `funlen` falha. Refatorar. |

**Recomendação**: 25 (mantido). **Limite duro**: 35 (subiu de 25).

---

## 🔗 Quem aplica

- **`backend-engineer`** (Go) — sempre que implementar
  função que pode passar de 25 linhas. Documentar via
  comentário `// Skill: pre-implementation-design` no
  topo da função OU no commit message.
- **`frontend-engineer`** (Vue/TypeScript) — sempre que
  implementar composable, helper, ou componente com
  lógica não-trivial.

## 🔗 Quem valida

- **`team-manager`** (sensor 09 verify-after-build):
  re-executa `make lint` e verifica que funções > 25
  têm documentação da decisão (skill aplicada).
- **`quality-assurance`** (sensor 02 unit tests):
  verifica que testes cobrem pelo menos 1 caso de
  borda por função > 25 linhas.

## 🔗 Quem detecta violação

- `golangci-lint run` com `funlen { lines: 35 }` —
  falha em > 35.
- `eslint max-lines-per-function: [error, { max: 35 }]`
  (frontend) — falha em > 35.
- Code review manual: se função está em 26-35 sem
  justificativa, pedir pra aplicar skill.

---

## Referências

- [`../code-style.md`](../stack/code-style.md) §"Funções / Tamanho" (regra completa)
- [`../stack/backend.md`](../stack/backend.md) §"Anti-patterns" (v1.10.0: limite 35)
- [`../personas/backend-engineer.md`](../personas/backend-engineer.md) §3-5 (TDD + limites)
- [`../personas/solutions-architect.md`](../personas/solutions-architect.md) DoD (validação no DoD)
- `harness/contrib/design-decisions.md` ADR-0020 (decisão completa)
- `harness/templates/.golangci.yml` `funlen` config
