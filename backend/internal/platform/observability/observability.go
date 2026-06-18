// Package observability configura a stack OpenTelemetry do serviço.
//
// Filosofia: instrumentar UMA única vez com OpenTelemetry e deixar o backend de
// cada sinal ser uma decisão de configuração, não de código.
//
//   - Métricas: sempre ligadas. São expostas no formato Prometheus em /metrics
//     (Prometheus raspa; Grafana visualiza).
//   - Tracing: instrumentação sempre presente, mas DORMENTE. Só passa a exportar
//     quando OTEL_EXPORTER_OTLP_ENDPOINT aponta para um OTel Collector. Sem
//     endpoint, o TracerProvider global é o no-op padrão do OTel (custo
//     desprezível) — o tracing distribuído entra na migração para
//     microsserviços (ver docs/architecture/observability.md).
package observability

import (
	"context"
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Config parametriza a inicialização da observabilidade.
type Config struct {
	ServiceName  string // identifica o serviço nas métricas/traces (ex.: "erp-api")
	ServiceEnv   string // ambiente de execução (ex.: "development", "production")
	OTLPEndpoint string // endpoint do OTel Collector; vazio = tracing dormente
}

// Provider agrega os recursos OTel criados, expõe o handler de métricas e
// concentra o desligamento gracioso (flush de spans pendentes etc.).
type Provider struct {
	// MetricsHandler serve as métricas no formato Prometheus; monte-o em /metrics.
	MetricsHandler http.Handler
	// TracingAtivo indica se o exportador OTLP de traces foi ligado.
	TracingAtivo bool

	shutdownFns []func(context.Context) error
}

// Shutdown encerra todos os providers registrados, agregando os erros.
func (p *Provider) Shutdown(ctx context.Context) error {
	var err error
	for _, fn := range p.shutdownFns {
		err = errors.Join(err, fn(ctx))
	}
	return err
}

// Setup inicializa o MeterProvider (exportado via Prometheus) e, quando há
// endpoint OTLP, o TracerProvider (exportado via OTLP/HTTP). Ambos são
// registrados como globais — os middlewares e adaptadores usam
// otel.Meter(...)/otel.Tracer(...) sem conhecer a configuração concreta.
func Setup(ctx context.Context, cfg Config) (*Provider, error) {
	res := resource.NewSchemaless(
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("deployment.environment", cfg.ServiceEnv),
	)

	p := &Provider{}

	// --- Métricas: OTel SDK -> exporter Prometheus -> /metrics ---------------
	reg := prometheus.NewRegistry()
	// Métricas de runtime do Go (GC, goroutines, memória) + do processo.
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	metricExp, err := promexporter.New(promexporter.WithRegisterer(reg))
	if err != nil {
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(metricExp),
	)
	otel.SetMeterProvider(mp)
	p.shutdownFns = append(p.shutdownFns, mp.Shutdown)
	p.MetricsHandler = promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	// Propagação de contexto de trace (W3C traceparent + baggage). Inofensivo
	// mesmo com tracing dormente; deixa o serviço pronto para correlação
	// distribuída assim que o Collector existir.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// --- Tracing: dormente até existir um OTel Collector ---------------------
	if cfg.OTLPEndpoint == "" {
		// Sem endpoint: mantém o no-op global do OTel. Instrumentação presente,
		// exportação desligada.
		return p, nil
	}

	// O exportador lê a configuração de OTEL_EXPORTER_OTLP_* do ambiente
	// (endpoint, headers, TLS), seguindo a convenção padrão do OpenTelemetry.
	traceExp, err := otlptracehttp.New(ctx)
	if err != nil {
		// Falha ao montar o exportador não pode derrubar o serviço: as métricas
		// já estão de pé. Devolve o provider com tracing desligado e o erro.
		return p, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp),
	)
	otel.SetTracerProvider(tp)
	p.shutdownFns = append(p.shutdownFns, tp.Shutdown)
	p.TracingAtivo = true

	return p, nil
}
