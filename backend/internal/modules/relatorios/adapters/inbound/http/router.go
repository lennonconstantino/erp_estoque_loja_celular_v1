package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// NewRelatoriosRouter monta as rotas do módulo relatorios com RBAC.
func NewRelatoriosRouter(h *Handler, authMgr *auth.Manager) chi.Router {
	r := chi.NewRouter()
	r.Use(authMgr.Authenticate)

	r.With(auth.RequirePerm("relatorios:read")).Get("/produtos/abaixo-do-minimo", h.ProdutosAbaixoDoMinimo)
	r.With(auth.RequirePerm("relatorios:read")).Get("/produtos/mais-vendidos", h.MaisVendidos)
	r.With(auth.RequirePerm("relatorios:read")).Get("/produtos/menos-vendidos", h.MenosVendidos)
	r.With(auth.RequirePerm("relatorios:read")).Get("/vendas", h.ResumoVendas)
	r.With(auth.RequirePerm("relatorios:read")).Get("/compras", h.ResumoCompras)

	return r
}
