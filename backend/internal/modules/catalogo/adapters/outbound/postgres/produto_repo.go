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

// ProdutoRepository persiste produtos no schema catalogo.
type ProdutoRepository struct {
	pool *pgxpool.Pool
}

// NewProdutoRepository cria o repositório.
func NewProdutoRepository(pool *pgxpool.Pool) *ProdutoRepository {
	return &ProdutoRepository{pool: pool}
}

var _ ports.ProdutoRepository = (*ProdutoRepository)(nil)

// colunasP lista as colunas de produto com cast ::float8 para evitar ambiguidade
// do pgx ao escanear NUMERIC(12,2) em float64.
const colunasP = `
	id_pro, cat_pro, desc_pro,
	p_custo_pro::float8, p_venda_pro::float8,
	estoque_m_pro, estoque_a_pro,
	garant_pro, COALESCE(mod_pro,''),
	disp_pro, atv_pro,
	created_at, updated_at`

func (r *ProdutoRepository) Create(ctx context.Context, p *domain.Produto) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO catalogo.produtos
			(id_pro, cat_pro, desc_pro, p_custo_pro, p_venda_pro,
			 estoque_m_pro, estoque_a_pro, garant_pro, mod_pro, disp_pro, atv_pro)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		p.ID, p.CategoriaID, p.Descricao, p.PrecoCusto, p.PrecoVenda,
		p.EstoqueMinimo, p.EstoqueAtual, p.GarantiaMeses, nilStr(p.Modelo),
		p.Disponivel, p.Ativo)
	return err
}

func (r *ProdutoRepository) Update(ctx context.Context, p *domain.Produto) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE catalogo.produtos SET
			cat_pro=$2, desc_pro=$3, p_custo_pro=$4, p_venda_pro=$5,
			estoque_m_pro=$6, garant_pro=$7, mod_pro=$8, atv_pro=$9
		WHERE id_pro=$1`,
		p.ID, p.CategoriaID, p.Descricao, p.PrecoCusto, p.PrecoVenda,
		p.EstoqueMinimo, p.GarantiaMeses, nilStr(p.Modelo), p.Ativo)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProdutoNaoEncontrado
	}
	return nil
}

func (r *ProdutoRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+colunasP+` FROM catalogo.produtos WHERE id_pro=$1`, id)
	return scanProduto(row)
}

func (r *ProdutoRepository) List(ctx context.Context, q string, categoriaID *uuid.UUID, limit, offset int) ([]domain.Produto, error) {
	var (
		rows interface {
			Next() bool
			Scan(dest ...any) error
			Err() error
			Close()
		}
		err error
	)
	base := `SELECT ` + colunasP + ` FROM catalogo.produtos`
	if categoriaID != nil {
		rows, err = r.pool.Query(ctx,
			base+` WHERE cat_pro=$1 AND ($2='' OR desc_pro ILIKE '%'||$2||'%') ORDER BY desc_pro LIMIT $3 OFFSET $4`,
			categoriaID, q, limit, offset)
	} else {
		rows, err = r.pool.Query(ctx,
			base+` WHERE ($1='' OR desc_pro ILIKE '%'||$1||'%') ORDER BY desc_pro LIMIT $2 OFFSET $3`,
			q, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Produto
	for rows.Next() {
		p, err := scanProduto(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func (r *ProdutoRepository) UpdateSaldo(ctx context.Context, id uuid.UUID, novoSaldo int, disponivel bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE catalogo.produtos SET estoque_a_pro=$2, disp_pro=$3 WHERE id_pro=$1`,
		id, novoSaldo, disponivel)
	return err
}

func scanProduto(s scanner) (*domain.Produto, error) {
	var p domain.Produto
	err := s.Scan(
		&p.ID, &p.CategoriaID, &p.Descricao,
		&p.PrecoCusto, &p.PrecoVenda,
		&p.EstoqueMinimo, &p.EstoqueAtual,
		&p.GarantiaMeses, &p.Modelo,
		&p.Disponivel, &p.Ativo,
		&p.CriadoEm, &p.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProdutoNaoEncontrado
		}
		return nil, err
	}
	return &p, nil
}

func nilStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
