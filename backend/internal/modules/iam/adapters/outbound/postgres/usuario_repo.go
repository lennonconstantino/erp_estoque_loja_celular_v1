// Package postgres implementa as portas outbound do IAM sobre PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/ports"
)

// UsuarioRepository persiste usuários no schema iam.
type UsuarioRepository struct {
	pool *pgxpool.Pool
}

// NewUsuarioRepository cria o repositório.
func NewUsuarioRepository(pool *pgxpool.Pool) *UsuarioRepository {
	return &UsuarioRepository{pool: pool}
}

var _ ports.UsuarioRepository = (*UsuarioRepository)(nil)

const colunasUsuario = `
	id_usr, nome_usr, email_usr, senha_hash_usr,
	atv_usr, dt_ult_acesso_usr, created_at, updated_at`

func (r *UsuarioRepository) Create(ctx context.Context, u *domain.Usuario) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO iam.usuarios (id_usr, nome_usr, email_usr, senha_hash_usr, atv_usr)
		VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Nome, u.Email, u.SenhaHash, u.Ativo)
	return err
}

func (r *UsuarioRepository) Update(ctx context.Context, u *domain.Usuario) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE iam.usuarios
		SET nome_usr=$2, email_usr=$3, atv_usr=$4
		WHERE id_usr=$1`,
		u.ID, u.Nome, u.Email, u.Ativo)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNaoEncontrado
	}
	return nil
}

func (r *UsuarioRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Usuario, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+colunasUsuario+` FROM iam.usuarios WHERE id_usr=$1`, id)
	return scanUsuario(row)
}

func (r *UsuarioRepository) FindByEmail(ctx context.Context, email string) (*domain.Usuario, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+colunasUsuario+` FROM iam.usuarios WHERE email_usr=$1`,
		strings.ToLower(email))
	return scanUsuario(row)
}

func (r *UsuarioRepository) List(ctx context.Context, limit, offset int) ([]domain.Usuario, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+colunasUsuario+` FROM iam.usuarios
		ORDER BY nome_usr
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Usuario
	for rows.Next() {
		u, err := scanUsuario(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

func (r *UsuarioRepository) AtualizarSenha(ctx context.Context, id uuid.UUID, senhaHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE iam.usuarios SET senha_hash_usr=$2 WHERE id_usr=$1`, id, senhaHash)
	return err
}

func (r *UsuarioRepository) AtualizarUltAcesso(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE iam.usuarios SET dt_ult_acesso_usr=now() WHERE id_usr=$1`, id)
	return err
}

// CarregarPermissoes retorna os nomes dos papéis e os códigos de permissão
// vinculados ao usuário.
func (r *UsuarioRepository) CarregarPermissoes(ctx context.Context, userID uuid.UUID) ([]string, []string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT p.nome_pap, perm.codigo_per
		FROM iam.usuario_papeis up
		JOIN iam.papeis p        ON p.id_pap  = up.pap_id
		JOIN iam.papel_permissoes pp ON pp.pap_id = up.pap_id
		JOIN iam.permissoes perm ON perm.id_per = pp.per_id
		WHERE up.usr_id = $1`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	rolesSet := map[string]struct{}{}
	permsSet := map[string]struct{}{}

	for rows.Next() {
		var role, perm string
		if err := rows.Scan(&role, &perm); err != nil {
			return nil, nil, err
		}
		rolesSet[role] = struct{}{}
		permsSet[perm] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	roles := make([]string, 0, len(rolesSet))
	for k := range rolesSet {
		roles = append(roles, k)
	}
	perms := make([]string, 0, len(permsSet))
	for k := range permsSet {
		perms = append(perms, k)
	}
	return roles, perms, nil
}

// VincularPapeis substitui os papéis do usuário pelos nomes fornecidos.
func (r *UsuarioRepository) VincularPapeis(ctx context.Context, userID uuid.UUID, papeis []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `DELETE FROM iam.usuario_papeis WHERE usr_id=$1`, userID)
	if err != nil {
		return err
	}

	for _, nome := range papeis {
		_, err = tx.Exec(ctx, `
			INSERT INTO iam.usuario_papeis (usr_id, pap_id)
			SELECT $1, id_pap FROM iam.papeis WHERE nome_pap=$2
			ON CONFLICT DO NOTHING`, userID, nome)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUsuario(s scanner) (*domain.Usuario, error) {
	var u domain.Usuario
	err := s.Scan(
		&u.ID, &u.Nome, &u.Email, &u.SenhaHash,
		&u.Ativo, &u.UltAcesso, &u.CriadoEm, &u.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNaoEncontrado
		}
		return nil, err
	}
	return &u, nil
}
