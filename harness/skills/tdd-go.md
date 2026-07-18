# Skill — tdd-go (TDD em Go)

> Skill para aplicar Test-Driven Development em Go, com table-driven
> tests, testify, e fakes in-memory.
> Usada por `backend-engineer`.

---

## Quando carregar

- Vai implementar uma função ou método novo.
- Vai corrigir um bug (TDD do regression test).
- Vai refatorar código existente.

---

## Os 3 passos (🔴 → 🟢 → 🛠️)

### 🔴 Red — escreva o teste de borda primeiro

1. Identifique **uma** unidade de comportamento.
2. Escreva um teste que **falhe** (porque o código não existe ou está
   errado).
3. Rode `go test ./<pacote>/...` e confirme que falha pelo motivo
   esperado.

```go
// internal/service/auth_test.go
package service_test

import (
    "context"
    "errors"
    "testing"

    "github.com/stretchr/testify/require"

    "my-service/internal/domain"
    "my-service/internal/repository"
    "my-service/internal/service"
)

func TestAuthService_Login(t *testing.T) {
    t.Run("returns token for valid credentials", func(t *testing.T) {
        // arrange
        repo := &fakeUserRepo{users: map[string]repository.User{
            "user@example.com": {Email: "user@example.com", PasswordHash: "$2a$10$validhash"},
        }}
        jwt := &fakeJWT{token: "valid.jwt.token"}
        svc := service.NewAuthService(repo, jwt)

        // act
        token, err := svc.Login(context.Background(), "user@example.com", "secret")

        // assert
        require.NoError(t, err)
        require.Equal(t, "valid.jwt.token", token)
    })
}
```

### 🟢 Green — implemente o mínimo

1. Escreva **apenas** o código que faz o teste passar.
2. **Não** antecipe features (YAGNI).
3. Rode `go test` e confirme verde.

```go
// internal/service/auth.go
package service

import (
    "context"
    "errors"

    "my-service/internal/domain"
    "my-service/internal/repository"
)

var (
    ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
    repo repository.UserRepository
    jwt  JWT
}

func NewAuthService(repo repository.UserRepository, jwt JWT) *AuthService {
    return &AuthService{repo: repo, jwt: jwt}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
    user, err := s.repo.GetByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return "", ErrInvalidCredentials
        }
        return "", err
    }
    if !checkPassword(password, user.PasswordHash) {
        return "", ErrInvalidCredentials
    }
    return s.jwt.Sign(user.ID)
}
```

### 🛠️ Refactor — limpe mantendo o teste verde

1. Extraia helpers, renomeie, organize.
2. **Sem** mudar comportamento.
3. Rode `go test` a cada mudança.

---

## Padrão: table-driven

Quando há múltiplos casos:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid simple", "user@example.com", false},
        {"valid subdomain", "user@mail.example.com", false},
        {"missing @", "userexample.com", true},
        {"missing user", "@example.com", true},
        {"missing domain", "user@", true},
        {"empty", "", true},
        {"spaces", "  user@example.com  ", false},
        {"unicode local", "üser@example.com", false},
        {"max local (64)", strings.Repeat("a", 64) + "@example.com", false},
        {"too long local (65)", strings.Repeat("a", 65) + "@example.com", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.input)
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

## Padrão: fakes in-memory

Para repository / dependências externas:

```go
type fakeUserRepo struct {
    users map[string]repository.User
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (repository.User, error) {
    u, ok := f.users[email]
    if !ok {
        return repository.User{}, repository.ErrNotFound
    }
    return u, nil
}

func (f *fakeUserRepo) Create(_ context.Context, u repository.User) error {
    f.users[u.Email] = u
    return nil
}
```

> **Regra:** fakes in-memory > mocks gerados > mocks manuais. Use mocks
> apenas para HTTP client externo, broker, etc.

---

## Casos de borda obrigatórios

Todo teste de unidade deve cobrir pelo menos:

- ✅ **Happy path** (1+ caso).
- ✅ **Entrada nula/vazia** (nil, "", []).
- ✅ **Entrada no limite** (max, min, zero).
- ✅ **Erro de dependência** (DB retorna erro, timeout).
- ✅ **Estado inválido** (recurso não encontrado, conflito).

---

## require vs assert

```go
// require: precondições (para o teste se falhar)
require.NoError(t, err)
require.NotNil(t, user)

// assert: validações (roda todas, mostra todas as falhas)
assert.Equal(t, expected, actual)
assert.Len(t, list, 3)
```

---

## Subtests

```go
// Subtest para organizar
t.Run("group", func(t *testing.T) {
    t.Run("case 1", func(t *testing.T) { ... })
    t.Run("case 2", func(t *testing.T) { ... })
})

// Em paralelo (cuidado: shared state!)
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()  // só se não houver shared state
        // ...
    })
}
```

---

## Mock de fronteira externa

Use `mockery` ou `uber-go/mock`:

```go
//go:generate mockery --name=UserRepository --outpkg=mocks --output=mocks
type UserRepository interface {
    GetByEmail(ctx context.Context, email string) (User, error)
}
```

```go
func TestWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockUserRepository(ctrl)
    repo.EXPECT().GetByEmail(gomock.Any(), "user@example.com").
        Return(User{Email: "user@example.com"}, nil)

    // ...
}
```

---

## Comandos úteis

```bash
# Rodar testes com race detector
go test -race ./...

# Com coverage
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out -o coverage.html

# Shuffle (detecta dependência entre testes)
go test -shuffle=on ./...

# Verbose (ver subtests)
go test -v ./internal/service/...

# Fail-fast
go test -failfast ./...

# Específico
go test -run TestAuthService_Login ./...
```

---

## Anti-padrões

- ❌ Testar implementação (white-box) em vez de comportamento
  (black-box).
- ❌ Mock do próprio código (só mocke fronteiras).
- ❌ Test flaky (depende de tempo, ordem, sleep).
- ❌ Skip sem justificativa (`t.Skip("TODO")`).
- ❌ Cobertura 100% mas sem teste de borda.
- ❌ Setup gigante compartilhado (refatore).
- ❌ Teste que testa o mock, não o código.

---

## Quem carrega

- `backend-engineer`.
