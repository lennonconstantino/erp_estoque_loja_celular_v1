# ERP — Estoque de Loja de Acessórios para Celular

Backend em **Go** com arquitetura hexagonal, PostgreSQL e autenticação JWT + RBAC.

## Brief do Cliente

| Documento | Conteúdo |
|-----------|----------|
| [Client Brief](client-brief.md) | Contexto de negócio, problema, módulos e definição de pronto |

## Processo

| Documento | Conteúdo |
|-----------|----------|
| [Mandates](mandates.md) | Definition of Done por tarefa (D1–D3), protocolo do agente juiz, templates e política de cobertura |

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
| [Observabilidade](architecture/observability.md) | OpenTelemetry, métricas (Prometheus/Grafana) e stack-alvo de traces/logs |

## Referência

| Documento | Conteúdo |
|-----------|----------|
| [Modelo de Dados](reference/data-model.md) | Tabelas, colunas e relacionamentos |
| [API REST](reference/api.md) | Endpoints por módulo |
| [Segurança](reference/security.md) | Autenticação JWT e autorização RBAC |
| [Checklist de Segurança](reference/checklist.md) | Gate de revisão pré-deploy adaptado à stack |

## Setup

| Documento | Conteúdo |
|-----------|----------|
| [Setup do Backend](setup/backend-setup.md) | Início rápido, Docker, variáveis de ambiente |
| [Supabase](setup/supabase-setup.md) | Provisionamento do PostgreSQL gerenciado (host de banco) |
| [Banco de Dados e Migrations](setup/database-migrations.md) | Inicialização, migrations e seed |
| [Setup do Frontend](setup/frontend-setup.md) | Convenções da SPA React, stack e variáveis |
| [Deploy no Railway](setup/railway-deployment.md) | Deploy do backend e frontend em produção |

## Runbooks

| Documento | Conteúdo |
|-----------|----------|
| [CircuitBreakerOpen](runbooks/circuit-breaker.md) | Diagnóstico e resposta quando um circuit breaker abre |
