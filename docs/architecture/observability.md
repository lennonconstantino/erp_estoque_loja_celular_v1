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
  docker compose -f docker-compose.observability.yml up -d
  # Grafana http://localhost:3000 (admin/admin) · Prometheus http://localhost:9090
  ```

  O Grafana já vem com o datasource Prometheus provisionado
  (`observability/grafana/provisioning/`).

### Por que começar mínimo

Tracing distribuído rende pouco quando **não há serviços distribuídos**: num
único processo, o trace é praticamente o stack trace. Subir Collector + Tempo +
Loki agora seria mais YAML de observabilidade do que valor entregue. O
investimento que importa hoje — e que torna a transição barata — é a
**instrumentação OTel**, que já está feita.

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
