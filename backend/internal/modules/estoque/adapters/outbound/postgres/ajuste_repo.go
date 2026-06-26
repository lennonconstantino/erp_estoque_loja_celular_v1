package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/ports"
)

// AjusteRepository implementa ports.AjusteRepository via pgx.
type AjusteRepository struct {
	pool *pgxpool.Pool
}

// NewAjusteRepository cria o repositório de ajustes.
func NewAjusteRepository(pool *pgxpool.Pool) *AjusteRepository {
	return &AjusteRepository{pool: pool}
}

// Conformidade em tempo de compilação.
var _ ports.AjusteRepository = (*AjusteRepository)(nil)

// Create insere um ajuste (append-only; trigger bloqueia UPDATE/DELETE no banco).
func (r *AjusteRepository) Create(ctx context.Context, a *domain.Ajuste) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO estoque.ajustes
			(id_ajs, pro_ajs, qtd_entrada_ajs, qtd_saida_ajs, mot_ajs, resp_ajs, dt_ajs)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		a.ID, a.ProdutoID, a.QtdEntrada, a.QtdSaida, a.Motivo, a.Responsavel, a.CriadoEm,
	)
	return err
}

// ListByProduto retorna ajustes de um produto, do mais recente para o mais antigo.
func (r *AjusteRepository) ListByProduto(ctx context.Context, produtoID uuid.UUID, limit, offset int) ([]domain.Ajuste, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id_ajs, pro_ajs, qtd_entrada_ajs, qtd_saida_ajs, mot_ajs, resp_ajs, dt_ajs
		FROM estoque.ajustes
		WHERE pro_ajs = $1
		ORDER BY dt_ajs DESC
		LIMIT $2 OFFSET $3`,
		produtoID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Ajuste
	for rows.Next() {
		var a domain.Ajuste
		if err := rows.Scan(
			&a.ID, &a.ProdutoID, &a.QtdEntrada, &a.QtdSaida,
			&a.Motivo, &a.Responsavel, &a.CriadoEm,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
