package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// NewVendasRouter monta as rotas do módulo vendas com RBAC.
func NewVendasRouter(h *Handler, authMgr *auth.Manager) chi.Router {
	r := chi.NewRouter()
	r.Use(authMgr.Authenticate)

	r.With(auth.RequirePerm("vendas:read")).Get("/", h.ListarVendas)
	r.With(auth.RequirePerm("vendas:write")).Post("/", h.CriarVenda)
	r.With(auth.RequirePerm("vendas:read")).Get("/{id}", h.BuscarVenda)
	r.With(auth.RequirePerm("vendas:write")).Post("/{id}/confirmar", h.ConfirmarVenda)

	return r
}
