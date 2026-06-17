# ERP — Estoque de Loja de Acessórios para Celular (v1)

**Backend** em **Go** com **arquitetura hexagonal**, domínios totalmente isolados
e caminho preparado para **microsserviços** (banco **PostgreSQL**, 1 schema por
bounded context). **Frontend** SPA em **React + Vite + TypeScript + Tailwind**.

> Estado atual: o módulo `clientes` está implementado no backend; os demais
> domínios (`iam`, `fornecedores`, `catalogo`, `compras`, `vendas`, `estoque`)
> existem como migrations + documentação e seguem o mesmo molde hexagonal.

## Estrutura do repositório

```
.
├── docker-compose.yml  # orquestração: db + migrate + api + frontend
├── backend/            # serviço Go (hexagonal)  — ver backend/CLAUDE.md
│   ├── cmd/            # entrypoints (api)
│   ├── internal/
│   │   ├── platform/   # infra compartilhada (db, auth, http, config, resilience)
│   │   └── modules/    # 1 pacote por domínio (iam, clientes, ... estoque)
│   ├── migrations/     # DDL versionada (golang-migrate)
│   ├── Dockerfile · Makefile · .env.example
│   └── go.mod
├── frontend/           # SPA React/Vite (pnpm)   — ver frontend/CLAUDE.md
│   ├── src/lib/        # http · api · auth · env
│   ├── src/pages/      # Login, Dashboard
│   └── Dockerfile (nginx) · vite.config.ts · package.json
└── docs/               # documentação (ver docs/README.md)
```

## Documentação

A documentação completa está em **[`docs/`](docs/README.md)**.

## Início rápido

### Stack completa (Docker)

```bash
cp backend/.env.example backend/.env
make up   # Postgres + migrations + API (:8080) + frontend (:80)
```

### Desenvolvimento local

Há um `Makefile` na raiz que orquestra backend, frontend e infra (`make help`
lista todos os alvos):

```bash
cp backend/.env.example backend/.env
make be-run         # API em :8080  (requer Postgres — use `make up` para subir só a infra)
make fe-install     # instala dependências do frontend (pnpm)
make fe-dev         # Vite dev server
```

Login inicial (seed): `admin@loja.local` / `admin123` — **troque em produção**.
