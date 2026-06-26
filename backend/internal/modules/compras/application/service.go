// Package application contém os casos de uso do contexto compras.
// Orquestra o domínio e as portas de saída; não conhece HTTP nem SQL.
package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/ports"
)

// Service implementa ports.CompraService.
type Service struct {
	repo       ports.CompraRepository
	catalogo   ports.CatalogoReader
	estoque    ports.EstoqueWriter
	fornecedor ports.FornecedorWriter
}

// NewService injeta as portas de saída.
func NewService(
	repo ports.CompraRepository,
	catalogo ports.CatalogoReader,
	estoque ports.EstoqueWriter,
	fornecedor ports.FornecedorWriter,
) *Service {
	return &Service{repo: repo, catalogo: catalogo, estoque: estoque, fornecedor: fornecedor}
}

// Conformidade em tempo de compilação.
var _ ports.CompraService = (*Service)(nil)

// ─── CompraService ───────────────────────────────────────────────────────────

// CriarCompra valida os itens, calcula o total e persiste a compra em RASCUNHO.
func (s *Service) CriarCompra(ctx context.Context, in ports.CriarCompraInput) (*domain.Compra, error) {
	if len(in.Itens) == 0 {
		return nil, domain.ErrCompraVazia
	}

	c, err := domain.NovaCompra(in.FornecedorID, in.NF, in.DtCompra)
	if err != nil {
		return nil, err
	}

	for _, item := range in.Itens {
		// Verifica existência do produto no catálogo.
		if err := s.catalogo.ExisteProduto(ctx, item.ProdutoID); err != nil {
			return nil, err
		}
		det, err := domain.NovoDetalheCompra(c.ID, item.ProdutoID, item.Quantidade, item.PrecoCompra, item.PrecoVenda)
		if err != nil {
			return nil, err
		}
		c.Itens = append(c.Itens, *det)
	}

	c.ValorTotal = c.CalcularTotal()

	if err := s.repo.Criar(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// ConfirmarCompra transiciona a compra para CONFIRMADA, lança movimentações de
// estoque e atualiza a data de última compra do fornecedor.
func (s *Service) ConfirmarCompra(ctx context.Context, id, responsavelID uuid.UUID) (*domain.Compra, error) {
	c, err := s.repo.BuscarPorID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := c.Confirmar(); err != nil {
		return nil, err
	}

	// Lança movimentação de COMPRA no razão de cada item.
	for _, item := range c.Itens {
		if err := s.estoque.RegistrarEntradaCompra(ctx, item.ProdutoID, c.ID, responsavelID, item.Quantidade); err != nil {
			return nil, err
		}
	}

	// Persiste a transição de status.
	if err := s.repo.AtualizarStatus(ctx, c.ID, c.Status); err != nil {
		return nil, err
	}

	// Atualiza dt_ult_comp_for no fornecedor (best-effort: não reverte a compra em caso de falha).
	_ = s.fornecedor.AtualizarUltimaCompra(ctx, c.FornecedorID, time.Now().UTC())

	return c, nil
}

// ListarCompras retorna os cabeçalhos das compras, opcionalmente filtrado por fornecedor.
func (s *Service) ListarCompras(ctx context.Context, fornecedorID *uuid.UUID, limit, offset int) ([]domain.Compra, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.Listar(ctx, fornecedorID, limit, offset)
}

// BuscarCompra retorna uma compra completa (com itens).
func (s *Service) BuscarCompra(ctx context.Context, id uuid.UUID) (*domain.Compra, error) {
	return s.repo.BuscarPorID(ctx, id)
}
