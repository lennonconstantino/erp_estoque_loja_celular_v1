// Package ports declara os contratos (interfaces) do contexto fornecedores.
package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/domain"
)

// CriarFornecedorInput é o comando de criação.
type CriarFornecedorInput struct {
	CNPJ         string
	RazaoSocial  string
	NomeFantasia string
	Email        string
	Telefone1    string
	Telefone2    string
	CEP          string
	Numero       string
	Complemento  string
	Rua          string
	Bairro       string
	Cidade       string
	UF           string
	Comercial    string
	Financeiro   string
}

// AtualizarFornecedorInput é o comando de atualização.
type AtualizarFornecedorInput struct {
	RazaoSocial  string
	NomeFantasia string
	Email        string
	Telefone1    string
	Telefone2    string
	CEP          string
	Numero       string
	Complemento  string
	Rua          string
	Bairro       string
	Cidade       string
	UF           string
	Comercial    string
	Financeiro   string
	Ativo        *bool
}

// FornecedorService é a porta de entrada oferecida pelo módulo.
type FornecedorService interface {
	Criar(ctx context.Context, in CriarFornecedorInput) (*domain.Fornecedor, error)
	Atualizar(ctx context.Context, id uuid.UUID, in AtualizarFornecedorInput) (*domain.Fornecedor, error)
	BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Fornecedor, error)
	BuscarPorCNPJ(ctx context.Context, cnpj string) (*domain.Fornecedor, error)
	Listar(ctx context.Context, q string, limit, offset int) ([]domain.Fornecedor, error)
	ConsultarCEP(ctx context.Context, cep string) (domain.Endereco, error)
}
