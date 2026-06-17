# 07 — Segurança (JWT + RBAC)

## Autenticação (JWT)

- **Access token** (curta duração, ex. 15min): JWT assinado (HS256) com claims
  `sub` (id do usuário), `roles`, `perms`, `exp`.
- **Refresh token** (longa duração, ex. 30 dias): opaco, com **hash** guardado em
  `iam.refresh_tokens`. Permite rotação e revogação (logout/roubo).
- Senhas: hash **bcrypt** em `senha_hash_usr` (nunca em texto puro).

### Fluxo

```
login(email,senha) ─► valida bcrypt ─► gera access(JWT) + refresh(opaco)
   │
   ├─ requisições usam: Authorization: Bearer <access>
   │
refresh(refresh_token) ─► confere hash + validade + !revogado ─► novo access
   │                                                           └─ rotaciona refresh
logout ─► marca refresh.revogado = true
```

## Autorização (RBAC)

Modelo **usuário → papéis → permissões**:

```
usuario ──< usuario_papeis >── papel ──< papel_permissoes >── permissao
                                                               (codigo: "recurso:acao")
```

- Permissões nomeadas `recurso:acao` — ex.: `vendas:write`, `clientes:read`,
  `iam:admin`.
- No login, as permissões efetivas do usuário são resolvidas e embutidas no JWT
  (claim `perms`), evitando ida ao banco a cada request.
- Cada rota declara a permissão exigida; um **middleware** (`platform/auth`)
  verifica se o token a contém.

### Papéis sugeridos

| Papel | Permissões |
|-------|------------|
| `ADMIN` | todas (incl. `iam:admin`) |
| `VENDEDOR` | `vendas:*`, `clientes:*`, `catalogo:read`, `estoque:read`, `relatorios:read` |
| `ESTOQUISTA` | `compras:*`, `estoque:*`, `catalogo:*`, `fornecedores:*`, `relatorios:read` |

### Middleware (pseudo-código)

```go
func RequirePerm(perm string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := auth.FromContext(r.Context())     // setado pelo Authenticate
            if !claims.Has(perm) {
                httpserver.Error(w, http.StatusForbidden, "PERM_DENIED", "...")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

r.With(auth.Authenticate, auth.RequirePerm("vendas:write")).
    Post("/vendas/{id}/confirmar", handler.Confirmar)
```

## Boas práticas adotadas

- `JWT_SECRET` via env; **trocar** o seed `admin123` em produção.
- `citext` para e-mails (login case-insensitive).
- Princípio do menor privilégio: rotas sempre exigem permissão específica.
- Refresh com rotação + revogação; auditoria via `created_at`/`updated_at`.
- (Recomendado p/ produção) rate limiting no `/auth/login` e HTTPS obrigatório.
