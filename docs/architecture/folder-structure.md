# 03 — Estrutura de Pastas

## Visão do repositório

```
. (raiz do projeto)
├── docker-compose.yml        # orquestração: db + migrate + api
├── backend/
│   ├── cmd/
│   │   ├── api/              # entrypoint do monólito modular (HTTP)
│   │   │   └── main.go
│   │   └── migrate/          # runner de migrations (opcional p/ embed)
│   │       └── main.go
│   ├── internal/
│   │   ├── platform/         # infraestrutura COMPARTILHADA (sem regra de negócio)
│   │   │   ├── config/       #   carregamento de .env / env vars
│   │   │   ├── database/     #   pool pgx, transações, helpers
│   │   │   ├── httpserver/   #   servidor chi, middlewares globais, error mapping
│   │   │   ├── auth/         #   emissão/validação de JWT, middleware RBAC
│   │   │   └── events/       #   barramento de eventos (in-process hoje)
│   │   └── modules/          # um pacote por BOUNDED CONTEXT (isolado)
│   │       ├── iam/
│   │       ├── clientes/
│   │       ├── fornecedores/
│   │       ├── catalogo/
│   │       ├── compras/
│   │       ├── vendas/
│   │       └── estoque/
│   ├── migrations/           # DDL versionada (.up.sql / .down.sql)
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   ├── Makefile
│   └── .env.example
├── docs/
└── frontend/
```

## Anatomia de um módulo (hexagonal)

Todo módulo segue o mesmo formato. Exemplo com `catalogo`:

```
internal/modules/catalogo/
├── domain/                   # CORE — sem dependências de infra
│   ├── produto.go            #   entidade Produto + invariantes (custo<venda, etc.)
│   ├── categoria.go
│   └── errors.go             #   erros de domínio (ErrMargemInvalida, ...)
├── application/              # CASOS DE USO
│   ├── service.go            #   orquestra domínio + ports de saída
│   └── dto.go                #   comandos/queries de entrada e saída
├── ports/                    # INTERFACES (contratos)
│   ├── inbound.go            #   ProdutoService (oferecido ao mundo)
│   └── outbound.go           #   ProdutoRepository, EstoqueReader (necessidades)
├── adapters/
│   ├── inbound/
│   │   └── http/             #   handlers chi -> chamam ports.inbound
│   │       ├── handler.go
│   │       └── router.go
│   └── outbound/
│       └── postgres/         #   implementa ports.outbound com pgx
│           └── produto_repo.go
└── module.go                 # WIRING: instancia repos, service e router (DI)
```

### Por que essa separação?

- **`domain/`** é testável sem banco nem HTTP — testes de regra rodam em memória.
- **`ports/`** torna explícito *o que o módulo oferece* e *o que ele exige*.
- **`adapters/`** é descartável/substituível: trocar Postgres por outro banco,
  ou HTTP por gRPC, não toca no domínio.
- **`module.go`** é o único ponto que conhece as implementações concretas e as
  conecta — facilita extrair o módulo para um `main.go` próprio depois.

## Regras de import (isolamento)

- ✅ `adapters` → `application` → `ports` → `domain` (sempre para dentro).
- ✅ `module.go` → tudo do próprio módulo + `platform`.
- ❌ `modules/vendas/...` **não** importa `modules/estoque/application`.
  Se `vendas` precisa de estoque, declara uma porta em
  `modules/vendas/ports/outbound.go` e o `main` injeta a implementação que vem
  de `estoque`. Assim o acoplamento é por **interface**, não por pacote.
- ✅ Qualquer módulo pode importar `internal/platform` (infra neutra).

## Wiring no `cmd/api/main.go`

```go
// pseudo-código de montagem
db   := database.NewPool(cfg)
bus  := events.NewInProcessBus()

estoqueMod  := estoque.New(db, bus)
catalogoMod := catalogo.New(db)
vendasMod   := vendas.New(db, bus, estoqueMod.Writer()) // injeção por porta

r := httpserver.NewRouter(cfg)
r.Mount("/api/v1/catalogo", catalogoMod.Router())
r.Mount("/api/v1/vendas",   vendasMod.Router())
// ...
```

Quando `vendas` virar microsserviço, `estoqueMod.Writer()` é trocado por um
client HTTP/gRPC (ou publisher de evento) — o módulo `vendas` não muda.
