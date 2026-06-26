package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/domain"
)

// UsuarioRepository é a porta de saída de persistência de usuários.
type UsuarioRepository interface {
	Create(ctx context.Context, u *domain.Usuario) error
	Update(ctx context.Context, u *domain.Usuario) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Usuario, error)
	FindByEmail(ctx context.Context, email string) (*domain.Usuario, error)
	List(ctx context.Context, limit, offset int) ([]domain.Usuario, error)
	AtualizarSenha(ctx context.Context, id uuid.UUID, senhaHash string) error
	AtualizarUltAcesso(ctx context.Context, id uuid.UUID) error
	// CarregarPermissoes retorna os papéis e permissões vinculados ao usuário.
	CarregarPermissoes(ctx context.Context, userID uuid.UUID) (roles []string, perms []string, err error)
	// VincularPapeis substitui os papéis do usuário pelos nomes fornecidos.
	VincularPapeis(ctx context.Context, userID uuid.UUID, papeis []string) error
}

// TokenStore é a porta de saída para armazenamento de refresh tokens.
type TokenStore interface {
	Criar(ctx context.Context, rt *domain.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	Revogar(ctx context.Context, id uuid.UUID) error
}
