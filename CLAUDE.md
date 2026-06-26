# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

ERP de estoque para loja de acessórios de celular. Backend Go (monólito modular, arquitetura hexagonal), Postgres, frontend React/Vite. Código, comentários e documentação estão em **português** — mantenha esse padrão ao escrever código novo.

## Commands

Há **um único `Makefile` na raiz** que orquestra backend, frontend e infra (faz
`-include backend/.env`). Rode todos os alvos a partir da raiz; `make help` lista
todos. Requer `backend/.env` (`cp backend/.env.example backend/.env`).

```bash
# Backend (Go)
make be-run          # go run ./cmd/api          (API local em :8080)
make be-build        # go build -o bin/api ./cmd/api
make be-test         # go test ./...             (toda a suíte)
make be-vet          # go vet ./...
make migrate-up / migrate-down / migrate-create name=add_xyz / reset

# Frontend (React/Vite, pnpm)
make fe-install / fe-dev / fe-build / fe-lint

# Infra (docker compose: db + migrate + api + frontend)
make up / down / logs

# Agregados
make build           # be-build + fe-build
make lint            # be-vet  + fe-lint
```

Pacote ou teste único: `go test` direto em `backend/` —
`cd backend && go test -run TestNome ./internal/modules/clientes/domain`.

Seed de login inicial: `admin@loja.local` / `admin123`.

## Definition of Done por tarefa (obrigatório)

Ao executar qualquer tarefa de um plano (fase do `PROMPT.md` ou item de
`docs/todos.md`), siga o mandate canônico em **[docs/mandates.md](docs/mandates.md)**
(resumo executável em §0.5 do [PROMPT.md](PROMPT.md)). Resumo:

- **D1 — Testes no mesmo passo.** Entregue os testes unitários junto da lógica de
  domínio/caso de uso (não adie para a Fase 8). Verifique com
  `go test -cover ./internal/modules/<dominio>/...` — meta `domain/` ≥ 80%,
  `application/` ≥ 70%. Frontend não tem testes: `pnpm tsc --noEmit` + `pnpm lint`.
- **D2 — Checklist vivo.** Derive um checklist da tarefa antes de codar e marque
  `[x]` os itens da fase em `docs/todos.md` ao concluí-los.
- **D3 — Agente juiz.** Antes de declarar a fase como PASS, acione um subagent
  **juiz independente** para julgar o diff contra a spec da fase e as Leis L1–L8.
  Veredito `CONFORME` libera o avanço; `NÃO CONFORME` bloqueia até corrigir.

Ordem: implementar → testes (D1) → checklist (D2) → critério `curl`/PASS da fase
→ juiz CONFORME (D3).

## Architecture

**Monólito modular com arquitetura hexagonal**, desenhado para extração futura em microsserviços. A documentação completa vive em [docs/](docs/README.md) — os pontos que exigem ler vários arquivos:

### Anatomia de um módulo (bounded context)

Cada domínio é um pacote em `backend/internal/modules/<dominio>/` com camadas estritas (dependências apontam só para dentro):

- `domain/` — entidades, invariantes e erros. Sem dependências de infra. Validações de negócio vivem aqui (ex.: `NovoCliente` valida CPF).
- `ports/` — interfaces. `inbound.go` = o que o módulo **oferece** (caso de uso, ex.: `ClienteService`); `outbound.go` = o que o módulo **exige** (ex.: `ClienteRepository`, `CepGateway`).
- `application/` — casos de uso (`Service`) que implementam a porta inbound e orquestram domínio + portas outbound. Não conhece HTTP nem SQL. Usa `var _ ports.X = (*Service)(nil)` para garantir conformidade em tempo de compilação.
- `adapters/inbound/http/` — `handler.go` (HTTP↔serviço) + `router.go` (rotas + RBAC).
- `adapters/outbound/postgres/` — implementação do repositório; `adapters/outbound/cep/` etc. — gateways externos.
- `module.go` — **composition root / DI** do contexto: o único lugar que conhece as implementações concretas, monta tudo e expõe `Router()`. Para adicionar um módulo: implemente as camadas, exponha `New(...)` em `module.go`, e monte em [cmd/api/main.go](backend/cmd/api/main.go) sob `/api/v1` (há exemplos comentados lá).

> Estado atual: **todos os bounded contexts estão implementados em Go** (`iam`, `clientes`, `fornecedores`, `catalogo`, `estoque`, `compras`, `vendas`, `relatorios`), seguindo este molde. O frontend cobre todas as telas e compartilha o kit de UI em `@/components/ui`. Resta a Fase 9 (deploy) — ver [docs/todos.md](docs/todos.md).

### Platform (infra compartilhada — `backend/internal/platform/`)

- `auth/` — JWT HS256 + RBAC. `Manager.Authenticate` (middleware) injeta `Claims` no contexto; `RequirePerm("recurso:acao")` protege rotas. Permissões viajam no próprio token (claim `perms`); não há lookup por request. Toda rota protegida exige uma permissão no formato `recurso:acao` (ex.: `clientes:read`, `clientes:write`).
- `httpserver/` — router chi com middlewares globais, `/health`, e helpers `JSON`/`Error`/`DecodeJSON`. Erros seguem o envelope `{"error":{"code","message"}}`. `DecodeJSON` rejeita campos desconhecidos e limita o body a 1 MiB.
- `resilience/` — Retry + Circuit Breaker + Bulkhead, compostos em `Policy`. **Aplicados nos adaptadores outbound** que chamam APIs externas (ver a config da policy de CEP em [module.go](backend/internal/modules/clientes/module.go)).
- `config/`, `database/` (pool pgx).

### Banco de dados — isolamento por schema

**1 schema Postgres por bounded context** (`iam`, `clientes`, …). Regra crítica: **não há foreign keys entre schemas diferentes** — a integridade referencial entre contextos é responsabilidade da aplicação (e, no futuro, de eventos/sagas). Isso permite extrair cada schema para um banco/serviço próprio sem reescrever DDL. Comunicação entre módulos se dá por **portas** (ex.: `vendas` usa `CatalogoReader`/`EstoqueWriter`), nunca por JOIN cross-schema.

Migrations em `backend/migrations/`, numeradas sequencialmente com pares `.up.sql`/`.down.sql` (golang-migrate). `000001_init` cria extensões + schemas; cada domínio tem sua própria migration; `000009_seed` popula dados iniciais.

Notas de modelagem relevantes: `estoque.movimentacoes` é um ledger **append-only** (fonte da verdade do saldo); `catalogo` mantém saldo materializado atualizado via porta; datas como `dt_ult_comp_cli`/`dt_ult_comp_for` são atualizadas por eventos (`VendaConfirmada`/`CompraConfirmada`). Detalhes em [docs/architecture/domains.md](docs/architecture/domains.md).

### Frontend (`frontend/`)

React 18 + Vite + TypeScript + Tailwind + shadcn/ui (`components.json`), gerenciado com **pnpm**. Alias `@` → `src/`. Camada de rede própria: `lib/http.ts` (fetch tipado) → `lib/api.ts` (timeout + refresh de token automático, `ApiError`) → `lib/auth.ts` (tokens). Em produção é servido por nginx (ver `frontend/nginx.conf` e `Dockerfile`).

Todas as telas compõem o **kit de UI** em `@/components/ui` (casca `PageShell` com `Sidebar` fixa + paleta de comandos ⌘K, `DataTable`, `Tabs`, `Modal`, `Field`, `StatusBadge`, `Button`) e estilizam **só com tokens semânticos** do tema — há **Dark/Light mode** via `@/lib/theme` (classe `.dark` no `<html>`, tokens em `src/index.css`), então nada de cores cruas. Detalhes em [docs/setup/frontend-setup.md](docs/setup/frontend-setup.md) e [frontend/CLAUDE.md](frontend/CLAUDE.md).
