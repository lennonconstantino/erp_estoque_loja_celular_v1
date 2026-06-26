// Package fornecedores monta (DI) o bounded context de fornecedores.
// É o único ponto que conhece as implementações concretas.
package fornecedores

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/adapters/inbound/http"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/adapters/outbound/cep"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/adapters/outbound/postgres"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/resilience"
)

// Module expõe o roteador e a porta cross-module do contexto fornecedores.
type Module struct {
	svc    *application.Service
	router chi.Router
}

// New monta o módulo. cepURL é a base da API de CEP (ex.: ViaCEP).
func New(pool *pgxpool.Pool, authMgr *auth.Manager, cepURL string) *Module {
	repo := postgres.NewFornecedorRepository(pool)

	cepPolicy := resilience.NewPolicy(
		resilience.RetryConfig{MaxAttempts: 5, InitialDelay: 100 * time.Millisecond, MaxDelay: 4 * time.Second, Multiplier: 2.0},
		resilience.CircuitBreakerConfig{FailureThreshold: 5, SuccessThreshold: 2, Timeout: 30 * time.Second},
		resilience.BulkheadConfig{MaxConcurrency: 10},
	)
	cepGateway := cep.NewGateway(cepURL, cepPolicy)

	svc     := application.NewService(repo, cepGateway)
	handler := httpadapter.NewHandler(svc)

	return &Module{svc: svc, router: httpadapter.NewRouter(handler, authMgr)}
}

// Router retorna o roteador do módulo para ser montado sob /api/v1/fornecedores.
func (m *Module) Router() chi.Router { return m.router }

// Writer retorna a interface FornecedorWriter para uso pelo módulo compras.
func (m *Module) Writer() *application.Service { return m.svc }
