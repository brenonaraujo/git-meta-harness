# Domain Experts — Exemplos de especialização

> O `domain-expert` é **sempre** especializado em um domínio
> específico. Esta pasta contém exemplos de especializações prontas
> para uso (ou como base para criar a sua).

---

## Como usar

1. **Escolha o exemplo mais próximo** do seu domínio.
2. **Copie** para `personas/domain-expert-<seu-dominio>.md`.
3. **Adapte** o glossário, regulamentação, edge cases, e referências.
4. **Mantenha** o formato de saída e a estrutura geral.
5. **Materialize** nos artefatos do tool (Claude / Hermes / Codex).
6. **Crie a label** `domain/<seu-dominio>` no repo.

> Ver [`../domain-expert.template.md`](../domain-expert.template.md)
> para o template genérico.

---

## Exemplos disponíveis

| Persona                              | Domínio                              | Quando usar                                          |
|--------------------------------------|--------------------------------------|------------------------------------------------------|
| [`domain-expert-banking.md`](./domain-expert-banking.md) | Fintech, pagamentos, Open Banking | Apps de pagamento, contas digitais, crédito, Pix, etc. |
| [`domain-expert-retail.md`](./domain-expert-retail.md)   | E-commerce, OMS, fulfillment        | Lojas virtuais, marketplaces, gestão de estoque.    |
| [`domain-expert-mandai.md`](./domain-expert-mandai.md)   | Domínio customizado (exemplo)        | Base para criar um domain-expert próprio.            |

---

## Lista de domínios comuns (sugestões para criar)

- `domain-expert-banking` — fintech
- `domain-expert-retail` — e-commerce
- `domain-expert-healthcare` — saúde (HL7, FHIR, HIPAA)
- `domain-expert-logistics` — logística e supply chain
- `domain-expert-edtech` — educação
- `domain-expert-gaming` — jogos
- `domain-expert-realestate` — imobiliário
- `domain-expert-legal` — jurídico
- `domain-expert-hr` — RH
- `domain-expert-saas` — SaaS genérico (multi-tenant, billing)
- `domain-expert-iot` — IoT
- `domain-expert-media` — mídia/streaming
- `domain-expert-gov` — governo (e-gov, licitações)
- `domain-expert-manufacturing` — indústria

> **Convenção:** sempre kebab-case, sem acentos, sem espaços. Use
> sufixos quando precisar degranular (ex.: `domain-expert-banking-pix`
> para um especialista apenas em Pix).

---

## Como o team-manager roteia

O `team-manager` decide qual `domain-expert-<x>` atribuir com base
em:

1. **Label `domain/<x>`** na issue (preferência).
2. **Análise do título/body** (palavras-chave do domínio).
3. **Pergunta explícita** ao autor (caso ambíguo).

```bash
# Se a issue já tem label:
gh issue view 42 --json labels | jq '.labels[].name'
# Se contém "domain/banking" → atribui a @domain-expert-banking

# Se não tem, o team-manager pergunta:
gh issue comment 42 --body "🤖 Esta issue parece ser do domínio
**`<x>`**, correto? Vou atribuir a `@domain-expert-<x>`."
```

---

## Quando ter múltiplos domain-experts no mesmo projeto

Se o projeto atravessa **múltiplos domínios**, o team-manager
atribui **um specialist por sub-issue** (não por issue-pai). Ex.:

- Projeto: e-commerce de alimentos.
- `domain-expert-retail` para fluxo de checkout.
- `domain-expert-logistics` para entrega.
- `domain-expert-payments` (sub-especialização de banking) para
  integração com gateway.
