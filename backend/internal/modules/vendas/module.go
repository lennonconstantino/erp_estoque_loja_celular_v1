// Package vendas monta (DI) o bounded context de vendas: instancia repositórios,
// serviço, gateway fiscal e roteador HTTP.
package vendas

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/adapters/inbound/http"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/adapters/outbound/fiscal"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/adapters/outbound/postgres"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/resilience"
)

// Module expõe o roteador do contexto vendas.
type Module struct {
	router chi.Router
}

// New monta o módulo: instancia repositório, gateway fiscal (com resiliência) e serviço.
// catReader, estoqueWriter e clienteWriter são interfaces duck-typed dos módulos vizinhos.
func New(
	pool *pgxpool.Pool,
	authMgr *auth.Manager,
	catReader ports.CatalogoReader,
	estoqueWriter ports.EstoqueWriter,
	clienteWriter ports.ClienteWriter,
) *Module {
	repo := postgres.NewVendaRepository(pool)

	fiscalPolicy := resilience.NewPolicy(
		resilience.RetryConfig{MaxAttempts: 3, InitialDelay: 200 * time.Millisecond, MaxDelay: 2 * time.Second, Multiplier: 2.0},
		resilience.CircuitBreakerConfig{FailureThreshold: 5, SuccessThreshold: 2, Timeout: 30 * time.Second},
		resilience.BulkheadConfig{MaxConcurrency: 5},
	)
	fiscalGateway := fiscal.NewGateway(fiscalPolicy)

	svc := application.NewService(repo, catReader, estoqueWriter, clienteWriter, fiscalGateway)
	handler := httpadapter.NewHandler(svc)

	return &Module{router: httpadapter.NewVendasRouter(handler, authMgr)}
}

// Router retorna o roteador para montagem em /api/v1/vendas.
func (m *Module) Router() chi.Router { return m.router }
