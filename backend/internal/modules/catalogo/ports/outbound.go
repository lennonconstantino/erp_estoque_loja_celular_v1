package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/domain"
)

// CategoriaRepository é a porta de saída de persistência de categorias.
type CategoriaRepository interface {
	Create(ctx context.Context, c *domain.Categoria) error
	Update(ctx context.Context, c *domain.Categoria) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Categoria, error)
	FindByDescricao(ctx context.Context, descricao string) (*domain.Categoria, error)
	List(ctx context.Context, q string, limit, offset int) ([]domain.Categoria, error)
	HasProdutos(ctx context.Context, id uuid.UUID) (bool, error)
}

// ProdutoRepository é a porta de saída de persistência de produtos.
type ProdutoRepository interface {
	Create(ctx context.Context, p *domain.Produto) error
	Update(ctx context.Context, p *domain.Produto) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Produto, error)
	List(ctx context.Context, q string, categoriaID *uuid.UUID, limit, offset int) ([]domain.Produto, error)
	UpdateSaldo(ctx context.Context, id uuid.UUID, novoSaldo int, disponivel bool) error
	// DecrementarSaldo atomicamente decrementa o saldo se estoque_a_pro >= quantidade.
	DecrementarSaldo(ctx context.Context, id uuid.UUID, quantidade int) (novoSaldo int, err error)
}
