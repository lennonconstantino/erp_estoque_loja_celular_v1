# Backend — ERP Estoque

Serviço HTTP em **Go 1.25** que implementa o ERP de estoque. Estruturado como
**monólito modular** com **arquitetura hexagonal** (Ports & Adapters): cada
domínio é isolado e pode ser extraído para um microsserviço próprio sem
reescrever a lógica de negócio.

- **Roteamento:** [go-chi/chi](https://github.com/go-chi/chi)
- **Banco:** PostgreSQL via [pgx](https://github.com/jackc/pgx) (1 schema por bounded context)
- **Auth:** JWT HS256 + RBAC por permissões (`recurso:acao`)
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate)

> **Estado atual:** apenas o módulo `clientes` está implementado em Go. Os demais
> domínios (`iam`, `fornecedores`, `catalogo`, `compras`, `vendas`, `estoque`)
> já têm schema (migrations) e documentação, e seguem o mesmo molde hexagonal.

## Pré-requisitos

- Go 1.25+
- PostgreSQL 16 (ou use o `docker compose` da raiz)
- [`migrate`](https://github.com/golang-migrate/migrate) CLI para os alvos de migration do Makefile

## Início rápido

Os comandos `make` são executados **na raiz do projeto** (o `Makefile` orquestra
backend, frontend e infra).

```bash
cp backend/.env.example backend/.env   # ajuste as variáveis se necessário
make up                                # sobe Postgres + migrations + API + frontend via Docker
```

Ou rodando a API localmente contra um Postgres já disponível:

```bash
cp backend/.env.example backend/.env
make migrate-up           # aplica o schema
make be-run               # API em http://localhost:8080
```

Healthcheck: `GET /health` → `{"status":"ok"}`.
Login inicial (seed): `admin@loja.local` / `admin123` — **troque em produção**.

## Comandos (Makefile da raiz)

| Comando | Descrição |
|---------|-----------|
| `make be-run` | `go run ./cmd/api` (API em `:8080`) |
| `make be-build` | compila o binário em `backend/bin/api` |
| `make be-test` | roda a suíte de testes do backend |
| `make be-vet` / `be-fmt` | `go vet` / `gofmt -w` |
| `make up` / `down` / `logs` | docker compose (db + migrate + api + frontend) |
| `make migrate-up` | aplica todas as migrations |
| `make migrate-down` | reverte a última migration |
| `make migrate-create name=add_xyz` | cria um par `.up.sql`/`.down.sql` |
| `make reset` | DROP total + migrate-up + seed |

Para rodar um pacote ou um único teste, chame `go test` direto em `backend/`:

```bash
cd backend
go test ./internal/modules/clientes/domain                # um pacote
go test -run TestValidarCPF ./internal/modules/clientes/domain   # um teste
```

## Arquitetura

### Camadas de um módulo

Cada bounded context vive em `internal/modules/<dominio>/`. As dependências
apontam **sempre para dentro** — o `domain` não conhece infraestrutura.

```
modules/clientes/
├── domain/        # entidade raiz + invariantes + erros sentinela (sem infra)
├── ports/
│   ├── inbound.go     # o que o módulo OFERECE (caso de uso) + Inputs
│   └── outbound.go    # o que o módulo EXIGE (repositório, gateways)
├── application/   # Service: implementa a porta inbound, orquestra o domínio
├── adapters/
│   ├── inbound/http/      # handler (JSON↔serviço) + router (rotas + RBAC)
│   └── outbound/
│       ├── postgres/      # implementa o repositório (porta outbound)
│       └── cep/           # gateway ViaCEP (porta outbound) com resiliência
└── module.go      # composition root (DI): monta tudo e expõe Router()
```

Fluxo de uma requisição:

```
HTTP → handler → ports.Service (application) → domain
                                  ↘ ports.Repository / ports.Gateway → Postgres / ViaCEP
```

### Plataforma compartilhada (`internal/platform/`)

| Pacote | Responsabilidade |
|--------|------------------|
| `config` | carrega `.env` + env vars com defaults de desenvolvimento |
| `database` | pool de conexões pgx |
| `httpserver` | router chi, middlewares globais, `/health`, helpers `JSON`/`Error`/`DecodeJSON` |
| `auth` | emissão/validação de JWT, middleware `Authenticate` + `RequirePerm` (RBAC) |
| `resilience` | `Retry → CircuitBreaker → Bulkhead`, compostos em `Policy` |

### Autorização

JWT HS256. As permissões viajam **dentro do token** (claim `perms`) — não há
consulta ao banco por request. Toda rota protegida é embrulhada por
`authMgr.Authenticate` + `auth.RequirePerm("recurso:acao")`. Convenção:
`<recurso>:read` e `<recurso>:write`.

### Resiliência

Adaptadores que chamam APIs externas embrulham a chamada numa `resilience.Policy`
(`Retry → CircuitBreaker → Bulkhead`). Erros **não-retriáveis** são marcados com
`resilience.Permanent(err)`; transitórios (timeout, 5xx) são deixados crus para o
Retry reagir. Exemplo: o gateway de CEP em `adapters/outbound/cep`.

### Banco de dados

**1 schema Postgres por bounded context** e **sem foreign keys entre schemas** —
a integridade referencial entre contextos é da aplicação (futuro: eventos/sagas).
Isso permite extrair cada schema para um banco/serviço próprio sem reescrever DDL.
Migrations em `migrations/`, pares `NNNNNN_nome.up.sql`/`.down.sql` aplicados em
sequência. Nunca edite uma migration já aplicada — crie uma nova.

## API

Tudo sob `/api/v1`. Erros seguem o envelope `{"error":{"code","message"}}` e
listas retornam `{"items":[...]}`.

### Clientes (`/api/v1/clientes`)

| Método | Rota | Permissão | Descrição |
|--------|------|-----------|-----------|
| `GET` | `/?q=&limit=&offset=` | `clientes:read` | lista/pesquisa |
| `GET` | `/{id}` | `clientes:read` | busca por ID |
| `GET` | `/by-cpf/{cpf}` | `clientes:read` | busca por CPF |
| `GET` | `/cep/{cep}` | `clientes:read` | consulta CEP (ViaCEP) |
| `POST` | `/` | `clientes:write` | cria cliente |
| `PUT` | `/{id}` | `clientes:write` | atualiza cliente |
| `DELETE` | `/{id}` | `clientes:write` | remove cliente |

Invariantes do domínio `clientes`: CPF com 11 dígitos válidos (e único), Nome e
E-mail obrigatórios; o endereço pode ser completado via consulta de CEP.

Referência completa em [`../docs/reference/api.md`](../docs/reference/api.md).

## Variáveis de ambiente

Veja `.env.example`. Principais:

| Variável | Default (dev) | Descrição |
|----------|---------------|-----------|
| `APP_ENV` | `development` | ambiente |
| `APP_PORT` | `8080` | porta HTTP |
| `DATABASE_URL` | `postgres://erp:erp_secret@localhost:5432/erp_estoque?sslmode=disable` | conexão Postgres |
| `JWT_SECRET` | `troque-este-segredo-em-producao` | segredo HS256 |
| `JWT_ACCESS_TTL` | `15m` | validade do access token |
| `CEP_API_URL` | `https://viacep.com.br/ws` | base da API de CEP |

## Build com Docker

`Dockerfile` multi-stage: imagem de build `golang:1.25-alpine` compila um binário
estático (`CGO_ENABLED=0`) e o serve numa imagem `alpine` enxuta, copiando junto a
pasta `migrations/`.

Mais detalhes em [`../docs/`](../docs/README.md) e nas convenções para o Claude
em [`CLAUDE.md`](CLAUDE.md).
