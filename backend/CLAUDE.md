# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Serviço Go do ERP: monólito modular com **arquitetura hexagonal**, preparado para extração em microsserviços. Código e comentários em **português** — siga o padrão. Veja a visão geral do repositório em [../CLAUDE.md](../CLAUDE.md).

## Commands

O `Makefile` fica na **raiz do projeto** (orquestra backend + frontend + infra) e
faz `-include backend/.env`. Rode os alvos a partir da raiz; exija `backend/.env`
(`cp backend/.env.example backend/.env`).

```bash
make be-run         # go run ./cmd/api        (API em :8080)
make be-build       # go build -o bin/api ./cmd/api  → backend/bin/api
make be-test        # go test ./...           (toda a suíte do backend)
make be-vet         # go vet ./...
make migrate-up / migrate-down / migrate-create name=add_xyz
make reset          # DROP total + migrate-up + seed
make up / down / logs   # docker compose (db + migrate + api + frontend)

# pacote ou teste único: go test direto em backend/
cd backend && go test ./internal/modules/clientes/domain
cd backend && go test -run TestValidarCPF ./internal/modules/clientes/domain
```

Não há linter configurado além de `go vet`; mantenha `gofmt` limpo.

## Estrutura de um módulo (bounded context)

Cada domínio vive em `internal/modules/<dominio>/` com camadas cujas dependências apontam **só para dentro** (`domain` não importa nada de infra). Use `internal/modules/clientes/` como referência canônica ao criar um novo módulo:

| Camada | Papel | Não pode conhecer |
|--------|-------|-------------------|
| `domain/` | Entidade raiz, invariantes (`Validar`), erros sentinela (`errors.go`) | HTTP, SQL, pgx |
| `ports/inbound.go` | Interface do caso de uso + structs de Input que o módulo **oferece** | — |
| `ports/outbound.go` | Interfaces que o módulo **exige** (repo, gateways) | implementações |
| `application/` | `Service` que implementa a porta inbound e orquestra domínio + portas outbound | HTTP, SQL |
| `adapters/inbound/http/` | `handler.go` (JSON↔serviço, DTOs) + `router.go` (rotas + RBAC) | regras de negócio |
| `adapters/outbound/postgres/`, `.../cep/` | implementam as portas outbound | regras de negócio |
| `module.go` | **Composition root (DI)**: instancia concretos, monta o serviço, expõe `Router()` | — |

Convenções obrigatórias ao seguir o molde:
- Confirme a conformidade em tempo de compilação: `var _ ports.X = (*Impl)(nil)`.
- Erros são **sentinelas de domínio** (`domain/errors.go`); o handler os traduz para HTTP em `writeDomainError` (ex.: `ErrNaoEncontrado`→404, `ErrCPFJaCadastrado`→409, validações→422). Propague erros, não status HTTP, fora do adaptador.
- O handler usa DTOs próprios (`clienteRequest`/`clienteResponse`) — nunca serialize a entidade de domínio direto.
- Para registrar o módulo: exponha `New(...)` em `module.go` e monte em [cmd/api/main.go](cmd/api/main.go) sob `/api/v1` (há exemplos comentados).

## Autorização (platform/auth)

JWT HS256. Toda rota protegida é embrulhada por `authMgr.Authenticate` (injeta `Claims` no contexto) + `auth.RequirePerm("recurso:acao")`. As permissões viajam **dentro do token** (claim `perms`) — não há consulta ao banco por request. Padrão de nomenclatura: `<recurso>:read` / `<recurso>:write`.

## Resiliência (platform/resilience)

Adaptadores outbound que chamam **APIs externas** devem embrulhar a chamada em uma `resilience.Policy`, composta como `Retry → CircuitBreaker → Bulkhead → fn`. Veja a montagem em [module.go](internal/modules/clientes/module.go) e o uso em [cep/viacep.go](internal/modules/clientes/adapters/outbound/cep/viacep.go). Erros **não-retriáveis** devem ser marcados com `resilience.Permanent(err)`; erros transitórios (timeouts, 5xx) são deixados crus para o Retry reagir.

## Banco de dados

**1 schema Postgres por bounded context**, **sem foreign keys entre schemas** — integridade referencial cross-context é responsabilidade da aplicação (futuro: eventos/sagas). Repositórios falam apenas com seu próprio schema (`clientes.clientes`, etc.) e mapeiam colunas `*_<sufixo>` (ex.: `id_cli`, `nome_cli`) ↔ campos da entidade.

Migrations em `migrations/`, pares `NNNNNN_nome.up.sql`/`.down.sql` (golang-migrate, sequenciais). `000001_init` cria extensões + schemas + `set_updated_at()`; `000009_seed` popula dados iniciais (login `admin@loja.local`/`admin123`); `000010_seed_demo` adiciona dados de demonstração idempotentes. Toda mudança de schema é uma nova migration — nunca edite uma já aplicada. Em dev use os alvos `make migrate-*` (CLI `migrate/migrate`); em produção, o runner Go embarcado `cmd/migrate` roda os mesmos arquivos como `/app/migrate up` (pre-deploy do Railway) e via `scripts/supabase-setup.sh`.

## Config

`config.Load()` lê `.env` (opcional) e env vars com defaults de desenvolvimento embutidos. Toda nova configuração entra na struct `Config` e é lida via `getenv`/`getdur`.
