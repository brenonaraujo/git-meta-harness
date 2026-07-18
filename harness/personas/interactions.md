# Interactions — Matriz de comunicação entre personas

> **Quem pode pedir o quê pra quem.** Este documento é o **protocolo
> de comunicação** do meta-harness. Ele resolve o problema de
> personas extrapolando seus papéis (ex.: `solutions-architect`
> atribuindo builders, `domain-expert` criando branch, etc.).

---

## 1. Regra de ouro

> **Cada persona tem um papel, e somente esse papel.** Comunicar-se
> com outra persona é permitido **apenas dentro do escopo do seu
> papel**. Tudo que está fora do seu papel deve ser **reportado
> via issue** (comentário) para o `team-manager` decidir.

| Persona              | Pode FAZER (dentro do papel)                          | Pode PEDIR ao team-manager   | Pode FALAR diretamente com (peer) |
|----------------------|------------------------------------------------------|------------------------------|-----------------------------------|
| `team-manager`       | Atribuir, **criar branch de feature/fix/chore**, mover labels, comentar, criar sub-issues, fechar issues, merge, tag | (n/a — é ele)                | Qualquer um (peer review)         |
| `domain-expert-<x>`  | Refinar história, postar ACs + edge cases            | "Atribuir o próximo passo"   | (raro — só em clarificações)       |
| `solutions-architect`| Definir DoD, validar 12-factor, propor padrões       | "Atribuir a builder X"       | (raro — só em decisões cruzadas)   |
| `backend-engineer`   | **Clonar a branch** (criada pelo team-manager), implementar, testar, abrir PR | "Validar com QA"            | `frontend-engineer` (mesma branch) |
| `frontend-engineer`  | **Clonar a branch** (criada pelo team-manager), implementar, testar, abrir PR | "Validar com QA"            | `backend-engineer` (mesma branch)  |
| `quality-assurance`  | Rodar sensores, smoke/load, aprovar ou reprovar      | "Merge e release"           | (raro — só em bugs achados)        |
| `devops-engineer`     | **Criar branch de release** (`release/vX.Y.Z`), manter pipeline, fazer release | (n/a — release é dele) | (raro)                             |
| `usuário`            | Validar snapshot, comentar, fechar                   | "Revalidar isso"            | (n/a)                              |

> **Linha vermelha do `team-manager`:** ele **NÃO escreve código
> de feature**. Criar branch é orquestração (você decide **onde**
> o trabalho vai acontecer); escrever código é engenharia. Ver
> ADR-0006.

---

## 2. Quem **NÃO** pode fazer o quê (anti-padrões)

> Estas ações são **bloqueadas** pelo invariante 15 do `AGENTS.md` §8
> e pelo sensor de revisão (ver `skills/github-code-review.md`).

| Persona              | NÃO pode...                                                                 | Por quê                                              |
|----------------------|-----------------------------------------------------------------------------|------------------------------------------------------|
| `team-manager`       | **Escrever código de feature** (essa é a única linha vermelha)               | Foco em orquestração, não em implementação.         |
| `team-manager`       | **Aprovar PR tecnicamente** (apenas merge após validação)                    | Aprovação técnica é dos reviewers + QA.              |
| `team-manager`       | **Rodar sensores, testes, builds**                                          | Isso é trabalho de QA / devops.                      |
| `domain-expert-<x>`  | **Atribuir para outra persona** (ex.: "@backend-engineer faz X")            | Atribuição é exclusiva do team-manager.              |
| `domain-expert-<x>`  | **Criar branch** ou **definir tecnologia**                                  | É especialista de domínio, não engenheiro.          |
| `solutions-architect`| **Atribuir diretamente** a outra persona (ex.: "@frontend, faz X")          | Quem atribui é o team-manager; SA só define DoD.     |
| `solutions-architect`| **Criar branch de feature** (ou fazer merge)                                | Criar branch é do team-manager; SA não implementa.   |
| `solutions-architect`| **Mudar o tipo da issue** (ex.: classificar como `type/infra`)              | Tipo é decidido pelo team-manager na triagem.        |
| `backend-engineer`   | **Criar branch** (apenas clonar a que o team-manager criou)                  | Branch é responsabilidade do team-manager.           |
| `backend-engineer`   | **Fechar issue**                                                            | Quem fecha é o team-manager.                         |
| `backend-engineer`   | **Aprovar o próprio PR**                                                     | Aprovação é de outro builder/QA.                    |
| `frontend-engineer`  | **Criar branch** (apenas clonar a que o team-manager criou)                  | Branch é responsabilidade do team-manager.           |
| `frontend-engineer`  | **Fechar issue**                                                            | Quem fecha é o team-manager.                         |
| `quality-assurance`  | **Fechar issue** ou **merge na main**                                       | QA aprova/reprova; team-manager mergeia.             |
| `quality-assurance`  | **Modificar código de feature** (só testa)                                  | QA não implementa.                                   |
| `devops-engineer`    | **Criar branch de feature** (apenas `release/vX.Y.Z` e `hotfix/`)            | Branch de feature é do team-manager.                 |
| qualquer um          | **Fechar issue sem validação do usuário**                                    | Invariante 10 do AGENTS.md §8.                       |

---

## 3. Protocolo de handoff (issue-mãe → sub-issues)

> Toda transição de uma persona pra outra **passa pelo
> `team-manager`**. Isso garante rastreabilidade.

```
Issue-mãe #42 (criada)
  ↓
team-manager tria (label triage + type/<x> + domain/<y>)
  ↓
team-manager atribui ao próximo (com briefing explícito)
  ↓
[Próxima persona] faz seu trabalho
  ↓
[Próxima persona] posta resultado + move label de saída
  ↓
[Próxima persona] comenta "🤖 pronto, próximo é X"
  ↓
team-manager: valida, atribui ao próximo OU fecha (se fim)
```

**Quem pode atribuir:** apenas `team-manager`.
**Quem pode comentar "pronto, próximo é X":** qualquer persona, mas
isso é apenas **sinalização** — o team-manager decide o que fazer.

---

## 4. Handoff entre personas de engenharia (mesma issue)

> Quando `backend-engineer` e `frontend-engineer` trabalham na
> **mesma** feature (issue única), eles compartilham a **mesma
> branch**. Quem cria a branch?

### Regra

- O **`team-manager` cria a branch** (uma única vez) e passa o
  nome no briefing.
- Ambos os builders (`backend-engineer` e `frontend-engineer`)
  fazem checkout da **mesma** branch criada.
- **Nunca** um builder cria a própria branch — sempre recebe do
  team-manager.

```bash
# team-manager (uma vez):
git checkout main
git pull origin main
git checkout -b feature/42-checkout-pix
git push -u origin feature/42-checkout-pix
gh issue comment 42 --body "🤖 **team-manager → @backend-engineer
e @frontend-engineer**

Branch: \`feature/42-checkout-pix\` (criada). Clonem e
implementem conforme o DoD."

# backend-engineer:
git fetch origin
git checkout feature/42-checkout-pix
# implementar backend + commitar

# frontend-engineer:
git fetch origin
git checkout feature/42-checkout-pix
# implementar frontend + commitar
```

> Em caso de dúvida, **comente na issue** — o team-manager
> coordena. **Nunca** crie duas branches para a mesma issue.

---

## 5. Exemplo de handoff correto vs errado

### ❌ Errado (visto no teste do bootstrap)

> `solutions-architect` posta: "Atribuir a frontend-engineer (label
> ready → in-progress após o team-manager criar a branch
> feature/1-bootstrap-skeleton)."

**Problemas:**
- `solutions-architect` está **atribuindo** (deveria ser
  team-manager).
- `solutions-architect` está **mencionando nome de branch** (não
  é responsabilidade dele).

### ✅ Correto

> `solutions-architect` posta: "DoD está pronto. A próxima etapa é
> o builder implementar a estrutura do projeto conforme definido.
> Não há pendências."
>
> `team-manager` lê, **cria a branch `feature/1-bootstrap-skeleton`**,
> atribui a `@frontend-engineer` com briefing: "Você assume.
> Branch: `feature/1-bootstrap-skeleton` (criada por mim). Clone,
> implemente o esqueleto conforme DoD do @solutions-architect, e
> abra PR."
>
> `@frontend-engineer` clona a branch, implementa, abre PR, posta
> "🤖 PR #2 aberto, label in-review aplicada, sensores locais
> verdes."

---

## 6. Quem fecha o quê

| Quem fecha            | O quê fecha                                              |
|-----------------------|----------------------------------------------------------|
| `team-manager`        | Issue (issue-mãe ou sub-issue), após validação do usuário |
| `backend-engineer`    | Sub-issue (se o team-manager permitir — raro)            |
| `frontend-engineer`   | Sub-issue (se o team-manager permitir — raro)            |
| `quality-assurance`   | **Nada.** QA aprova/reprova, não fecha.                  |
| `solutions-architect` | **Nada.**                                                 |
| `domain-expert`       | **Nada.**                                                 |
| `devops-engineer`     | PR de release (via `gh pr merge`), mas a **issue** é fechada pelo team-manager |
| `usuário`             | **Nada** (mas pode pedir fechamento)                      |

---

## 7. Detecção de violação (sensor automático)

> Esta matriz vira um **sensor** que pode ser executado no CI
> para detectar violações:

```bash
# scripts/check-interactions.sh (esboço)
# Para cada PR:
# - Identificar o autor.
# - Verificar se o autor violou alguma regra desta matriz.
# - Ex.: se autor é @domain-expert, e o PR cria branch → ❌
# - Ex.: se autor é @team-manager, e o PR tem código de feature → ❌
# (heurística simples; não exaustiva)
```

Ver [`sensors/09-interactions-audit.md`](../sensors/09-interactions-audit.md)
quando for implementado.

---

## 8. Quem carrega

Todas as personas devem carregar este documento. Ele é a fonte
de verdade sobre **quem pode fazer o quê** e **quem pode falar
com quem**.

Em caso de dúvida sobre "isso é minha responsabilidade?", sempre
consulte esta matriz primeiro; em caso de ambiguidade, **pergunte
ao `team-manager`** (que tem a visão completa).
