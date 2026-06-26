// Package application contém os casos de uso do contexto estoque. Orquestra
// o domínio e as portas de saída; não conhece HTTP nem detalhes de banco.
package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/ports"
)

// Service implementa EstoqueService.
type Service struct {
	movs    ports.MovimentacaoRepository
	ajustes ports.AjusteRepository
	catalogo ports.CatalogoWriter
}

// NewService injeta as portas de saída.
func NewService(movs ports.MovimentacaoRepository, ajustes ports.AjusteRepository, catalogo ports.CatalogoWriter) *Service {
	return &Service{movs: movs, ajustes: ajustes, catalogo: catalogo}
}

// Conformidade em tempo de compilação.
var _ ports.EstoqueService = (*Service)(nil)
var _ ports.EstoqueWriter  = (*Service)(nil)

// ─── EstoqueService ──────────────────────────────────────────────────────────

// LancarAjuste valida, persiste o ajuste, gera movimentações no razão e
// atualiza o saldo materializado no catálogo.
func (s *Service) LancarAjuste(ctx context.Context, in ports.LancarAjusteInput) (*domain.Ajuste, error) {
	// 1. Valida domínio.
	ajs, err := domain.NovoAjuste(in.ProdutoID, in.ResponsavelID, in.QtdEntrada, in.QtdSaida, in.Motivo)
	if err != nil {
		return nil, err
	}

	// 2. Saldo atual vem do próprio razão (ledger é a fonte da verdade).
	saldoAtual, err := s.movs.ConsultarSaldoAtual(ctx, in.ProdutoID)
	if err != nil {
		return nil, err
	}

	// 3. Calcula novo saldo e valida suficiência para saída.
	novoSaldo := saldoAtual + ajs.QtdEntrada - ajs.QtdSaida
	if novoSaldo < 0 {
		return nil, domain.ErrSaldoInsuficiente
	}

	// 4. Persiste o documento de ajuste.
	if err := s.ajustes.Create(ctx, ajs); err != nil {
		return nil, err
	}

	// 5. Lança movimentações no razão.
	// Entrada: saldo aumenta.
	cursor := saldoAtual
	if ajs.QtdEntrada > 0 {
		mov := domain.NovaMovimentacao(
			ajs.ProdutoID, domain.TipoAjusteEntrada, ajs.QtdEntrada,
			cursor, cursor+ajs.QtdEntrada,
			"AJUSTE", &ajs.ID, &ajs.Responsavel,
		)
		if err := s.movs.Create(ctx, mov); err != nil {
			return nil, err
		}
		cursor += ajs.QtdEntrada
	}
	// Saída: saldo diminui.
	if ajs.QtdSaida > 0 {
		mov := domain.NovaMovimentacao(
			ajs.ProdutoID, domain.TipoAjusteSaida, ajs.QtdSaida,
			cursor, cursor-ajs.QtdSaida,
			"AJUSTE", &ajs.ID, &ajs.Responsavel,
		)
		if err := s.movs.Create(ctx, mov); err != nil {
			return nil, err
		}
	}

	// 6. Atualiza saldo materializado no catálogo.
	if err := s.catalogo.AtualizarSaldo(ctx, ajs.ProdutoID, novoSaldo); err != nil {
		return nil, err
	}

	return ajs, nil
}

// ConsultarMovimentacoes retorna o razão de um produto (mais recente → mais antigo).
func (s *Service) ConsultarMovimentacoes(ctx context.Context, produtoID uuid.UUID, limit, offset int) ([]domain.Movimentacao, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.movs.ListByProduto(ctx, produtoID, limit, offset)
}

// ConsultarAjustes retorna os ajustes manuais de um produto (mais recente → mais antigo).
func (s *Service) ConsultarAjustes(ctx context.Context, produtoID uuid.UUID, limit, offset int) ([]domain.Ajuste, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.ajustes.ListByProduto(ctx, produtoID, limit, offset)
}

// ─── EstoqueWriter (cross-module) ────────────────────────────────────────────

// RegistrarEntradaCompra lança uma movimentação do tipo COMPRA no razão e
// atualiza o saldo materializado no catálogo. Chamado pelo módulo compras.
func (s *Service) RegistrarEntradaCompra(ctx context.Context, produtoID, compraID, responsavelID uuid.UUID, quantidade int) error {
	saldoAtual, err := s.movs.ConsultarSaldoAtual(ctx, produtoID)
	if err != nil {
		return err
	}
	novoSaldo := saldoAtual + quantidade
	mov := domain.NovaMovimentacao(
		produtoID, domain.TipoCompra, quantidade,
		saldoAtual, novoSaldo,
		"COMPRA", &compraID, &responsavelID,
	)
	if err := s.movs.Create(ctx, mov); err != nil {
		return err
	}
	return s.catalogo.AtualizarSaldo(ctx, produtoID, novoSaldo)
}
