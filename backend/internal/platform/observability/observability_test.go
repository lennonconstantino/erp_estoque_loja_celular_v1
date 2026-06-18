package observability_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/observability"
)

// Sem endpoint OTLP, o tracing deve ficar dormente, mas as métricas devem
// funcionar: o handler /metrics responde 200 com texto no formato Prometheus.
func TestSetup_MetricasLigadasTracingDormente(t *testing.T) {
	obs, err := observability.Setup(context.Background(), observability.Config{
		ServiceName:  "erp-test",
		ServiceEnv:   "test",
		OTLPEndpoint: "", // sem Collector
	})
	if err != nil {
		t.Fatalf("Setup retornou erro: %v", err)
	}
	t.Cleanup(func() { _ = obs.Shutdown(context.Background()) })

	if obs.TracingAtivo {
		t.Error("TracingAtivo deveria ser false sem OTLPEndpoint")
	}
	if obs.MetricsHandler == nil {
		t.Fatal("MetricsHandler não deveria ser nil")
	}

	rec := httptest.NewRecorder()
	obs.MetricsHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; esperado 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Result().Body)
	// go_goroutines é registrada pelo collector de runtime do Go — prova que o
	// pipeline de métricas está de pé e exportando no formato Prometheus.
	if !strings.Contains(string(body), "go_goroutines") {
		t.Errorf("corpo de /metrics não contém métricas de runtime do Go:\n%s", body)
	}
}

// Endpoint OTLP malformado não pode derrubar o serviço: as métricas continuam
// de pé e o erro é devolvido para o chamador decidir.
func TestSetup_EndpointInvalidoNaoDerrubaMetricas(t *testing.T) {
	obs, err := observability.Setup(context.Background(), observability.Config{
		ServiceName:  "erp-test",
		OTLPEndpoint: "://endereço inválido",
	})
	if obs == nil {
		t.Fatal("Provider não deveria ser nil mesmo com endpoint inválido")
	}
	t.Cleanup(func() { _ = obs.Shutdown(context.Background()) })

	// O exportador pode falhar (err != nil) ou aceitar a config preguiçosamente;
	// em qualquer caso, as métricas precisam continuar funcionando.
	_ = err
	if obs.MetricsHandler == nil {
		t.Error("MetricsHandler não deveria ser nil após falha de tracing")
	}
}
