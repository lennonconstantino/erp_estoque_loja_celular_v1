# ERP — Estoque de Loja de Acessórios para Celular

Documentação do **backend** do sistema de controle de estoque para loja de
acessórios de celular. O sistema cobre cadastros, compras, vendas, ajuste de
estoque e relatórios, com autenticação **JWT + RBAC**.

- **Linguagem:** Go
- **Arquitetura:** Hexagonal (Ports & Adapters), com domínios totalmente
  isolados, preparada para evolução para **microsserviços**.
- **Banco:** PostgreSQL (1 schema por bounded context).

## Índice

| # | Documento | Conteúdo |
|---|-----------|----------|
| 01 | [Visão Geral](01-visao-geral.md) | Escopo, módulos e regras de negócio |
| 02 | [Arquitetura](02-arquitetura.md) | Hexagonal, isolamento de domínios, fluxo de dependências |
| 03 | [Estrutura de Pastas](03-estrutura-de-pastas.md) | Layout do repositório e de cada módulo |
| 04 | [Domínios (Bounded Contexts)](04-dominios.md) | Responsabilidade e regras de cada domínio |
| 05 | [Modelo de Dados](05-modelo-de-dados.md) | Tabelas, colunas e relacionamentos |
| 06 | [API REST](06-api-rest.md) | Endpoints por módulo |
| 07 | [Segurança](07-seguranca.md) | Autenticação JWT e autorização RBAC |
| 08 | [Banco de Dados e Migrations](08-banco-e-migrations.md) | Inicialização, migrations e seed |
| 09 | [Roadmap para Microsserviços](09-roadmap-microservices.md) | Caminho de extração dos módulos |

## Módulos (a partir dos diagramas)

```
Login/Auth ── Usuários
    │
   Menu ── Clientes
        ├─ Fornecedores
        ├─ Categorias ──┐
        ├─ Produtos ◄───┘
        ├─ Compras   (entrada de estoque)
        ├─ Vendas    (saída de estoque)
        ├─ Ajuste de Estoque
        └─ Relatórios
```

## Início rápido

```bash
cp backend/.env.example backend/.env
docker compose up -d     # sobe Postgres + migrations + API
# ou, localmente:
cd backend
make migrate-up          # cria schemas e tabelas
make run                 # inicia a API em http://localhost:8080
```

Usuário inicial (seed): `admin@loja.local` / `admin123` — **troque em produção**.
