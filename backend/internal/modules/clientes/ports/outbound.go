// Package ports declara os contratos (interfaces) do contexto clientes:
// o que o módulo OFERECE (inbound) e o que ele EXIGE (outbound).
package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/domain"
)

// ClienteRepository é a porta de saída de persistência.
type ClienteRepository interface {
	Create(ctx context.Context, c *domain.Cliente) error
	Update(ctx context.Context, c *domain.Cliente) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Cliente, error)
	FindByCPF(ctx context.Context, cpf string) (*domain.Cliente, error)
	List(ctx context.Context, q string, limit, offset int) ([]domain.Cliente, error)
}

// CepGateway é a porta de saída para consulta de CEP (dependência externa).
// O adaptador que a implementa aplica as políticas de resiliência.
type CepGateway interface {
	Lookup(ctx context.Context, cep string) (domain.Endereco, error)
}
