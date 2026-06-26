package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/domain"
)

// VendaRepository é a porta de saída de persistência de vendas.
type VendaRepository interface {
	Criar(ctx context.Context, v *domain.Venda) error
	BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Venda, error)
	Listar(ctx context.Context, clienteID *uuid.UUID, limit, offset int) ([]domain.Venda, error)
	AtualizarStatus(ctx context.Context, id uuid.UUID, status domain.StatusVenda) error
}

// CatalogoReader é a porta de saída para consultar produtos.
// Satisfeita pela interface catalogo.CatalogoReader via duck typing.
type CatalogoReader interface {
	ExisteProduto(ctx context.Context, id uuid.UUID) error
	ConsultarSaldoProduto(ctx context.Context, id uuid.UUID) (int, error)
}

// EstoqueWriter é a porta de saída para registrar saídas no razão de estoque.
// Satisfeita por estoque.EstoqueWriter via duck typing.
type EstoqueWriter interface {
	RegistrarSaidaVenda(ctx context.Context, produtoID, vendaID, responsavelID uuid.UUID, quantidade int) error
}

// ClienteWriter é a porta de saída para atualizar metadados de cliente.
// Satisfeita por clientes.Service via duck typing.
type ClienteWriter interface {
	AtualizarUltimaVenda(ctx context.Context, clienteID uuid.UUID, data time.Time) error
}

// FiscalGateway é a porta de saída para emissão de documentos fiscais.
type FiscalGateway interface {
	EmitirDocumento(ctx context.Context, v *domain.Venda) (numDoc string, err error)
}
