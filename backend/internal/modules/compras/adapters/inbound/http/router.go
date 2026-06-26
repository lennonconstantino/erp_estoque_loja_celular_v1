package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// NewComprasRouter monta as rotas do módulo compras com RBAC.
func NewComprasRouter(h *Handler, authMgr *auth.Manager) chi.Router {
	r := chi.NewRouter()
	r.Use(authMgr.Authenticate)

	r.With(auth.RequirePerm("compras:read")).Get("/", h.ListarCompras)
	r.With(auth.RequirePerm("compras:write")).Post("/", h.CriarCompra)
	r.With(auth.RequirePerm("compras:read")).Get("/{id}", h.BuscarCompra)
	r.With(auth.RequirePerm("compras:write")).Post("/{id}/confirmar", h.ConfirmarCompra)

	return r
}
