package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/ports"
)

// TokenStore persiste refresh tokens no schema iam.
type TokenStore struct {
	pool *pgxpool.Pool
}

// NewTokenStore cria o store.
func NewTokenStore(pool *pgxpool.Pool) *TokenStore {
	return &TokenStore{pool: pool}
}

var _ ports.TokenStore = (*TokenStore)(nil)

func (s *TokenStore) Criar(ctx context.Context, rt *domain.RefreshToken) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO iam.refresh_tokens (id_rt, usr_id, token_hash, expira_em)
		VALUES ($1, $2, $3, $4)`,
		rt.ID, rt.UsuarioID, rt.TokenHash, rt.ExpiraEm)
	return err
}

func (s *TokenStore) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id_rt, usr_id, token_hash, expira_em, revogado, created_at
		FROM iam.refresh_tokens
		WHERE token_hash=$1`, hash)

	var rt domain.RefreshToken
	err := row.Scan(&rt.ID, &rt.UsuarioID, &rt.TokenHash, &rt.ExpiraEm, &rt.Revogado, &rt.CriadoEm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTokenInvalido
		}
		return nil, err
	}
	return &rt, nil
}

func (s *TokenStore) Revogar(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE iam.refresh_tokens SET revogado=TRUE WHERE id_rt=$1`, id)
	return err
}
