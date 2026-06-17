# ERP — Estoque de Loja de Acessórios para Celular

Backend em **Go** com arquitetura hexagonal, PostgreSQL e autenticação JWT + RBAC.

## Brief do Cliente

| Documento | Conteúdo |
|-----------|----------|
| [Client Brief](client-brief.md) | Contexto de negócio, problema, módulos e definição de pronto |

## Arquitetura

| Documento | Conteúdo |
|-----------|----------|
| [Arquitetura (visão consolidada)](architecture.md) | Topologia, fluxo de requisição, deploy local/Railway e não-objetivos |
| [Visão Geral](architecture/overview.md) | Escopo, módulos e regras de negócio |
| [Arquitetura Hexagonal](architecture/hexagonal.md) | Ports & Adapters, fluxo de dependências, stack |
| [Estrutura de Pastas](architecture/folder-structure.md) | Layout do repositório e anatomia de um módulo |
| [Domínios (Bounded Contexts)](architecture/domains.md) | Responsabilidade e regras de cada domínio |
| [Roadmap para Microsserviços](architecture/microservices-roadmap.md) | Fases de evolução e estratégia de extração |
| [Resilience Stack](architecture/resilience.md) | Circuit Breaker, Bulkhead e Retry nos adaptadores de saída |

## Referência

| Documento | Conteúdo |
|-----------|----------|
| [Modelo de Dados](reference/data-model.md) | Tabelas, colunas e relacionamentos |
| [API REST](reference/api.md) | Endpoints por módulo |
| [Segurança](reference/security.md) | Autenticação JWT e autorização RBAC |

## Guias

| Documento | Conteúdo |
|-----------|----------|
| [Setup do Backend](guides/backend-setup.md) | Início rápido, Docker, variáveis de ambiente |
| [Banco de Dados e Migrations](guides/database-migrations.md) | Inicialização, migrations e seed |
| [Setup do Frontend](guides/frontend-setup.md) | Configuração do frontend |
