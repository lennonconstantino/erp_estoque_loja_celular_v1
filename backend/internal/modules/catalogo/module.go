// Package catalogo monta (DI) o bounded context de catálogo: instancia
// repositórios, serviço e roteadores HTTP para categorias e produtos.
// Expõe CatalogoReader e CatalogoWriter para consumo por outros módulos.
package catalogo

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/adapters/inbound/http"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/adapters/outbound/postgres"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// Module expõe os roteadores e as portas cross-module do contexto catálogo.
type Module struct {
	svc              *application.Service
	categoriasRouter chi.Router
	produtosRouter   chi.Router
}

// New monta o módulo: instancia repositórios, serviço e roteadores.
func New(pool *pgxpool.Pool, authMgr *auth.Manager) *Module {
	catsRepo  := postgres.NewCategoriaRepository(pool)
	prodsRepo := postgres.NewProdutoRepository(pool)

	svc     := application.NewService(catsRepo, prodsRepo)
	handler := httpadapter.NewHandler(svc, svc)

	return &Module{
		svc:              svc,
		categoriasRouter: httpadapter.NewCategoriasRouter(handler, authMgr),
		produtosRouter:   httpadapter.NewProdutosRouter(handler, authMgr),
	}
}

// CategoriasRouter retorna o roteador de categorias para montagem em /api/v1/categorias.
func (m *Module) CategoriasRouter() chi.Router { return m.categoriasRouter }

// ProdutosRouter retorna o roteador de produtos para montagem em /api/v1/produtos.
func (m *Module) ProdutosRouter() chi.Router { return m.produtosRouter }

// Reader retorna a interface CatalogoReader para uso por compras/vendas.
func (m *Module) Reader() ports.CatalogoReader { return m.svc }

// Writer retorna a interface CatalogoWriter para uso pelo módulo estoque.
func (m *Module) Writer() ports.CatalogoWriter { return m.svc }
