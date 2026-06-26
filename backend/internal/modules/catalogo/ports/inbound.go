package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/domain"
)

// ─── Inputs ──────────────────────────────────────────────────────────────────

type CriarCategoriaInput struct{ Descricao string }

type AtualizarCategoriaInput struct{ Descricao string }

type CriarProdutoInput struct {
	CategoriaID   uuid.UUID
	Descricao     string
	PrecoCusto    float64
	PrecoVenda    float64
	EstoqueMinimo int
	GarantiaMeses int
	Modelo        string
}

type AtualizarProdutoInput struct {
	Descricao     string
	CategoriaID   uuid.UUID
	PrecoCusto    float64
	PrecoVenda    float64
	EstoqueMinimo int
	GarantiaMeses int
	Modelo        string
	Ativo         *bool
}

// ─── Portas inbound (o que o módulo oferece) ─────────────────────────────────

// CategoriaService é a porta de entrada para operações sobre categorias.
type CategoriaService interface {
	CriarCategoria(ctx context.Context, in CriarCategoriaInput) (*domain.Categoria, error)
	AtualizarCategoria(ctx context.Context, id uuid.UUID, in AtualizarCategoriaInput) (*domain.Categoria, error)
	RemoverCategoria(ctx context.Context, id uuid.UUID) error
	BuscarCategoriaPorID(ctx context.Context, id uuid.UUID) (*domain.Categoria, error)
	ListarCategorias(ctx context.Context, q string, limit, offset int) ([]domain.Categoria, error)
}

// ProdutoService é a porta de entrada para operações sobre produtos.
type ProdutoService interface {
	CriarProduto(ctx context.Context, in CriarProdutoInput) (*domain.Produto, error)
	AtualizarProduto(ctx context.Context, id uuid.UUID, in AtualizarProdutoInput) (*domain.Produto, error)
	BuscarProdutoPorID(ctx context.Context, id uuid.UUID) (*domain.Produto, error)
	ListarProdutos(ctx context.Context, q string, categoriaID *uuid.UUID, limit, offset int) ([]domain.Produto, error)
}

// CatalogoReader é a interface exposta para que compras/vendas possam consultar
// produtos (consumida como porta outbound por outros módulos).
type CatalogoReader interface {
	BuscarProduto(ctx context.Context, id uuid.UUID) (*domain.Produto, error)
	// ExisteProduto retorna nil se o produto existir ou erro de domínio caso contrário.
	ExisteProduto(ctx context.Context, id uuid.UUID) error
	// ConsultarSaldoProduto retorna o saldo materializado atual de um produto.
	ConsultarSaldoProduto(ctx context.Context, id uuid.UUID) (int, error)
}

// CatalogoWriter é a interface exposta para que estoque atualize saldo e
// disponibilidade de um produto (consumida como porta outbound por estoque).
type CatalogoWriter interface {
	AtualizarSaldo(ctx context.Context, produtoID uuid.UUID, novoSaldo int) error
	// DecrementarSaldo atomicamente decrementa o saldo se houver estoque suficiente.
	DecrementarSaldo(ctx context.Context, produtoID uuid.UUID, quantidade int) (novoSaldo int, err error)
}
