// Package ports declara os contratos do contexto IAM.
package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/domain"
)

// LoginInput é o comando de autenticação.
type LoginInput struct {
	Email string
	Senha string
}

// LoginOutput carrega os tokens emitidos após autenticação bem-sucedida.
type LoginOutput struct {
	AccessToken  string
	RefreshToken string
}

// CriarUsuarioInput é o comando de criação de usuário.
type CriarUsuarioInput struct {
	Nome   string
	Email  string
	Senha  string
	Papeis []string // nomes dos papéis (ex.: "ADMIN", "VENDEDOR")
}

// AtualizarUsuarioInput é o comando de atualização de usuário.
type AtualizarUsuarioInput struct {
	Nome  string
	Email string
	Ativo *bool  // nil = não altera
	Senha string // vazio = não altera senha
}

// AuthService é a porta de entrada (caso de uso) oferecida pelo módulo IAM.
type AuthService interface {
	Login(ctx context.Context, in LoginInput) (LoginOutput, error)
	Refresh(ctx context.Context, refreshToken string) (LoginOutput, error)
	Logout(ctx context.Context, refreshToken string) error
	CriarUsuario(ctx context.Context, in CriarUsuarioInput) (*domain.Usuario, error)
	AtualizarUsuario(ctx context.Context, id uuid.UUID, in AtualizarUsuarioInput) (*domain.Usuario, error)
	ListarUsuarios(ctx context.Context, limit, offset int) ([]domain.Usuario, error)
}
