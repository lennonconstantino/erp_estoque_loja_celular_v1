// Package application contém os casos de uso do contexto IAM.
// Orquestra domínio e portas de saída; não conhece HTTP nem detalhes de banco.
package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
)

// Service implementa ports.AuthService.
type Service struct {
	repo       ports.UsuarioRepository
	tokens     ports.TokenStore
	authMgr    *auth.Manager
	refreshTTL time.Duration
}

// NewService injeta as dependências.
func NewService(repo ports.UsuarioRepository, tokens ports.TokenStore, authMgr *auth.Manager, refreshTTL time.Duration) *Service {
	return &Service{repo: repo, tokens: tokens, authMgr: authMgr, refreshTTL: refreshTTL}
}

var _ ports.AuthService = (*Service)(nil)

// Login autentica o usuário e retorna um par de tokens.
func (s *Service) Login(ctx context.Context, in ports.LoginInput) (ports.LoginOutput, error) {
	u, err := s.repo.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNaoEncontrado) {
			return ports.LoginOutput{}, domain.ErrCredenciaisInvalidas
		}
		return ports.LoginOutput{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.SenhaHash), []byte(in.Senha)); err != nil {
		return ports.LoginOutput{}, domain.ErrCredenciaisInvalidas
	}

	if !u.Ativo {
		return ports.LoginOutput{}, domain.ErrUsuarioInativo
	}

	roles, perms, err := s.repo.CarregarPermissoes(ctx, u.ID)
	if err != nil {
		return ports.LoginOutput{}, err
	}

	accessToken, err := s.authMgr.GenerateAccess(u.ID.String(), roles, perms)
	if err != nil {
		return ports.LoginOutput{}, err
	}

	rawRT, rt, err := s.novoRefreshToken(u.ID)
	if err != nil {
		return ports.LoginOutput{}, err
	}
	if err := s.tokens.Criar(ctx, rt); err != nil {
		return ports.LoginOutput{}, err
	}

	_ = s.repo.AtualizarUltAcesso(ctx, u.ID) // best-effort

	return ports.LoginOutput{AccessToken: accessToken, RefreshToken: rawRT}, nil
}

// Refresh valida o refresh token, emite novos tokens e rotaciona o armazenado.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (ports.LoginOutput, error) {
	hash := hashToken(refreshToken)
	rt, err := s.tokens.FindByHash(ctx, hash)
	if err != nil || rt.Revogado || time.Now().After(rt.ExpiraEm) {
		return ports.LoginOutput{}, domain.ErrTokenInvalido
	}

	u, err := s.repo.FindByID(ctx, rt.UsuarioID)
	if err != nil {
		return ports.LoginOutput{}, err
	}
	if !u.Ativo {
		return ports.LoginOutput{}, domain.ErrUsuarioInativo
	}

	if err := s.tokens.Revogar(ctx, rt.ID); err != nil {
		return ports.LoginOutput{}, err
	}

	roles, perms, err := s.repo.CarregarPermissoes(ctx, u.ID)
	if err != nil {
		return ports.LoginOutput{}, err
	}

	accessToken, err := s.authMgr.GenerateAccess(u.ID.String(), roles, perms)
	if err != nil {
		return ports.LoginOutput{}, err
	}

	rawRT, novoRT, err := s.novoRefreshToken(u.ID)
	if err != nil {
		return ports.LoginOutput{}, err
	}
	if err := s.tokens.Criar(ctx, novoRT); err != nil {
		return ports.LoginOutput{}, err
	}

	return ports.LoginOutput{AccessToken: accessToken, RefreshToken: rawRT}, nil
}

// Logout revoga o refresh token informado.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hash := hashToken(refreshToken)
	rt, err := s.tokens.FindByHash(ctx, hash)
	if err != nil {
		return nil // token já inválido — logout silencioso
	}
	return s.tokens.Revogar(ctx, rt.ID)
}

// CriarUsuario valida, hasha a senha e persiste o novo usuário.
func (s *Service) CriarUsuario(ctx context.Context, in ports.CriarUsuarioInput) (*domain.Usuario, error) {
	if len(in.Senha) < 8 {
		return nil, domain.ErrSenhaFraca
	}

	if existing, err := s.repo.FindByEmail(ctx, in.Email); err == nil && existing != nil {
		return nil, domain.ErrEmailJaCadastrado
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u, err := domain.NovoUsuario(in.Nome, in.Email, string(hash))
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	if len(in.Papeis) > 0 {
		if err := s.repo.VincularPapeis(ctx, u.ID, in.Papeis); err != nil {
			return nil, err
		}
	}

	return u, nil
}

// AtualizarUsuario altera dados do usuário.
func (s *Service) AtualizarUsuario(ctx context.Context, id uuid.UUID, in ports.AtualizarUsuarioInput) (*domain.Usuario, error) {
	if in.Senha != "" && len(in.Senha) < 8 {
		return nil, domain.ErrSenhaFraca
	}

	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.Nome != "" || in.Email != "" {
		nome := in.Nome
		if nome == "" {
			nome = u.Nome
		}
		email := in.Email
		if email == "" {
			email = u.Email
		}
		if err := u.AtualizarDados(nome, email); err != nil {
			return nil, err
		}
	}

	if in.Ativo != nil {
		u.Ativo = *in.Ativo
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}

	if in.Senha != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Senha), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		if err := s.repo.AtualizarSenha(ctx, id, string(hash)); err != nil {
			return nil, err
		}
	}

	return u, nil
}

// ListarUsuarios retorna a lista paginada de usuários.
func (s *Service) ListarUsuarios(ctx context.Context, limit, offset int) ([]domain.Usuario, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

// novoRefreshToken gera um token aleatório e retorna o valor bruto e a entidade
// pronta para persistir (com o hash).
func (s *Service) novoRefreshToken(userID uuid.UUID) (raw string, rt *domain.RefreshToken, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	raw = hex.EncodeToString(b)
	rt = &domain.RefreshToken{
		ID:        uuid.New(),
		UsuarioID: userID,
		TokenHash: hashToken(raw),
		ExpiraEm:  time.Now().UTC().Add(s.refreshTTL),
		CriadoEm: time.Now().UTC(),
	}
	return
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
