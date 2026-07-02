# ERP — Estoque de Loja de Acessórios para Celular

Backend em **Go** com arquitetura hexagonal, PostgreSQL e autenticação JWT + RBAC.

> **Status:** em produção desde 2026-07-01 (Railway + Supabase). Retrospectiva da
> subida e lições de deploy em [Lições Aprendidas](licoes-aprendidas.md).

## Brief do Cliente

| Documento | Conteúdo |
|-----------|----------|
| [Client Brief](client-brief.md) | Contexto de negócio, problema, módulos e definição de pronto |

## Processo

| Documento | Conteúdo |
|-----------|----------|
| [Playbook de Planejamento](playbook-planejamento.md) | Método reproduzível para planejar e executar um projeto em fases (destilado deste): sistema de documentos, Leis vs Regras, micro-ciclo D1–D3, macro-ciclo de maturidade e armadilhas conhecidas |
| [Mandates](mandates.md) | Definition of Done por tarefa (D1–D3), protocolo do agente juiz, templates e política de cobertura |
| [Template de Spec (PROMPT)](templates/PROMPT.template.md) | Esqueleto preenchível de especificação executável para um projeto novo |

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
| [Design System](reference/design-system.md) | Kit de UI, tokens/tema, paleta + contraste e baseline de acessibilidade |

## Setup

| Documento | Conteúdo |
|-----------|----------|
| [Setup do Backend](setup/backend-setup.md) | Início rápido, Docker, variáveis de ambiente |
| [Supabase](setup/supabase-setup.md) | Provisionamento do PostgreSQL gerenciado (host de banco) |
| [Banco de Dados e Migrations](setup/database-migrations.md) | Inicialização, migrations e seed |
| [Setup do Frontend](setup/frontend-setup.md) | Convenções da SPA React, stack e variáveis |
| [Deploy no Railway](setup/railway-deployment.md) | Deploy do backend e frontend em produção |
| [CI/CD](setup/cicd.md) | Esteira GitHub Actions (gate) + auto-deploy Railway, staging→produção, fitness functions |

## Runbooks

| Documento | Conteúdo |
|-----------|----------|
| [CircuitBreakerOpen](runbooks/circuit-breaker.md) | Diagnóstico e resposta quando um circuit breaker abre |

## Retrospectivas

| Documento | Conteúdo |
|-----------|----------|
| [Lições Aprendidas — Deploy Fase 9](licoes-aprendidas.md) | Topologia de produção, sequência do deploy, mapa mental e runbook dos dois bugs de provedor (IPv6/pooler e nginx `$PORT`) — com diagramas Mermaid |
