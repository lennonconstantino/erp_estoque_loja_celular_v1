// Package postgres implementa ports.ClienteRepository sobre o schema
// clientes.clientes usando pgx.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/ports"
)

// ClienteRepository persiste clientes no PostgreSQL.
type ClienteRepository struct {
	pool *pgxpool.Pool
}

// NewClienteRepository cria o repositório.
func NewClienteRepository(pool *pgxpool.Pool) *ClienteRepository {
	return &ClienteRepository{pool: pool}
}

var _ ports.ClienteRepository = (*ClienteRepository)(nil)

const colunas = `
	id_cli, cpf_cli, nome_cli, email_cli,
	COALESCE(tel_cli,''), COALESCE(cep_cli,''), COALESCE(rua_cli,''),
	COALESCE(num_cli,''), COALESCE(comp_cli,''), COALESCE(bai_cli,''),
	COALESCE(cit_cli,''), COALESCE(uf_cli,''),
	dt_ult_comp_cli, atv_cli, created_at, updated_at`

func (r *ClienteRepository) Create(ctx context.Context, c *domain.Cliente) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO clientes.clientes
			(id_cli, cpf_cli, nome_cli, email_cli, tel_cli, cep_cli, rua_cli,
			 num_cli, comp_cli, bai_cli, cit_cli, uf_cli, dt_ult_comp_cli, atv_cli)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		c.ID, c.CPF, c.Nome, c.Email, nilStr(c.Telefone), nilStr(c.CEP), nilStr(c.Rua),
		nilStr(c.Numero), nilStr(c.Complemento), nilStr(c.Bairro), nilStr(c.Cidade),
		nilStr(c.UF), c.UltimaCompra, c.Ativo)
	return err
}

func (r *ClienteRepository) Update(ctx context.Context, c *domain.Cliente) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE clientes.clientes SET
			nome_cli=$2, email_cli=$3, tel_cli=$4, cep_cli=$5, rua_cli=$6,
			num_cli=$7, comp_cli=$8, bai_cli=$9, cit_cli=$10, uf_cli=$11, atv_cli=$12
		WHERE id_cli=$1`,
		c.ID, c.Nome, c.Email, nilStr(c.Telefone), nilStr(c.CEP), nilStr(c.Rua),
		nilStr(c.Numero), nilStr(c.Complemento), nilStr(c.Bairro), nilStr(c.Cidade),
		nilStr(c.UF), c.Ativo)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNaoEncontrado
	}
	return nil
}

func (r *ClienteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM clientes.clientes WHERE id_cli=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNaoEncontrado
	}
	return nil
}

func (r *ClienteRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Cliente, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+colunas+` FROM clientes.clientes WHERE id_cli=$1`, id)
	return scanCliente(row)
}

func (r *ClienteRepository) FindByCPF(ctx context.Context, cpf string) (*domain.Cliente, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+colunas+` FROM clientes.clientes WHERE cpf_cli=$1`, cpf)
	return scanCliente(row)
}

func (r *ClienteRepository) List(ctx context.Context, q string, limit, offset int) ([]domain.Cliente, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+colunas+` FROM clientes.clientes
		WHERE ($1 = '' OR nome_cli ILIKE '%'||$1||'%' OR cpf_cli LIKE $1||'%')
		ORDER BY nome_cli
		LIMIT $2 OFFSET $3`, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Cliente
	for rows.Next() {
		c, err := scanCliente(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

// AtualizarUltimaVenda atualiza o campo dt_ult_comp_cli quando uma venda é confirmada.
func (r *ClienteRepository) AtualizarUltimaVenda(ctx context.Context, id uuid.UUID, data time.Time) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE clientes.clientes SET dt_ult_comp_cli=$2 WHERE id_cli=$1`, id, data)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNaoEncontrado
	}
	return nil
}

// scanner abstrai pgx.Row e pgx.Rows (ambos têm Scan).
type scanner interface {
	Scan(dest ...any) error
}

func scanCliente(s scanner) (*domain.Cliente, error) {
	var c domain.Cliente
	err := s.Scan(
		&c.ID, &c.CPF, &c.Nome, &c.Email,
		&c.Telefone, &c.CEP, &c.Rua,
		&c.Numero, &c.Complemento, &c.Bairro,
		&c.Cidade, &c.UF,
		&c.UltimaCompra, &c.Ativo, &c.CriadoEm, &c.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNaoEncontrado
		}
		return nil, err
	}
	return &c, nil
}

// nilStr converte "" em NULL para colunas opcionais.
func nilStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
