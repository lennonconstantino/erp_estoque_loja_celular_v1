package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/domain"
)

// CompraRepository é a porta de saída de persistência de compras.
type CompraRepository interface {
	Criar(ctx context.Context, c *domain.Compra) error
	BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Compra, error)
	Listar(ctx context.Context, fornecedorID *uuid.UUID, limit, offset int) ([]domain.Compra, error)
	AtualizarStatus(ctx context.Context, id uuid.UUID, status domain.StatusCompra) error
}

// CatalogoReader verifica existência de produtos no catálogo.
// Implementada pelo módulo catálogo (duck typing — sem importar catalogo/domain).
type CatalogoReader interface {
	ExisteProduto(ctx context.Context, id uuid.UUID) error
}

// EstoqueWriter lança movimentações de entrada de estoque.
// Implementada pelo módulo estoque (duck typing).
type EstoqueWriter interface {
	RegistrarEntradaCompra(ctx context.Context, produtoID, compraID, responsavelID uuid.UUID, quantidade int) error
}

// FornecedorWriter atualiza metadados do fornecedor após confirmação da compra.
// Implementada pelo módulo fornecedores (duck typing).
type FornecedorWriter interface {
	AtualizarUltimaCompra(ctx context.Context, fornecedorID uuid.UUID, data time.Time) error
}
