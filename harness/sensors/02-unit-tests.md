# Sensor 02 — Unit Tests (TDD, table-driven, coverage)

> **Objetivo:** garantir que o código tem testes de borda escritos
> **antes** da implementação (TDD) e cobertura mínima aceitável.
> **Quando roda:** `pre-commit` (local) + CI (todo push/PR).
> **Falha → ação:** **bloqueia merge** se coverage < 80% nos pacotes
> alterados ou se testes falham.

---

## Estratégia

**TDD em 3 passos:**

1. **🔴 Red** — escreva o teste de borda **antes** do código.
2. **🟢 Green** — implemente o mínimo para o teste passar.
3. **🛠️ Refactor** — limpe mantendo o teste verde.

**Padrão de teste:**

- **Table-driven tests** com `testify/assert` ou `testify/require`.
- **Use `require` em precondições** (que devem parar o teste).
- **Use `assert` em validações** (que devem rodar todas para ver todas as falhas).
- **Mocks nas fronteiras** (DB, HTTP client externo, broker), **fakes
  em memória** para o resto. Não mocke seu próprio código.
- **Cobertura mínima:** 80% de **branches** (não LOC) nos pacotes
  alterados pelo PR.
- **Testes de borda obrigatórios:**
  - Entrada nula/vazia/máxima.
  - Erro de dependência (DB caído, timeout, 4xx/5xx).
  - Concorrência (se aplicável).
  - Estado inválido (ex.: usuário bloqueado tentando logar).

---

## Comandos exatos

### Backend (Go)

```bash
# Roda todos os testes com race detector + coverage
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Gera relatório HTML
go tool cover -html=coverage.out -o coverage.html

# Verifica cobertura por pacote (fail se < threshold)
go test -cover ./... | awk '/coverage:/ {gsub("%","",$2); if($2+0 < 80) {print "FAIL: "FILENAME" coverage "$2"%"; exit 1}}'

# Roda com shuffle (detecta dependência entre testes)
go test -shuffle=on ./...
```

### Frontend (Nuxt/TS)

```bash
pnpm test --run --coverage
pnpm test:e2e  # E2E (criado pelo QA)
```

---

## Exemplo de teste (table-driven, com testify)

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {name: "valid simple", email: "user@example.com", wantErr: false},
        {name: "valid subdomain", email: "user@mail.example.com", wantErr: false},
        {name: "missing @", email: "userexample.com", wantErr: true},
        {name: "missing user", email: "@example.com", wantErr: true},
        {name: "missing domain", email: "user@", wantErr: true},
        {name: "empty", email: "", wantErr: true},
        {name: "spaces", email: "  user@example.com  ", wantErr: false},
        {name: "unicode local", email: "üser@example.com", wantErr: false},
        {name: "very long local (max 64)", email: strings.Repeat("a", 64) + "@example.com", wantErr: false},
        {name: "too long local (65)", email: strings.Repeat("a", 65) + "@example.com", wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## Exemplo (componente Vue, Vitest)

```ts
import { describe, it, expect, beforeEach } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { setActivePinia, createPinia } from 'pinia'
import LoginForm from '~/components/feature/auth/LoginForm.vue'
import { useAuthStore } from '~/stores/auth'

describe('LoginForm', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('submits credentials and calls auth.login', async () => {
    const wrapper = await mountSuspended(LoginForm)
    const auth = useAuthStore()
    const loginSpy = vi.spyOn(auth, 'login').mockResolvedValue()

    await wrapper.find('input[name="email"]').setValue('user@example.com')
    await wrapper.find('input[name="password"]').setValue('secret')
    await wrapper.find('form').trigger('submit.prevent')

    expect(loginSpy).toHaveBeenCalledWith({
      email: 'user@example.com',
      password: 'secret',
    })
  })

  it('shows error when email is empty', async () => {
    const wrapper = await mountSuspended(LoginForm)
    await wrapper.find('form').trigger('submit.prevent')
    expect(wrapper.text()).toContain('Email is required')
  })
})
```

---

## Exit codes

- `0` — todos os testes passam, coverage ≥ 80% nos pacotes alterados.
- `1` — falha de teste ou coverage abaixo do threshold.

---

## Thresholds

| Métrica                       | Limite          |
|-------------------------------|-----------------|
| Cobertura de branch           | **≥ 80%** nos pacotes alterados |
| Cobertura de linha            | reportar, não bloquear |
| Testes flaky                  | **0** (rodar com `-shuffle=on` e `-race`) |
| Tempo total de testes (unit)  | **≤ 5 min** (alertar se > 10 min) |

---

## Onde pluga no pipeline

### CI (`.github/workflows/ci.yml`)

```yaml
test:
  name: Test + Coverage
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with: { go-version: '1.22' }
    - run: go test -race -shuffle=on -coverprofile=coverage.out -covermode=atomic ./...
    - name: Check coverage
      run: |
        go test -cover ./... | tee coverage.txt
        # Adicione seu checker de threshold aqui
    - uses: actions/upload-artifact@v4
      with:
        name: coverage-report
        path: coverage.html

    - uses: actions/setup-node@v4
      with: { node-version: '20' }
    - run: pnpm install --frozen-lockfile
    - run: pnpm test --run --coverage
    - uses: actions/upload-artifact@v4
      with:
        name: frontend-coverage
        path: coverage/
```

### Diff coverage (no PR)

Use `go-test-coverage` ou ação customizada para exigir coverage
mínima **apenas nos arquivos alterados** pelo PR.

---

## Falha típica & remediação

| Falha                                       | Como corrigir                                              |
|---------------------------------------------|------------------------------------------------------------|
| Coverage < 80%                              | Adicionar testes de borda. **Não** marcar `//nolint:cover`.|
| Test flaky                                  | Remover dependência de `time.Now()`, usar fakes.           |
| Test lento                                  | Marcar com `t.Short()` ou refatorar para usar fake.        |
| Mock quebrou após refactor                  | Regenerar mock (`mockery` / `mockgen`) ou ajustar.         |
| Race condition em teste                     | Adicionar `-race` e corrigir; usar `sync.Mutex` no código. |

---

## Quem roda

- **Local:** `backend-engineer`, `frontend-engineer` (TDD — teste
  primeiro).
- **CI:** workflow `ci.yml`.
- **Falha:** bloqueia merge.
