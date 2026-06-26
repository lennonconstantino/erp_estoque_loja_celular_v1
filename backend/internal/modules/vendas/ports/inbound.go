package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/domain"
)

// CriarDetalheInput dados de um item da venda.
type CriarDetalheInput struct {
	ProdutoID     uuid.UUID
	Quantidade    int
	PrecoUnitario float64
}

// CriarVendaInput comando de criação de uma venda em rascunho.
type CriarVendaInput struct {
	ClienteID       *uuid.UUID
	ConsumidorFinal bool
	FormaPgto       domain.FormaPgto
	DocFiscal       domain.DocFiscal
	Desconto        float64
	Itens           []CriarDetalheInput
}

// VendaService é a porta de entrada (caso de uso) do módulo vendas.
type VendaService interface {
	CriarVenda(ctx context.Context, in CriarVendaInput) (*domain.Venda, error)
	ConfirmarVenda(ctx context.Context, id, responsavelID uuid.UUID) (*domain.Venda, error)
	ListarVendas(ctx context.Context, clienteID *uuid.UUID, limit, offset int) ([]domain.Venda, error)
	BuscarVenda(ctx context.Context, id uuid.UUID) (*domain.Venda, error)
}
