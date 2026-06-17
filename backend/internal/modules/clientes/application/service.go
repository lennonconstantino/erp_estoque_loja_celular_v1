// Package application contém os casos de uso do contexto clientes. Orquestra o
// domínio e as portas de saída; não conhece HTTP nem detalhes de banco.
package application

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/ports"
)

// Service implementa ports.ClienteService.
type Service struct {
	repo ports.ClienteRepository
	cep  ports.CepGateway
}

// NewService injeta as portas de saída.
func NewService(repo ports.ClienteRepository, cep ports.CepGateway) *Service {
	return &Service{repo: repo, cep: cep}
}

// garante em tempo de compilação que Service satisfaz a porta de entrada.
var _ ports.ClienteService = (*Service)(nil)

// Criar valida o CPF, garante unicidade, completa o endereço por CEP e persiste.
func (s *Service) Criar(ctx context.Context, in ports.CriarClienteInput) (*domain.Cliente, error) {
	c, err := domain.NovoCliente(in.CPF, in.Nome, in.Email)
	if err != nil {
		return nil, err
	}

	// unicidade do CPF
	if existente, err := s.repo.FindByCPF(ctx, c.CPF); err != nil {
		if !errors.Is(err, domain.ErrNaoEncontrado) {
			return nil, err
		}
	} else if existente != nil {
		return nil, domain.ErrCPFJaCadastrado
	}

	c.Telefone = strings.TrimSpace(in.Telefone)
	c.CEP = domain.NormalizarCPF(in.CEP) // só dígitos
	c.Numero = in.Numero
	c.Complemento = in.Complemento
	c.Rua, c.Bairro, c.Cidade, c.UF = in.Rua, in.Bairro, in.Cidade, strings.ToUpper(in.UF)

	s.completarEndereco(ctx, c)

	if err := c.Validar(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Atualizar altera contato e endereço de um cliente existente.
func (s *Service) Atualizar(ctx context.Context, id uuid.UUID, in ports.AtualizarClienteInput) (*domain.Cliente, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := c.AtualizarContato(in.Nome, in.Email, in.Telefone); err != nil {
		return nil, err
	}
	c.CEP = domain.NormalizarCPF(in.CEP)
	c.Numero = in.Numero
	c.Complemento = in.Complemento
	if in.Rua != "" || in.Bairro != "" || in.Cidade != "" || in.UF != "" {
		c.Rua, c.Bairro, c.Cidade, c.UF = in.Rua, in.Bairro, in.Cidade, strings.ToUpper(in.UF)
	}
	s.completarEndereco(ctx, c)

	if err := c.Validar(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Remover(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Cliente, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) BuscarPorCPF(ctx context.Context, cpf string) (*domain.Cliente, error) {
	return s.repo.FindByCPF(ctx, domain.NormalizarCPF(cpf))
}

func (s *Service) Listar(ctx context.Context, q string, limit, offset int) ([]domain.Cliente, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, q, limit, offset)
}

func (s *Service) ConsultarCEP(ctx context.Context, cep string) (domain.Endereco, error) {
	return s.cep.Lookup(ctx, domain.NormalizarCPF(cep))
}

// completarEndereco preenche rua/bairro/cidade/UF via CEP quando há CEP e a rua
// está vazia. Falha de CEP não impede o cadastro (consulta é best-effort).
func (s *Service) completarEndereco(ctx context.Context, c *domain.Cliente) {
	if c.CEP == "" || c.Rua != "" {
		return
	}
	if end, err := s.cep.Lookup(ctx, c.CEP); err == nil {
		c.AplicarEndereco(end)
	}
}
