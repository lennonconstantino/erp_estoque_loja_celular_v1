package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// NewRouter monta as rotas do módulo, protegidas por autenticação + RBAC.
// Permissões: clientes:read (consulta) e clientes:write (mutação).
func NewRouter(h *Handler, authMgr *auth.Manager) chi.Router {
	r := chi.NewRouter()
	r.Use(authMgr.Authenticate)

	// leitura
	r.With(auth.RequirePerm("clientes:read")).Get("/", h.Listar)
	r.With(auth.RequirePerm("clientes:read")).Get("/{id}", h.BuscarPorID)
	r.With(auth.RequirePerm("clientes:read")).Get("/by-cpf/{cpf}", h.BuscarPorCPF)
	r.With(auth.RequirePerm("clientes:read")).Get("/cep/{cep}", h.ConsultarCEP)

	// escrita
	r.With(auth.RequirePerm("clientes:write")).Post("/", h.Criar)
	r.With(auth.RequirePerm("clientes:write")).Put("/{id}", h.Atualizar)
	r.With(auth.RequirePerm("clientes:write")).Delete("/{id}", h.Remover)

	return r
}
