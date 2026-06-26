// Package postgres implementa ports.VendaRepository sobre o schema vendas usando pgx.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/ports"
)

// VendaRepository persiste vendas no PostgreSQL.
type VendaRepository struct {
	pool *pgxpool.Pool
}

// NewVendaRepository cria o repositório.
func NewVendaRepository(pool *pgxpool.Pool) *VendaRepository {
	return &VendaRepository{pool: pool}
}

var _ ports.VendaRepository = (*VendaRepository)(nil)

// Criar insere o cabeçalho e todos os itens em uma única transação.
func (r *VendaRepository) Criar(ctx context.Context, v *domain.Venda) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `
		INSERT INTO vendas.venda_master
			(id_venda, dt_venda, val_venda, dsc_venda, forma_pgto_venda,
			 cli_venda, c_final_venda, doc_fiscal_venda, status_venda, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		v.ID, v.DtVenda, v.ValorTotal, v.Desconto, string(v.FormaPgto),
		v.ClienteID, v.ConsumidorFinal, string(v.DocFiscal), string(v.Status),
		v.CriadoEm, v.AtualizadoEm,
	)
	if err != nil {
		return err
	}

	for _, item := range v.Itens {
		_, err = tx.Exec(ctx, `
			INSERT INTO vendas.detalhe_vendas
				(id_dt_venda, venda_id, pro_venda, qtd_venda, pre_venda_dt)
			VALUES ($1,$2,$3,$4,$5)`,
			item.ID, item.VendaID, item.ProdutoID, item.Quantidade, item.PrecoUnitario,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// BuscarPorID retorna a venda com seus itens.
func (r *VendaRepository) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Venda, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id_venda, dt_venda, val_venda::float8, dsc_venda::float8,
		       forma_pgto_venda, cli_venda, c_final_venda, doc_fiscal_venda,
		       status_venda, created_at, updated_at
		FROM vendas.venda_master WHERE id_venda=$1`, id)

	v, err := scanVenda(row)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id_dt_venda, venda_id, pro_venda, qtd_venda, pre_venda_dt::float8
		FROM vendas.detalhe_vendas WHERE venda_id=$1
		ORDER BY id_dt_venda`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var det domain.DetalheVenda
		if err := rows.Scan(&det.ID, &det.VendaID, &det.ProdutoID, &det.Quantidade, &det.PrecoUnitario); err != nil {
			return nil, err
		}
		v.Itens = append(v.Itens, det)
	}
	return v, rows.Err()
}

// Listar retorna cabeçalhos das vendas (sem itens), opcionalmente filtrado por cliente.
func (r *VendaRepository) Listar(ctx context.Context, clienteID *uuid.UUID, limit, offset int) ([]domain.Venda, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id_venda, dt_venda, val_venda::float8, dsc_venda::float8,
		       forma_pgto_venda, cli_venda, c_final_venda, doc_fiscal_venda,
		       status_venda, created_at, updated_at
		FROM vendas.venda_master
		WHERE ($1::uuid IS NULL OR cli_venda = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, clienteID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Venda
	for rows.Next() {
		v, err := scanVenda(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

// AtualizarStatus altera o status de uma venda existente.
func (r *VendaRepository) AtualizarStatus(ctx context.Context, id uuid.UUID, status domain.StatusVenda) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE vendas.venda_master SET status_venda=$2, updated_at=$3 WHERE id_venda=$1`,
		id, string(status), time.Now().UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrVendaNaoEncontrada
	}
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanVenda(s scanner) (*domain.Venda, error) {
	var (
		v          domain.Venda
		formaPgto  string
		docFiscal  string
		status     string
	)
	err := s.Scan(
		&v.ID, &v.DtVenda, &v.ValorTotal, &v.Desconto,
		&formaPgto, &v.ClienteID, &v.ConsumidorFinal, &docFiscal,
		&status, &v.CriadoEm, &v.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVendaNaoEncontrada
		}
		return nil, err
	}
	v.FormaPgto = domain.FormaPgto(formaPgto)
	v.DocFiscal = domain.DocFiscal(docFiscal)
	v.Status = domain.StatusVenda(status)
	return &v, nil
}
