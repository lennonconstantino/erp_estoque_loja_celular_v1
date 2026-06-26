// Package postgres implementa os repositórios do contexto catálogo sobre pgx.
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/ports"
)

// CategoriaRepository persiste categorias no schema catalogo.
type CategoriaRepository struct {
	pool *pgxpool.Pool
}

// NewCategoriaRepository cria o repositório.
func NewCategoriaRepository(pool *pgxpool.Pool) *CategoriaRepository {
	return &CategoriaRepository{pool: pool}
}

var _ ports.CategoriaRepository = (*CategoriaRepository)(nil)

func (r *CategoriaRepository) Create(ctx context.Context, c *domain.Categoria) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO catalogo.categorias (id_cat, desc_cat) VALUES ($1, $2)`,
		c.ID, c.Descricao)
	return err
}

func (r *CategoriaRepository) Update(ctx context.Context, c *domain.Categoria) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE catalogo.categorias SET desc_cat=$2 WHERE id_cat=$1`,
		c.ID, c.Descricao)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCategoriaNaoEncontrada
	}
	return nil
}

func (r *CategoriaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM catalogo.categorias WHERE id_cat=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCategoriaNaoEncontrada
	}
	return nil
}

func (r *CategoriaRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Categoria, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id_cat, desc_cat, created_at FROM catalogo.categorias WHERE id_cat=$1`, id)
	return scanCategoria(row)
}

func (r *CategoriaRepository) FindByDescricao(ctx context.Context, descricao string) (*domain.Categoria, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id_cat, desc_cat, created_at FROM catalogo.categorias WHERE lower(desc_cat)=lower($1)`, descricao)
	return scanCategoria(row)
}

func (r *CategoriaRepository) List(ctx context.Context, q string, limit, offset int) ([]domain.Categoria, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id_cat, desc_cat, created_at FROM catalogo.categorias
		WHERE ($1 = '' OR desc_cat ILIKE '%'||$1||'%')
		ORDER BY desc_cat
		LIMIT $2 OFFSET $3`, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Categoria
	for rows.Next() {
		c, err := scanCategoria(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (r *CategoriaRepository) HasProdutos(ctx context.Context, id uuid.UUID) (bool, error) {
	var tem bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM catalogo.produtos WHERE cat_pro=$1)`, id).Scan(&tem)
	return tem, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCategoria(s scanner) (*domain.Categoria, error) {
	var c domain.Categoria
	err := s.Scan(&c.ID, &c.Descricao, &c.CriadoEm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoriaNaoEncontrada
		}
		return nil, err
	}
	return &c, nil
}
