<!-- domain-expert-refinement.md (v1.13.0)
     Comentário canônico de refinamento do domain-expert.
     SEMPRE use este template (copie e preencha) — sensor 13
     `feature-flow` valida que o comentário tem pelo menos
     1 AC e 1 edge case (BLOQUEANTE).
     Ver skill `domain-refinement` + invariante 24. -->

## 🎯 Refinamento — `<domain-expert-<domínio>>`

> **Issue:** #<id>
> **Tipo:** `type/feature` / `type/bug`
> **Domínio:** `domain/<nome>`

### Persona (Quem se beneficia)

```
Quem é o usuário? Que papel ele tem? O que ele está tentando
fazer? Em que contexto?
```

### Comportamento esperado (O que precisa acontecer)

```
Descreva o comportamento em termos de domínio, SEM citar UI
nem tecnologia. Use a Cerca de Design (§4.1.1) e a Cerca
Técnica (§4.1.2) do team-manager.
```

### Por que importa (Valor de negócio)

```
Qual problema de negócio resolve? Qual SLO/SLA esperado?
Que regulamentação se aplica (LGPD, PCI-DSS, etc)?
```

### ✅ Critérios de aceite (ACs)

1. AC-1: ...
2. AC-2: ...
3. AC-3: ...
(mínimo 1 — sensor 13 valida; recomendado 5-12 por issue)

### ⚠️ Edge cases do domínio

1. EC-1: ... (input inválido, race condition, conflito de estado, etc)
2. EC-2: ...
(mínimo 1 — sensor 13 valida; recomendado 3-8 por issue)

### 🔗 Dependências

- Depende de #X (issue/blocker externo)
- Bloqueia #Y (não pode iniciar sem este)
- Nenhuma

### 🧪 Validação

Como QA/builder vai verificar? (Critérios objetivos: dados
de entrada, saída esperada, condições de erro).

---

> **Próximo passo:** quando terminar, adicionar a label
> `refined` à issue. O `team-manager` então aciona o
> `solutions-architect` para definir o DoD técnico.
