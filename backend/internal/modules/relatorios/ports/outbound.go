package ports

import (
	"context"
	"time"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/domain"
)

// RelatorioRepository define as consultas que o módulo exige do banco.
type RelatorioRepository interface {
	ListarAbaixoDoMinimo(ctx context.Context) ([]domain.ProdutoAbaixoMinimo, error)
	ListarMaisVendidos(ctx context.Context, de, ate time.Time, limite int) ([]domain.ProdutoVendido, error)
	ListarMenosVendidos(ctx context.Context, de, ate time.Time, limite int) ([]domain.ProdutoVendido, error)
	AggregarVendas(ctx context.Context, de, ate time.Time) (*domain.ResumoVendas, error)
	AggregarCompras(ctx context.Context, de, ate time.Time) (*domain.ResumoCompras, error)
}
