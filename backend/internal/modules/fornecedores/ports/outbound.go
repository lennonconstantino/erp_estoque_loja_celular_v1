package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/domain"
)

// FornecedorRepository é a porta de saída de persistência.
type FornecedorRepository interface {
	Create(ctx context.Context, f *domain.Fornecedor) error
	Update(ctx context.Context, f *domain.Fornecedor) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Fornecedor, error)
	FindByCNPJ(ctx context.Context, cnpj string) (*domain.Fornecedor, error)
	List(ctx context.Context, q string, limit, offset int) ([]domain.Fornecedor, error)
	AtualizarUltimaCompra(ctx context.Context, id uuid.UUID, data time.Time) error
}

// CepGateway é a porta de saída para consulta de CEP.
type CepGateway interface {
	Lookup(ctx context.Context, cep string) (domain.Endereco, error)
}
