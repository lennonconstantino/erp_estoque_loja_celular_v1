package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// NewRouter monta as rotas do módulo IAM.
// Autenticação (login/refresh/logout) são públicas.
// Gerenciamento de usuários exige iam:admin.
func NewRouter(h *Handler, authMgr *auth.Manager) chi.Router {
	r := chi.NewRouter()

	// rotas públicas
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)
	r.Post("/auth/logout", h.Logout)

	// rotas protegidas (somente ADMIN)
	r.Group(func(r chi.Router) {
		r.Use(authMgr.Authenticate)
		r.Use(auth.RequirePerm("iam:admin"))
		r.Get("/usuarios", h.ListarUsuarios)
		r.Post("/usuarios", h.CriarUsuario)
		r.Patch("/usuarios/{id}", h.AtualizarUsuario)
	})

	return r
}
