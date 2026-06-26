package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// NewEstoqueRouter monta as rotas do contexto estoque sob /estoque.
// Rotas montadas:
//
//	POST /estoque/ajustes          — lança ajuste manual  (estoque:write)
//	GET  /estoque/{produtoId}      — razão de movimentações (estoque:read)
//	GET  /estoque/{produtoId}/ajustes — ajustes manuais    (estoque:read)
func NewEstoqueRouter(h *Handler, authMgr *auth.Manager) chi.Router {
	r := chi.NewRouter()
	r.Use(authMgr.Authenticate)

	r.With(auth.RequirePerm("estoque:write")).Post("/ajustes", h.LancarAjuste)
	r.With(auth.RequirePerm("estoque:read")).Get("/{produtoId}", h.ConsultarMovimentacoes)
	r.With(auth.RequirePerm("estoque:read")).Get("/{produtoId}/ajustes", h.ConsultarAjustes)

	return r
}
