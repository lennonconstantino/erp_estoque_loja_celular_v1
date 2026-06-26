// Package postgres implementa ports.FornecedorRepository sobre o schema
// fornecedores.fornecedores usando pgx.
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/ports"
)

// FornecedorRepository persiste fornecedores no PostgreSQL.
type FornecedorRepository struct {
	pool *pgxpool.Pool
}

// NewFornecedorRepository cria o repositório.
func NewFornecedorRepository(pool *pgxpool.Pool) *FornecedorRepository {
	return &FornecedorRepository{pool: pool}
}

var _ ports.FornecedorRepository = (*FornecedorRepository)(nil)

const colunas = `
	id_for, cnpj_for, razao_for, nome_fant_for, email_for,
	tel1_for, COALESCE(tel2_for,''),
	COALESCE(cep_for,''), COALESCE(rua_for,''), COALESCE(num_for,''),
	COALESCE(comp_for,''), COALESCE(bai_for,''), COALESCE(cit_for,''), COALESCE(uf_for,''),
	comercial_for, COALESCE(financeiro_for,''),
	dt_ult_comp_for, atv_for, created_at, updated_at`

func (r *FornecedorRepository) Create(ctx context.Context, f *domain.Fornecedor) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO fornecedores.fornecedores
			(id_for, cnpj_for, razao_for, nome_fant_for, email_for,
			 tel1_for, tel2_for, cep_for, rua_for, num_for,
			 comp_for, bai_for, cit_for, uf_for,
			 comercial_for, financeiro_for, dt_ult_comp_for, atv_for)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		f.ID, f.CNPJ, f.RazaoSocial, f.NomeFantasia, f.Email,
		f.Telefone1, nilStr(f.Telefone2),
		nilStr(f.CEP), nilStr(f.Rua), nilStr(f.Numero),
		nilStr(f.Complemento), nilStr(f.Bairro), nilStr(f.Cidade), nilStr(f.UF),
		f.Comercial, nilStr(f.Financeiro), f.UltimaCompra, f.Ativo)
	return err
}

func (r *FornecedorRepository) Update(ctx context.Context, f *domain.Fornecedor) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE fornecedores.fornecedores SET
			razao_for=$2, nome_fant_for=$3, email_for=$4,
			tel1_for=$5, tel2_for=$6, cep_for=$7, rua_for=$8, num_for=$9,
			comp_for=$10, bai_for=$11, cit_for=$12, uf_for=$13,
			comercial_for=$14, financeiro_for=$15, atv_for=$16
		WHERE id_for=$1`,
		f.ID, f.RazaoSocial, f.NomeFantasia, f.Email,
		f.Telefone1, nilStr(f.Telefone2),
		nilStr(f.CEP), nilStr(f.Rua), nilStr(f.Numero),
		nilStr(f.Complemento), nilStr(f.Bairro), nilStr(f.Cidade), nilStr(f.UF),
		f.Comercial, nilStr(f.Financeiro), f.Ativo)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNaoEncontrado
	}
	return nil
}

func (r *FornecedorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM fornecedores.fornecedores WHERE id_for=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNaoEncontrado
	}
	return nil
}

func (r *FornecedorRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Fornecedor, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+colunas+` FROM fornecedores.fornecedores WHERE id_for=$1`, id)
	return scanFornecedor(row)
}

func (r *FornecedorRepository) FindByCNPJ(ctx context.Context, cnpj string) (*domain.Fornecedor, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+colunas+` FROM fornecedores.fornecedores WHERE cnpj_for=$1`, cnpj)
	return scanFornecedor(row)
}

func (r *FornecedorRepository) List(ctx context.Context, q string, limit, offset int) ([]domain.Fornecedor, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+colunas+` FROM fornecedores.fornecedores
		WHERE ($1 = '' OR razao_for ILIKE '%'||$1||'%' OR nome_fant_for ILIKE '%'||$1||'%' OR cnpj_for LIKE $1||'%')
		ORDER BY razao_for
		LIMIT $2 OFFSET $3`, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Fornecedor
	for rows.Next() {
		f, err := scanFornecedor(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *f)
	}
	return out, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFornecedor(s scanner) (*domain.Fornecedor, error) {
	var f domain.Fornecedor
	err := s.Scan(
		&f.ID, &f.CNPJ, &f.RazaoSocial, &f.NomeFantasia, &f.Email,
		&f.Telefone1, &f.Telefone2,
		&f.CEP, &f.Rua, &f.Numero,
		&f.Complemento, &f.Bairro, &f.Cidade, &f.UF,
		&f.Comercial, &f.Financeiro,
		&f.UltimaCompra, &f.Ativo, &f.CriadoEm, &f.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNaoEncontrado
		}
		return nil, err
	}
	return &f, nil
}

func nilStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
