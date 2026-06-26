// Package iam monta (DI) o bounded context de identidade e acesso.
package iam

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/adapters/inbound/http"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/adapters/outbound/postgres"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// Module expõe o roteador do contexto para montagem no servidor.
type Module struct {
	router chi.Router
}

// New monta o módulo IAM.
func New(pool *pgxpool.Pool, authMgr *auth.Manager, refreshTTL time.Duration) *Module {
	repo := postgres.NewUsuarioRepository(pool)
	store := postgres.NewTokenStore(pool)
	svc := application.NewService(repo, store, authMgr, refreshTTL)
	handler := httpadapter.NewHandler(svc)

	return &Module{router: httpadapter.NewRouter(handler, authMgr)}
}

// Router retorna o roteador para ser montado sob /api/v1.
func (m *Module) Router() chi.Router { return m.router }
