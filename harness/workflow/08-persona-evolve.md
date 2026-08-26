# Workflow 08 — Persona evolve (instanciar + melhorar)

> Playbook curto do `team-manager`: quando criar um
> `domain-expert-<x>`, quando rodar `gmh evolve`, e
> quando parar. **Condição de parada = humano valida.**
> Persona files **não** mudam sem essa validação.
>
> Detalhe conceitual: [`docs/EVOLVE.md`](../../docs/EVOLVE.md).
> Decisão: ADR-0030.

---

## Duas camadas (não misturar)

| Camada | O que é | O que o team-manager **não** pede a ela |
|---|---|---|
| **Hermes** | Runtime OS/tool harness (terminal, fs, `gh`, browsers, memória de sessão) | Decidir quais personas existem |
| **git-meta-harness** | Delivery harness que **materializa** personas, skills, sensors, roteamento GitHub | Substituir as tools do Hermes |

A saída do `gmh` é **contexto** (agents / skills / tools)
**deste** projeto. Hermes executa esse contexto.

---

## Quando criar um domain-expert

Criar **na seed / adopt** (ou no primeiro `domain/<x>`
sem persona correspondente):

```bash
gmh personas create --domain <x> --context "..."
gmh personas create --domain <x> --from-spec spec.md
```

**Invariante:** nunca um `domain-expert` genérico
(ADR-0003). Recusar `--domain generic`, `--domain
domain-expert`, ou domínio vazio.

**Não criar** se:

- a issue é `type/technical`, `type/infra`,
  `type/tech-debt` ou `type/docs` (smart routing pula
  domain-expert — ver `00-issue-lifecycle.md` §0);
- já existe `harness/personas/domain-expert-<slug>.md`
  para aquele domínio.

Depois de criar: `gmh memory write` para persistir o
snapshot em `harness/memory/snapshot.json`.

---

## Quando evoluir

Rodar `gmh evolve` quando o histórico de issues +
comentários mostrar **lacunas recorrentes** em AC,
DoD ou skills — não a cada issue.

```bash
gmh evolve --from-dir ./fixtures --apply
```

`--from-dir` lê fixtures de comentários (`*.json` ou
`comments.json`). `--apply` grava **somente**
`harness/memory/traces/<utc>/` (`comments.json`,
`proposal.md`, `PROMPT.md`). **Não** sobrescreve
personas.

Abrir `PROMPT.md` no Hermes team-manager (camada OS/
tool). O proposer é o agente; o CLI só materializa o
prompt. **Não** chamar o proposer Python do Stanford
IRIS.

---

## Condição de parada

```
proposta escrita
    → humano lê proposal.md / PROMPT.md
    → "validado"  → team-manager (ou humano) edita
                    persona/skill files
    → rejeição    → traces ficam; persona files
                    intactos; loop para
```

O loop **para** quando o humano valida ou rejeita.
Enquanto isso não acontece:

- não editar `harness/personas/*.md`;
- não tratar `--apply` como licença para patch;
- não fechar a issue de evolve.

Mesma regra do ciclo de entrega: sensores verdes **não**
fecham o loop; o humano fecha.

---

## Sequência mínima

```bash
gmh personas create --domain <x> --from-spec spec.md
gmh memory write
gmh evolve --from-dir ./fixtures --apply
# humano valida → só então alterar persona files
```

---

## Ver também

- [`docs/EVOLVE.md`](../../docs/EVOLVE.md)
- [`00-issue-lifecycle.md`](./00-issue-lifecycle.md) §0
  (quando o domain-expert entra no fluxo)
- ADR-0003 (sempre `domain-expert-<domínio>`)
- ADR-0030 (instantiate + evolve from issue traces)
