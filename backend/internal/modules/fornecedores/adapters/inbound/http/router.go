package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// NewRouter monta as rotas do módulo, protegidas por autenticação + RBAC.
// Permissões: fornecedores:read (consulta) e fornecedores:write (mutação).
func NewRouter(h *Handler, authMgr *auth.Manager) chi.Router {
	r := chi.NewRouter()
	r.Use(authMgr.Authenticate)

	// leitura
	r.With(auth.RequirePerm("fornecedores:read")).Get("/", h.Listar)
	r.With(auth.RequirePerm("fornecedores:read")).Get("/{id}", h.BuscarPorID)
	r.With(auth.RequirePerm("fornecedores:read")).Get("/by-cnpj/{cnpj}", h.BuscarPorCNPJ)
	r.With(auth.RequirePerm("fornecedores:read")).Get("/cep/{cep}", h.ConsultarCEP)

	// escrita
	r.With(auth.RequirePerm("fornecedores:write")).Post("/", h.Criar)
	r.With(auth.RequirePerm("fornecedores:write")).Put("/{id}", h.Atualizar)

	return r
}
