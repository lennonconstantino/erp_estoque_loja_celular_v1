// Package application contém os casos de uso do contexto vendas. Orquestra o
// domínio e as portas de saída; não conhece HTTP nem detalhes de banco.
package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/ports"
)

// Service implementa VendaService.
type Service struct {
	repo      ports.VendaRepository
	catalogo  ports.CatalogoReader
	estoque   ports.EstoqueWriter
	cliente   ports.ClienteWriter
	fiscal    ports.FiscalGateway
}

// NewService injeta as portas de saída.
func NewService(
	repo ports.VendaRepository,
	catalogo ports.CatalogoReader,
	estoque ports.EstoqueWriter,
	cliente ports.ClienteWriter,
	fiscal ports.FiscalGateway,
) *Service {
	return &Service{repo: repo, catalogo: catalogo, estoque: estoque, cliente: cliente, fiscal: fiscal}
}

var _ ports.VendaService = (*Service)(nil)

// CriarVenda valida itens, cria a venda em rascunho e persiste.
func (s *Service) CriarVenda(ctx context.Context, in ports.CriarVendaInput) (*domain.Venda, error) {
	if len(in.Itens) == 0 {
		return nil, domain.ErrVendaVazia
	}

	venda, err := domain.NovaVenda(in.ClienteID, in.ConsumidorFinal, in.FormaPgto, in.DocFiscal, in.Desconto)
	if err != nil {
		return nil, err
	}

	for _, i := range in.Itens {
		if err := s.catalogo.ExisteProduto(ctx, i.ProdutoID); err != nil {
			return nil, err
		}
		det, err := domain.NovoDetalheVenda(venda.ID, i.ProdutoID, i.Quantidade, i.PrecoUnitario)
		if err != nil {
			return nil, err
		}
		venda.Itens = append(venda.Itens, *det)
	}

	if err := venda.CalcularTotal(); err != nil {
		return nil, err
	}

	if err := s.repo.Criar(ctx, venda); err != nil {
		return nil, err
	}
	return venda, nil
}

// ConfirmarVenda valida saldo, baixa estoque, confirma a venda e emite documento fiscal.
func (s *Service) ConfirmarVenda(ctx context.Context, id, responsavelID uuid.UUID) (*domain.Venda, error) {
	venda, err := s.repo.BuscarPorID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Transição de estado no domínio (verifica status e itens).
	if err := venda.Confirmar(); err != nil {
		return nil, err
	}

	// Validação antecipada de saldo (fast-path; o DecreremntarSaldo é o guarda atômico).
	for _, item := range venda.Itens {
		saldo, err := s.catalogo.ConsultarSaldoProduto(ctx, item.ProdutoID)
		if err != nil {
			return nil, err
		}
		if saldo < item.Quantidade {
			return nil, domain.ErrSaldoInsuficiente
		}
	}

	// Baixa de estoque (atômica por produto — guard contra corrida de confirmações).
	for _, item := range venda.Itens {
		if err := s.estoque.RegistrarSaidaVenda(ctx, item.ProdutoID, venda.ID, responsavelID, item.Quantidade); err != nil {
			if err.Error() == domain.ErrSaldoInsuficiente.Error() {
				return nil, domain.ErrSaldoInsuficiente
			}
			return nil, err
		}
	}

	// Persiste o novo status.
	if err := s.repo.AtualizarStatus(ctx, venda.ID, venda.Status); err != nil {
		return nil, err
	}

	// Emissão de documento fiscal (best-effort — falha não reverte a venda).
	if numDoc, err := s.fiscal.EmitirDocumento(ctx, venda); err == nil {
		venda.DocFiscalNumero = numDoc
	}

	// Atualiza última venda do cliente (best-effort).
	if venda.ClienteID != nil {
		_ = s.cliente.AtualizarUltimaVenda(ctx, *venda.ClienteID, time.Now().UTC())
	}

	return venda, nil
}

// ListarVendas retorna as vendas com paginação, opcionalmente filtradas por cliente.
func (s *Service) ListarVendas(ctx context.Context, clienteID *uuid.UUID, limit, offset int) ([]domain.Venda, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.Listar(ctx, clienteID, limit, offset)
}

// BuscarVenda retorna uma venda completa pelo ID.
func (s *Service) BuscarVenda(ctx context.Context, id uuid.UUID) (*domain.Venda, error) {
	return s.repo.BuscarPorID(ctx, id)
}
