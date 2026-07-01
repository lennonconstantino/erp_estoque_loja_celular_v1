package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/domain"
)

// CriarClienteInput é o comando de criação de cliente.
type CriarClienteInput struct {
	CPF         string
	Nome        string
	Email       string
	Telefone    string
	CEP         string
	Numero      string
	Complemento string
	// Rua/Bairro/Cidade/UF podem vir vazios e serem preenchidos via CEP.
	Rua    string
	Bairro string
	Cidade string
	UF     string
}

// AtualizarClienteInput é o comando de atualização (contato + endereço + status).
type AtualizarClienteInput struct {
	Nome        string
	Email       string
	Telefone    string
	CEP         string
	Numero      string
	Complemento string
	Rua         string
	Bairro      string
	Cidade      string
	UF          string
	Ativo       bool
}

// ClienteService é a porta de entrada (caso de uso) oferecida pelo módulo.
type ClienteService interface {
	Criar(ctx context.Context, in CriarClienteInput) (*domain.Cliente, error)
	Atualizar(ctx context.Context, id uuid.UUID, in AtualizarClienteInput) (*domain.Cliente, error)
	Remover(ctx context.Context, id uuid.UUID) error
	BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Cliente, error)
	BuscarPorCPF(ctx context.Context, cpf string) (*domain.Cliente, error)
	Listar(ctx context.Context, q string, limit, offset int) ([]domain.Cliente, error)
	ConsultarCEP(ctx context.Context, cep string) (domain.Endereco, error)
}
