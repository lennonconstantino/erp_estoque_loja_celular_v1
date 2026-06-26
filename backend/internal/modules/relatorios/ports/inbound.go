// Package ports define as interfaces do módulo relatorios.
package ports

import (
	"context"
	"time"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/domain"
)

// RelatorioService expõe os casos de uso de relatório.
type RelatorioService interface {
	// ProdutosAbaixoDoMinimo retorna produtos ativos cujo estoque atual
	// está abaixo do estoque mínimo, ordenados pela maior defasagem.
	ProdutosAbaixoDoMinimo(ctx context.Context) ([]domain.ProdutoAbaixoMinimo, error)

	// MaisVendidos retorna até `limite` produtos com maior quantidade
	// vendida em vendas CONFIRMADAS no intervalo [de, ate].
	MaisVendidos(ctx context.Context, de, ate time.Time, limite int) ([]domain.ProdutoVendido, error)

	// MenosVendidos retorna até `limite` produtos com menor quantidade
	// vendida em vendas CONFIRMADAS no intervalo [de, ate].
	MenosVendidos(ctx context.Context, de, ate time.Time, limite int) ([]domain.ProdutoVendido, error)

	// ResumoVendas agrega as vendas CONFIRMADAS no intervalo [de, ate].
	ResumoVendas(ctx context.Context, de, ate time.Time) (*domain.ResumoVendas, error)

	// ResumoCompras agrega as compras CONFIRMADAS no intervalo [de, ate].
	ResumoCompras(ctx context.Context, de, ate time.Time) (*domain.ResumoCompras, error)
}
