package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/ports"
)

// MovimentacaoRepository implementa ports.MovimentacaoRepository via pgx.
type MovimentacaoRepository struct {
	pool *pgxpool.Pool
}

// NewMovimentacaoRepository cria o repositório de movimentações.
func NewMovimentacaoRepository(pool *pgxpool.Pool) *MovimentacaoRepository {
	return &MovimentacaoRepository{pool: pool}
}

// Conformidade em tempo de compilação.
var _ ports.MovimentacaoRepository = (*MovimentacaoRepository)(nil)

// Create insere uma movimentação no razão (append-only).
func (r *MovimentacaoRepository) Create(ctx context.Context, m *domain.Movimentacao) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO estoque.movimentacoes
			(id_mov, pro_mov, tipo_mov, qtd_mov, saldo_ant_mov, saldo_atu_mov,
			 origem_tipo, origem_id, resp_mov, dt_mov)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		m.ID, m.ProdutoID, string(m.Tipo), m.Quantidade,
		m.SaldoAntes, m.SaldoDepois,
		m.OrigemTipo, m.OrigemID, m.Responsavel, m.CriadoEm,
	)
	return err
}

// ListByProduto retorna movimentações de um produto, do mais recente para o mais antigo.
func (r *MovimentacaoRepository) ListByProduto(ctx context.Context, produtoID uuid.UUID, limit, offset int) ([]domain.Movimentacao, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id_mov, pro_mov, tipo_mov, qtd_mov, saldo_ant_mov, saldo_atu_mov,
		       COALESCE(origem_tipo,''), origem_id, resp_mov, dt_mov
		FROM estoque.movimentacoes
		WHERE pro_mov = $1
		ORDER BY dt_mov DESC
		LIMIT $2 OFFSET $3`,
		produtoID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Movimentacao
	for rows.Next() {
		var m domain.Movimentacao
		var tipoStr string
		if err := rows.Scan(
			&m.ID, &m.ProdutoID, &tipoStr, &m.Quantidade,
			&m.SaldoAntes, &m.SaldoDepois,
			&m.OrigemTipo, &m.OrigemID, &m.Responsavel, &m.CriadoEm,
		); err != nil {
			return nil, err
		}
		m.Tipo = domain.TipoMovimentacao(tipoStr)
		out = append(out, m)
	}
	return out, rows.Err()
}

// ConsultarSaldoAtual retorna o saldo após a última movimentação do produto.
// Retorna 0 se não houver movimentações anteriores (produto recém-criado).
func (r *MovimentacaoRepository) ConsultarSaldoAtual(ctx context.Context, produtoID uuid.UUID) (int, error) {
	var saldo int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(saldo_atu_mov, 0)
		FROM estoque.movimentacoes
		WHERE pro_mov = $1
		ORDER BY dt_mov DESC
		LIMIT 1`,
		produtoID,
	).Scan(&saldo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return saldo, nil
}
