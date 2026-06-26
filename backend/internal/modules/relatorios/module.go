// Package relatorios monta (DI) o bounded context de relatórios.
// É um módulo somente-leitura: não expõe escrita nem ports de saída além do banco.
package relatorios

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/adapters/inbound/http"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/adapters/outbound/postgres"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// Module expõe o roteador do contexto relatorios.
type Module struct {
	router chi.Router
}

// New monta o módulo de relatórios.
func New(pool *pgxpool.Pool, authMgr *auth.Manager) *Module {
	repo    := postgres.NewRelatorioRepository(pool)
	svc     := application.NewService(repo)
	handler := httpadapter.NewHandler(svc)

	return &Module{
		router: httpadapter.NewRelatoriosRouter(handler, authMgr),
	}
}

// Router retorna o roteador para montagem em /api/v1/relatorios.
func (m *Module) Router() chi.Router { return m.router }
