// Package postgres implementa RelatorioRepository consultando o banco de dados.
// As queries cruzam schemas (catalogo, vendas, compras) usando LEFT JOIN via ID —
// sem foreign keys entre schemas, conforme regra de isolamento de contextos.
package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/ports"
)

// RelatorioRepository implementa ports.RelatorioRepository com pgx.
type RelatorioRepository struct {
	pool *pgxpool.Pool
}

var _ ports.RelatorioRepository = (*RelatorioRepository)(nil)

// NewRelatorioRepository cria o repositório.
func NewRelatorioRepository(pool *pgxpool.Pool) *RelatorioRepository {
	return &RelatorioRepository{pool: pool}
}

// ListarAbaixoDoMinimo retorna produtos ativos com estoque abaixo do mínimo.
func (r *RelatorioRepository) ListarAbaixoDoMinimo(ctx context.Context) ([]domain.ProdutoAbaixoMinimo, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id_pro, desc_pro, estoque_a_pro, estoque_m_pro,
		       (estoque_m_pro - estoque_a_pro) AS defasagem
		FROM   catalogo.produtos
		WHERE  atv_pro = TRUE AND estoque_a_pro < estoque_m_pro
		ORDER  BY defasagem DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []domain.ProdutoAbaixoMinimo
	for rows.Next() {
		var p domain.ProdutoAbaixoMinimo
		if err := rows.Scan(&p.ID, &p.Descricao, &p.EstoqueAtual, &p.EstoqueMinimo, &p.Defasagem); err != nil {
			return nil, err
		}
		lista = append(lista, p)
	}
	return lista, rows.Err()
}

// ListarMaisVendidos retorna os `limite` produtos mais vendidos no período.
func (r *RelatorioRepository) ListarMaisVendidos(ctx context.Context, de, ate time.Time, limite int) ([]domain.ProdutoVendido, error) {
	return r.listarVendidosPorOrdem(ctx, de, ate, limite, "DESC")
}

// ListarMenosVendidos retorna os `limite` produtos menos vendidos no período.
func (r *RelatorioRepository) ListarMenosVendidos(ctx context.Context, de, ate time.Time, limite int) ([]domain.ProdutoVendido, error) {
	return r.listarVendidosPorOrdem(ctx, de, ate, limite, "ASC")
}

func (r *RelatorioRepository) listarVendidosPorOrdem(
	ctx context.Context, de, ate time.Time, limite int, ordem string,
) ([]domain.ProdutoVendido, error) {
	// Construção da query com ordem variável; ordem só pode ser "ASC" ou "DESC"
	// (valores fixos no código — sem interpolação de entrada do usuário).
	q := `
		SELECT dv.pro_venda,
		       COALESCE(p.desc_pro, '(produto removido)'),
		       SUM(dv.qtd_venda)::int                             AS total_vendido,
		       SUM(dv.qtd_venda * dv.pre_venda_dt)::float8       AS total_valor
		FROM   vendas.detalhe_vendas  dv
		JOIN   vendas.venda_master    vm ON vm.id_venda = dv.venda_id
		LEFT   JOIN catalogo.produtos  p  ON p.id_pro   = dv.pro_venda
		WHERE  vm.status_venda = 'CONFIRMADA'
		  AND  vm.dt_venda >= $1 AND vm.dt_venda <= $2
		GROUP  BY dv.pro_venda, p.desc_pro
		ORDER  BY total_vendido ` + ordem + `
		LIMIT  $3`

	rows, err := r.pool.Query(ctx, q, de, ate, limite)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []domain.ProdutoVendido
	for rows.Next() {
		var pv domain.ProdutoVendido
		var id uuid.UUID
		if err := rows.Scan(&id, &pv.Descricao, &pv.TotalVendido, &pv.TotalValor); err != nil {
			return nil, err
		}
		pv.ProdutoID = id
		lista = append(lista, pv)
	}
	return lista, rows.Err()
}

// AggregarVendas consolida as vendas CONFIRMADAS no intervalo.
func (r *RelatorioRepository) AggregarVendas(ctx context.Context, de, ate time.Time) (*domain.ResumoVendas, error) {
	var total int
	var valor float64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int,
		       COALESCE(SUM(val_venda), 0)::float8
		FROM   vendas.venda_master
		WHERE  status_venda = 'CONFIRMADA'
		  AND  dt_venda >= $1 AND dt_venda <= $2`,
		de, ate,
	).Scan(&total, &valor)
	if err != nil {
		return nil, err
	}

	ticket := 0.0
	if total > 0 {
		ticket = valor / float64(total)
	}
	return &domain.ResumoVendas{
		TotalVendas: total,
		ValorTotal:  valor,
		TicketMedio: ticket,
		De:          de,
		Ate:         ate,
	}, nil
}

// AggregarCompras consolida as compras CONFIRMADAS no intervalo.
// dt_compra é DATE no banco; comparamos com TIMESTAMPTZ convertido para date.
func (r *RelatorioRepository) AggregarCompras(ctx context.Context, de, ate time.Time) (*domain.ResumoCompras, error) {
	var total int
	var valor float64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int,
		       COALESCE(SUM(val_compra), 0)::float8
		FROM   compras.compra_master
		WHERE  status_compra = 'CONFIRMADA'
		  AND  dt_compra >= $1::date AND dt_compra <= $2::date`,
		de, ate,
	).Scan(&total, &valor)
	if err != nil {
		return nil, err
	}
	return &domain.ResumoCompras{
		TotalCompras: total,
		ValorTotal:   valor,
		De:           de,
		Ate:          ate,
	}, nil
}
