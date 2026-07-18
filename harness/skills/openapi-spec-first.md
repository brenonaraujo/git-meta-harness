# Skill — openapi-spec-first

> Skill para desenvolver APIs Go seguindo o padrão **spec-first** com
> OpenAPI 3.1 + `oapi-codegen` (ou `ogen`).
> Usada por `backend-engineer` (implementação) e `solutions-architect`
> (design).

---

## Quando carregar

- Vai criar um endpoint novo.
- Vai modificar contrato de endpoint existente.
- Vai regenerar tipos do OpenAPI.

---

## Workflow spec-first

```
1. solutions-architect edita api/openapi.yaml
2. backend-engineer regenera internal/api/openapi.gen.go
3. backend-engineer implementa contra o código gerado
4. CI roda openapi-diff (detecta breaking changes)
5. CI roda schemathesis (valida runtime)
```

---

## Estrutura do contrato

```
api/
└── openapi.yaml       # contrato canônico
internal/
└── api/
    └── openapi.gen.go # gerado, NUNCA editar manualmente
```

---

## Comandos

### Regenerar (Make target)

```makefile
# Makefile
oas:
	oapi-codegen --config=codegen.yaml api/openapi.yaml
```

`codegen.yaml`:

```yaml
package: api
output: internal/api/openapi.gen.go
generate:
  models: true
  gin-server: true
  embedded-spec: true
  strict-server: true
```

### Validar sintaxe

```bash
# Spectral
npx @stoplight/spectral-cli lint api/openapi.yaml

# openapi-validator (Node)
npx @apidevtools/swagger-cli validate api/openapi.yaml
```

### Diff entre branches (no CI)

```bash
oasdiff diff origin/main HEAD --format json > diff.json
# Breaking change? exit code 1
```

### Validação runtime (Schemathesis)

```bash
pip install schemathesis
# Subir o serviço antes
docker compose -f deploy/docker-compose.yml up -d
schemathesis run http://localhost:8080/api/v1 \
  --schema api/openapi.yaml \
  --checks all
```

---

## Exemplo de OpenAPI

```yaml
openapi: 3.1.0
info:
  title: Auth Service
  version: 0.1.0
  contact:
    name: Time Auth
    email: auth@example.com
  license:
    name: MIT
servers:
  - url: http://localhost:8080/api/v1
    description: Local
  - url: https://auth.example.com/api/v1
    description: Production

paths:
  /auth/login:
    post:
      operationId: Login
      tags: [auth]
      summary: Login with email and password
      description: |
        Authenticates a user and returns a JWT token.
        Returns 401 on invalid credentials.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/TooManyRequests'
        '500':
          $ref: '#/components/responses/InternalError'

  /auth/me:
    get:
      operationId: GetCurrentUser
      tags: [auth]
      summary: Get the current authenticated user
      security:
        - bearerAuth: []
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '401':
          $ref: '#/components/responses/Unauthorized'

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    LoginRequest:
      type: object
      required: [email, password]
      properties:
        email:
          type: string
          format: email
          example: user@example.com
        password:
          type: string
          minLength: 8
          example: secret123
    LoginResponse:
      type: object
      required: [token, expires_at, user]
      properties:
        token:
          type: string
          description: JWT token
        expires_at:
          type: string
          format: date-time
        user:
          $ref: '#/components/schemas/User'
    User:
      type: object
      required: [id, email, name, created_at]
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
          format: email
        name:
          type: string
        role:
          type: string
          enum: [user, admin]
        created_at:
          type: string
          format: date-time

    Error:
      type: object
      required: [code, message]
      properties:
        code:
          type: string
          description: Machine-readable error code
          example: invalid_credentials
        message:
          type: string
          description: Human-readable error message
          example: Invalid email or password

  responses:
    BadRequest:
      description: Bad request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Unauthorized:
      description: Unauthorized
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    TooManyRequests:
      description: Too many requests
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    InternalError:
      description: Internal server error
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
```

---

## Implementação (handler)

```go
// internal/handler/auth.go
package handler

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"

    "my-service/internal/api"
    "my-service/internal/service"
)

type AuthHandler struct {
    svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
    return &AuthHandler{svc: svc}
}

// Login implements api.ServerInterface
func (h *AuthHandler) Login(c *gin.Context) {
    var req api.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, api.Error{
            Code: "invalid_request", Message: err.Error(),
        })
        return
    }
    res, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        if errors.Is(err, service.ErrInvalidCredentials) {
            c.JSON(http.StatusUnauthorized, api.Error{
                Code: "invalid_credentials", Message: "Invalid email or password",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, api.Error{
            Code: "internal_error", Message: "Internal error",
        })
        return
    }
    c.JSON(http.StatusOK, res)
}
```

---

## Boas práticas

- **Operation IDs claros**: `Login`, `GetCurrentUser`, `CreateOrder`
  (camelCase, sem verbos redundantes).
- **Tags** agrupam por domínio (`auth`, `users`, `orders`).
- **Responses** sempre com schema (mesmo Error).
- **Exemplos** em todos os campos.
- **Descriptions** em operações e schemas (human-readable).
- **Status codes** semânticos:
  - `200` OK
  - `201` Created
  - `204` No Content
  - `400` Bad Request (validação)
  - `401` Unauthorized (sem auth)
  - `403` Forbidden (sem permissão)
  - `404` Not Found
  - `409` Conflict
  - `422` Unprocessable Entity (validação semântica)
  - `429` Too Many Requests
  - `500` Internal Server Error
  - `503` Service Unavailable

---

## Detecção de breaking change (no CI)

```yaml
# .github/workflows/ci.yml
contract:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with: { fetch-depth: 0 }
    - uses: oasdiff/oasdiff-action@v1
      with:
        base: origin/${{ github.base_ref }}
        revision: HEAD
        fail-on: breaking
```

**Breaking changes** (alguns exemplos):

- Remover endpoint.
- Remover campo obrigatório.
- Renomear campo.
- Mudar tipo de campo.
- Mudar status code de sucesso para erro (ou vice-versa).
- Adicionar campo obrigatório.

**Não-breaking:**

- Adicionar endpoint.
- Adicionar campo opcional.
- Adicionar response code novo.
- Adicionar enum value (em alguns casos, breaking para cliente).

---

## Quem carrega

- `solutions-architect` (design do contrato).
- `backend-engineer` (implementação).
