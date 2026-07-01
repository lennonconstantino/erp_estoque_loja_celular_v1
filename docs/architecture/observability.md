# 10 — Observabilidade

> **Decisão (2026-06-18):** instrumentar com **OpenTelemetry** desde já, mas
> rodar apenas o backend mínimo de **métricas** (Prometheus + Grafana) enquanto
> formos um monólito. O resto da stack (traces + logs centralizados + alertas)
> entra na **migração para microsserviços** — ver
> [microservices-roadmap.md](microservices-roadmap.md).

## Princípio: instrumentar uma vez, escolher o backend por configuração

A regra de ouro é a mesma do resto da arquitetura hexagonal: **o código de
negócio não conhece o backend de telemetria**. Instrumentamos com a API do
OpenTelemetry (vendor-neutral) e cada sinal vira uma decisão de _configuração_,
não de _código_:

| Sinal | Como é coletado | Backend hoje | Ligado? |
|-------|-----------------|--------------|---------|
| **Métricas** | OTel SDK → exporter Prometheus em `/metrics` | Prometheus + Grafana | ✅ sempre |
| **Traces** | OTel SDK (`otelhttp`) → OTLP | OTel Collector → Tempo | 💤 dormente |
| **Logs** | `slog` estruturado em stdout | (docker logs / Loki depois) | parcial |

"Dormente" significa: a instrumentação de tracing **já está no código** (middleware
`otelhttp`, propagação W3C `traceparent`), mas o `TracerProvider` global é o no-op
padrão do OTel até `OTEL_EXPORTER_OTLP_ENDPOINT` apontar para um Collector. Custo
desprezível agora; ligar traces depois é setar uma env var — zero mudança de código.

## O que está montado agora (Fase 1 — monólito)

- **Backend** — `internal/platform/observability` inicializa o `MeterProvider`
  (exporter Prometheus) e registra métricas de runtime do Go (GC, goroutines,
  memória). O middleware `otelhttp` em `platform/httpserver` mede latência,
  volume e status de toda requisição HTTP. Métricas expostas em `GET /metrics`.
- **Infra** — `docker-compose.observability.yml` (arquivo **separado** do compose
  principal) sobe Prometheus + Grafana sob demanda:

  ```bash
  # Subir
  docker compose -f docker-compose.observability.yml up -d
  # Grafana http://localhost:3000 (admin/admin) · Prometheus http://localhost:9090

  # Derrubar (stack de obs isolada)
  docker compose -f docker-compose.observability.yml down

  # Teardown completo (app + obs + volumes de dados do Prometheus/Grafana)
  ./scripts/docker/teardown.sh --obs
  ```

  O Grafana já vem com o datasource Prometheus provisionado
  (`observability/grafana/provisioning/`).

### Por que começar mínimo

Tracing distribuído rende pouco quando **não há serviços distribuídos**: num
único processo, o trace é praticamente o stack trace. Subir Collector + Tempo +
Loki agora seria mais YAML de observabilidade do que valor entregue. O
investimento que importa hoje — e que torna a transição barata — é a
**instrumentação OTel**, que já está feita.

## Produção hoje (Railway) — proteção do `/metrics` + push gerenciado

Em produção **não** subimos Prometheus/Grafana _stateful_ no Railway: para um
monólito de baixo tráfego, o custo/operação de rodar dois serviços com volume não
se paga (o modelo _pull_ ainda encaixa mal em containers efêmeros). A escolha é
**push para um backend gerenciado** (Grafana Cloud), com três peças:

> 🚀 **Ativação passo a passo** (tokens, serviço Alloy no Railway, carga de
> alertas, verificação): [docs/setup/observability-activation.md](../setup/observability-activation.md).

**1. `/metrics` protegido.** O endpoint fica no mesmo servidor HTTP público da API,
então `httpserver.ProtectMetrics` controla o acesso conforme o ambiente:

| `METRICS_TOKEN` | Ambiente | Comportamento |
|-----------------|----------|---------------|
| vazio | dev | **aberto** — Prometheus local raspa sem credencial |
| vazio | produção | **fechado (404)** — fail-safe, nunca expõe métricas sem proteção |
| definido | qualquer | exige `Authorization: Bearer <token>` (comparação em tempo constante) |

Ou seja, subir em produção sem configurar `METRICS_TOKEN` já é seguro por padrão
(métricas indisponíveis). Para habilitar o scraping, gere um token
(`openssl rand -hex 32`) e informe o **mesmo valor** ao agente que raspa.

**2. Coleta via Grafana Alloy → Grafana Cloud (remote_write).** Um serviço
[`grafana/alloy`](../../observability/alloy/config.alloy) no Railway raspa o
`/metrics` do backend pela rede privada (`*.railway.internal`), autenticando com o
`ERP_METRICS_TOKEN` (= `METRICS_TOKEN` do backend), e faz `remote_write` para o
Grafana Cloud. Zero infra _stateful_ nossa; retenção e dashboards são gerenciados.

> **Não contradiz** a decisão de "Collector _e_ Alloy juntos ❌" abaixo: aqui roda
> **um só** agente (Alloy), que é a distribuição recomendada pelo próprio Grafana
> Cloud para push. A escolha Collector-vs-Alloy da stack self-hosted da Fase 3+
> continua valendo — são fases distintas, nunca simultâneas.

**3. Alertas.** [`observability/alerts/erp-alerts.yml`](../../observability/alerts/erp-alerts.yml)
traz 4 alertas em formato Prometheus (carregados localmente via `rule_files`; no
Grafana Cloud, via `mimirtool rules load`):

| Alerta | Dispara quando | Severidade |
|--------|----------------|------------|
| `ERPBackendDown` | `up == 0` por 2min (processo caiu / reiniciando) | critical |
| `ERPBackendUnreachable` | `absent(up)` por 5min (Alloy não alcança o backend) | critical |
| `ERPHighErrorRate` | > 5% de 5xx por 5min (inclui falha de banco/pooler e fiscal) | critical |
| `ERPHighLatencyP99` | p99 de latência HTTP > 1s por 10min | warning |

> **Lacuna consciente:** não há alerta específico de _circuit breaker aberto_ nem
> de _erro de conexão com o banco_ porque `resilience/` e `database/` ainda **não
> emitem métricas próprias** — só o middleware HTTP e o runtime Go. A falha dessas
> dependências aparece como 5xx, coberta por `ERPHighErrorRate`. Alertar na causa
> raiz exige instrumentar esses pacotes com OTel (trabalho futuro).

## Stack-alvo na migração para microsserviços (Fase 3+)

Quando os bounded contexts começarem a virar processos separados (ver as fases
em [microservices-roadmap.md](microservices-roadmap.md)), liga-se a stack
completa, toda centrada no OpenTelemetry e na suíte Grafana:

```
            ┌─────────────────────────────────────────────┐
serviços ──▶│  OTel Collector  (1 ponto de coleta/roteio)  │
(OTel SDK)  └───────┬──────────────┬──────────────┬────────┘
                    │ métricas      │ traces        │ logs
                    ▼               ▼               ▼
              Prometheus          Tempo            Loki
                    └──────────────┴───────────────┘
                                   ▼
                                Grafana   ◀── trace ⇄ log ⇄ métrica
                                   │
                              Alertmanager (alertas via Prometheus)
```

Componentes-alvo e o porquê de cada escolha:

| Componente | Papel | Por que este |
|------------|-------|--------------|
| **OTel Collector** | recebe/processa/roteia métricas, traces e logs | desacopla apps dos backends; troca de backend não toca em serviço |
| **Prometheus** | métricas + regras de alerta | padrão de mercado, pull-based |
| **Tempo** | backend de traces | integra nativo com Grafana e correlaciona trace↔log↔métrica |
| **Loki** | logs centralizados | mesma suíte/linguagem de query do Grafana |
| **Grafana** | visualização única dos 3 sinais | painel único correlacionado |
| **Alertmanager** | roteamento de alertas | par natural do Prometheus |

### Decisões deliberadas (o que NÃO usar)

- **Jaeger / Zipkin** ❌ — backends de tracing em silo, sem a correlação nativa
  trace↔log↔métrica que Tempo dá dentro do Grafana. Como já emitimos OTLP, Tempo
  é plugar-e-usar.
- **Promtail** ❌ — em modo manutenção/depreciação pela Grafana. Coleta de logs
  fica no **OTel Collector** (ou Alloy), não no Promtail.
- **Collector _e_ Alloy juntos** ❌ — Alloy é uma distribuição do próprio OTel
  Collector; rodar os dois é redundância. Escolha **um** (preferência: OTel
  Collector, vendor-neutral).

> Em resumo: hoje **métricas**; o gancho de **traces** já está no código,
> dormente; **logs centralizados** e **alertas** entram junto da primeira
> extração de serviço. A instrumentação não muda — só a configuração.
