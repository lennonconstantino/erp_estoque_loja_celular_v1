// Package postgres implementa ports.CompraRepository sobre o schema compras usando pgx.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/ports"
)

// CompraRepository persiste compras no PostgreSQL.
type CompraRepository struct {
	pool *pgxpool.Pool
}

// NewCompraRepository cria o repositório.
func NewCompraRepository(pool *pgxpool.Pool) *CompraRepository {
	return &CompraRepository{pool: pool}
}

var _ ports.CompraRepository = (*CompraRepository)(nil)

// Criar insere o cabeçalho e todos os itens em uma única transação.
func (r *CompraRepository) Criar(ctx context.Context, c *domain.Compra) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `
		INSERT INTO compras.compra_master
			(id_compra, for_compra, nf_compra, dt_compra, val_compra, status_compra, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.FornecedorID, nilStr(c.NF), c.DtCompra, c.ValorTotal, string(c.Status),
		c.CriadoEm, c.AtualizadoEm,
	)
	if err != nil {
		return err
	}

	for _, item := range c.Itens {
		_, err = tx.Exec(ctx, `
			INSERT INTO compras.detalhe_compras
				(id_dt_compra, compra_id, pro_dt_compra, qtd_dt_compra,
				 pre_compra_dt_compra, pre_venda_dt_compra, margem_dt_compra)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			item.ID, item.CompraID, item.ProdutoID, item.Quantidade,
			item.PrecoCompra, item.PrecoVenda, item.Margem,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// BuscarPorID retorna a compra com seus itens.
func (r *CompraRepository) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Compra, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id_compra, for_compra, COALESCE(nf_compra,''), dt_compra,
		       val_compra, status_compra, created_at, updated_at
		FROM compras.compra_master WHERE id_compra=$1`, id)

	c, err := scanCompra(row)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id_dt_compra, compra_id, pro_dt_compra, qtd_dt_compra,
		       pre_compra_dt_compra, pre_venda_dt_compra, COALESCE(margem_dt_compra,0)
		FROM compras.detalhe_compras WHERE compra_id=$1
		ORDER BY id_dt_compra`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var det domain.DetalheCompra
		if err := rows.Scan(
			&det.ID, &det.CompraID, &det.ProdutoID, &det.Quantidade,
			&det.PrecoCompra, &det.PrecoVenda, &det.Margem,
		); err != nil {
			return nil, err
		}
		c.Itens = append(c.Itens, det)
	}
	return c, rows.Err()
}

// Listar retorna os cabeçalhos das compras (sem itens), opcionalmente filtrado por fornecedor.
func (r *CompraRepository) Listar(ctx context.Context, fornecedorID *uuid.UUID, limit, offset int) ([]domain.Compra, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id_compra, for_compra, COALESCE(nf_compra,''), dt_compra,
		       val_compra, status_compra, created_at, updated_at
		FROM compras.compra_master
		WHERE ($1::uuid IS NULL OR for_compra = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, fornecedorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Compra
	for rows.Next() {
		c, err := scanCompra(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

// AtualizarStatus altera o status de uma compra existente.
func (r *CompraRepository) AtualizarStatus(ctx context.Context, id uuid.UUID, status domain.StatusCompra) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE compras.compra_master SET status_compra=$2, updated_at=$3 WHERE id_compra=$1`,
		id, string(status), time.Now().UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCompraNaoEncontrada
	}
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanCompra(s scanner) (*domain.Compra, error) {
	var c domain.Compra
	var status string
	err := s.Scan(
		&c.ID, &c.FornecedorID, &c.NF, &c.DtCompra,
		&c.ValorTotal, &status, &c.CriadoEm, &c.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompraNaoEncontrada
		}
		return nil, err
	}
	c.Status = domain.StatusCompra(status)
	return &c, nil
}

func nilStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
