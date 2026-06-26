package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// NewCategoriasRouter monta as rotas de categorias com autenticação + RBAC.
func NewCategoriasRouter(h *Handler, authMgr *auth.Manager) chi.Router {
	r := chi.NewRouter()
	r.Use(authMgr.Authenticate)

	r.With(auth.RequirePerm("catalogo:read")).Get("/", h.ListarCategorias)
	r.With(auth.RequirePerm("catalogo:read")).Get("/{id}", h.BuscarCategoriaPorID)
	r.With(auth.RequirePerm("catalogo:write")).Post("/", h.CriarCategoria)
	r.With(auth.RequirePerm("catalogo:write")).Put("/{id}", h.AtualizarCategoria)
	r.With(auth.RequirePerm("catalogo:write")).Delete("/{id}", h.RemoverCategoria)

	return r
}

// NewProdutosRouter monta as rotas de produtos com autenticação + RBAC.
func NewProdutosRouter(h *Handler, authMgr *auth.Manager) chi.Router {
	r := chi.NewRouter()
	r.Use(authMgr.Authenticate)

	r.With(auth.RequirePerm("catalogo:read")).Get("/", h.ListarProdutos)
	r.With(auth.RequirePerm("catalogo:read")).Get("/{id}", h.BuscarProdutoPorID)
	r.With(auth.RequirePerm("catalogo:write")).Post("/", h.CriarProduto)
	r.With(auth.RequirePerm("catalogo:write")).Put("/{id}", h.AtualizarProduto)

	return r
}
