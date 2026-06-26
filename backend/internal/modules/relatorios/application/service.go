// Package application implementa os casos de uso do módulo relatorios.
package application

import (
	"context"
	"time"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/ports"
)

// Service implementa ports.RelatorioService delegando ao repositório.
type Service struct {
	repo ports.RelatorioRepository
}

var _ ports.RelatorioService = (*Service)(nil)

// NewService cria um Service com o repositório injetado.
func NewService(repo ports.RelatorioRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ProdutosAbaixoDoMinimo(ctx context.Context) ([]domain.ProdutoAbaixoMinimo, error) {
	return s.repo.ListarAbaixoDoMinimo(ctx)
}

func (s *Service) MaisVendidos(ctx context.Context, de, ate time.Time, limite int) ([]domain.ProdutoVendido, error) {
	if limite <= 0 {
		limite = 10
	}
	return s.repo.ListarMaisVendidos(ctx, de, ate, limite)
}

func (s *Service) MenosVendidos(ctx context.Context, de, ate time.Time, limite int) ([]domain.ProdutoVendido, error) {
	if limite <= 0 {
		limite = 10
	}
	return s.repo.ListarMenosVendidos(ctx, de, ate, limite)
}

func (s *Service) ResumoVendas(ctx context.Context, de, ate time.Time) (*domain.ResumoVendas, error) {
	return s.repo.AggregarVendas(ctx, de, ate)
}

func (s *Service) ResumoCompras(ctx context.Context, de, ate time.Time) (*domain.ResumoCompras, error) {
	return s.repo.AggregarCompras(ctx, de, ate)
}
