// Package ports declara os contratos do contexto compras.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/domain"
)

// ─── Inputs ──────────────────────────────────────────────────────────────────

// CriarDetalheInput representa um item dentro do comando de criação de compra.
type CriarDetalheInput struct {
	ProdutoID   uuid.UUID
	Quantidade  int
	PrecoCompra float64
	PrecoVenda  float64
}

// CriarCompraInput é o comando de criação de uma compra em rascunho.
type CriarCompraInput struct {
	FornecedorID uuid.UUID
	NF           string
	DtCompra     time.Time
	Itens        []CriarDetalheInput
}

// ─── Porta inbound ───────────────────────────────────────────────────────────

// CompraService é a porta de entrada do módulo compras.
type CompraService interface {
	// CriarCompra cria uma compra em status RASCUNHO com os itens informados.
	CriarCompra(ctx context.Context, in CriarCompraInput) (*domain.Compra, error)

	// ConfirmarCompra transiciona para CONFIRMADA, lança movimentações de estoque
	// e atualiza a data de última compra do fornecedor.
	ConfirmarCompra(ctx context.Context, id, responsavelID uuid.UUID) (*domain.Compra, error)

	// ListarCompras retorna o cabeçalho das compras, opcionalmente filtrado por fornecedor.
	ListarCompras(ctx context.Context, fornecedorID *uuid.UUID, limit, offset int) ([]domain.Compra, error)

	// BuscarCompra retorna uma compra completa (com itens).
	BuscarCompra(ctx context.Context, id uuid.UUID) (*domain.Compra, error)
}
