// Package estoque monta (DI) o bounded context de estoque: instancia
// repositórios, serviço e roteador HTTP. Recebe CatalogoWriter do módulo
// catálogo via injeção para atualizar o saldo materializado em produtos.
package estoque

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/adapters/inbound/http"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/adapters/outbound/postgres"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// Module expõe o roteador e a porta cross-module do contexto estoque.
type Module struct {
	svc    *application.Service
	router chi.Router
}

// New monta o módulo: instancia repositórios, serviço e roteador.
// catWriter é implementado pelo módulo catálogo (catalogo.Module.Writer()).
func New(pool *pgxpool.Pool, authMgr *auth.Manager, catWriter ports.CatalogoWriter) *Module {
	movsRepo    := postgres.NewMovimentacaoRepository(pool)
	ajustesRepo := postgres.NewAjusteRepository(pool)

	svc     := application.NewService(movsRepo, ajustesRepo, catWriter)
	handler := httpadapter.NewHandler(svc)

	return &Module{
		svc:    svc,
		router: httpadapter.NewEstoqueRouter(handler, authMgr),
	}
}

// Router retorna o roteador para montagem em /api/v1/estoque.
func (m *Module) Router() chi.Router { return m.router }

// Writer retorna a interface EstoqueWriter para uso por compras e vendas.
func (m *Module) Writer() ports.EstoqueWriter { return m.svc }
