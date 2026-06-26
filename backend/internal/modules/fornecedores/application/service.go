// Package application contém os casos de uso do contexto fornecedores.
// Orquestra domínio e portas de saída; não conhece HTTP nem detalhes de banco.
package application

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/ports"
)

// Service implementa ports.FornecedorService.
type Service struct {
	repo ports.FornecedorRepository
	cep  ports.CepGateway
}

// NewService injeta as portas de saída.
func NewService(repo ports.FornecedorRepository, cep ports.CepGateway) *Service {
	return &Service{repo: repo, cep: cep}
}

var _ ports.FornecedorService = (*Service)(nil)

// Criar valida o CNPJ, garante unicidade, completa o endereço por CEP e persiste.
func (s *Service) Criar(ctx context.Context, in ports.CriarFornecedorInput) (*domain.Fornecedor, error) {
	f, err := domain.NovoFornecedor(in.CNPJ, in.RazaoSocial, in.NomeFantasia, in.Email, in.Telefone1, in.Comercial)
	if err != nil {
		return nil, err
	}

	if existente, err := s.repo.FindByCNPJ(ctx, f.CNPJ); err != nil {
		if !errors.Is(err, domain.ErrNaoEncontrado) {
			return nil, err
		}
	} else if existente != nil {
		return nil, domain.ErrCNPJJaCadastrado
	}

	f.Telefone2 = strings.TrimSpace(in.Telefone2)
	f.Financeiro = strings.TrimSpace(in.Financeiro)
	f.CEP = domain.NormalizarDigitos(in.CEP)
	f.Numero = in.Numero
	f.Complemento = in.Complemento
	f.Rua, f.Bairro, f.Cidade, f.UF = in.Rua, in.Bairro, in.Cidade, strings.ToUpper(in.UF)

	s.completarEndereco(ctx, f)

	if err := f.Validar(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

// Atualizar altera dados de um fornecedor existente.
func (s *Service) Atualizar(ctx context.Context, id uuid.UUID, in ports.AtualizarFornecedorInput) (*domain.Fornecedor, error) {
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := f.AtualizarDados(
		in.RazaoSocial, in.NomeFantasia, in.Email,
		in.Telefone1, in.Telefone2, in.Comercial, in.Financeiro,
	); err != nil {
		return nil, err
	}

	f.CEP = domain.NormalizarDigitos(in.CEP)
	f.Numero = in.Numero
	f.Complemento = in.Complemento
	if in.Rua != "" || in.Bairro != "" || in.Cidade != "" || in.UF != "" {
		f.Rua, f.Bairro, f.Cidade, f.UF = in.Rua, in.Bairro, in.Cidade, strings.ToUpper(in.UF)
	}
	s.completarEndereco(ctx, f)

	if in.Ativo != nil {
		f.Ativo = *in.Ativo
	}

	if err := s.repo.Update(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Fornecedor, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) BuscarPorCNPJ(ctx context.Context, cnpj string) (*domain.Fornecedor, error) {
	return s.repo.FindByCNPJ(ctx, domain.NormalizarDigitos(cnpj))
}

func (s *Service) Listar(ctx context.Context, q string, limit, offset int) ([]domain.Fornecedor, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, q, limit, offset)
}

func (s *Service) ConsultarCEP(ctx context.Context, cep string) (domain.Endereco, error) {
	return s.cep.Lookup(ctx, domain.NormalizarDigitos(cep))
}

// completarEndereco preenche rua/bairro/cidade/UF via CEP (best-effort).
func (s *Service) completarEndereco(ctx context.Context, f *domain.Fornecedor) {
	if f.CEP == "" || f.Rua != "" {
		return
	}
	if end, err := s.cep.Lookup(ctx, f.CEP); err == nil {
		f.AplicarEndereco(end)
	}
}
