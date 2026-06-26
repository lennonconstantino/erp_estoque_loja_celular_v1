// Package compras monta (DI) o bounded context de compras: instancia repositórios,
// serviço e roteador HTTP. Recebe CatalogoReader, EstoqueWriter e FornecedorWriter
// dos módulos catálogo, estoque e fornecedores via injeção.
package compras

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/adapters/inbound/http"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/adapters/outbound/postgres"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// Module expõe o roteador do contexto compras.
type Module struct {
	router chi.Router
}

// New monta o módulo.
// catReader vem de catalogo.Module.Reader(), estoqueWriter de estoque.Module.Writer(),
// fornWriter de fornecedores.Module.Writer().
func New(
	pool *pgxpool.Pool,
	authMgr *auth.Manager,
	catReader ports.CatalogoReader,
	estoqueWriter ports.EstoqueWriter,
	fornWriter ports.FornecedorWriter,
) *Module {
	repo    := postgres.NewCompraRepository(pool)
	svc     := application.NewService(repo, catReader, estoqueWriter, fornWriter)
	handler := httpadapter.NewHandler(svc)

	return &Module{
		router: httpadapter.NewComprasRouter(handler, authMgr),
	}
}

// Router retorna o roteador para montagem em /api/v1/compras.
func (m *Module) Router() chi.Router { return m.router }
