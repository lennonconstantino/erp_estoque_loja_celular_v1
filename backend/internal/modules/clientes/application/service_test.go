package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/ports"
)

// ─── Mocks ───────────────────────────────────────────────────────────────────

type mockRepo struct {
	clientes  map[uuid.UUID]*domain.Cliente
	atualizado *domain.Cliente
}

func newMockRepo() *mockRepo { return &mockRepo{clientes: make(map[uuid.UUID]*domain.Cliente)} }

func (m *mockRepo) Create(_ context.Context, c *domain.Cliente) error {
	m.clientes[c.ID] = c
	return nil
}

func (m *mockRepo) Update(_ context.Context, c *domain.Cliente) error {
	copia := *c
	m.atualizado = &copia
	m.clientes[c.ID] = c
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.clientes, id)
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Cliente, error) {
	c, ok := m.clientes[id]
	if !ok {
		return nil, domain.ErrNaoEncontrado
	}
	copia := *c
	return &copia, nil
}

func (m *mockRepo) FindByCPF(_ context.Context, cpf string) (*domain.Cliente, error) {
	for _, c := range m.clientes {
		if c.CPF == cpf {
			copia := *c
			return &copia, nil
		}
	}
	return nil, domain.ErrNaoEncontrado
}

func (m *mockRepo) List(_ context.Context, _ string, _, _ int) ([]domain.Cliente, error) {
	return nil, nil
}

func (m *mockRepo) AtualizarUltimaVenda(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return nil
}

// mockCep nunca é acionado nos casos abaixo (todos já têm rua preenchida ou CEP
// vazio), então retorna endereço vazio sem erro.
type mockCep struct{}

func (mockCep) Lookup(_ context.Context, _ string) (domain.Endereco, error) {
	return domain.Endereco{}, nil
}

// helper: cria e persiste um cliente ativo com endereço já preenchido (para não
// disparar a consulta de CEP na atualização).
func semearCliente(t *testing.T, repo *mockRepo) *domain.Cliente {
	t.Helper()
	c, err := domain.NovoCliente("529.982.247-25", "Maria", "maria@x.com")
	if err != nil {
		t.Fatalf("setup: NovoCliente retornou erro: %v", err)
	}
	c.Rua = "Rua A" // evita completarEndereco
	repo.clientes[c.ID] = c
	return c
}

func inputBase() ports.AtualizarClienteInput {
	return ports.AtualizarClienteInput{
		Nome: "Maria", Email: "maria@x.com", Rua: "Rua A",
	}
}

// ─── Testes ──────────────────────────────────────────────────────────────────

func TestAtualizar_InativaClienteQuandoAtivoFalse(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewService(repo, mockCep{})
	c := semearCliente(t, repo)
	if !c.Ativo {
		t.Fatalf("pré-condição: cliente deveria estar ativo")
	}

	in := inputBase()
	in.Ativo = false
	out, err := svc.Atualizar(context.Background(), c.ID, in)
	if err != nil {
		t.Fatalf("Atualizar retornou erro: %v", err)
	}
	if out.Ativo {
		t.Fatalf("esperava cliente inativo após Atualizar com Ativo=false")
	}
	if repo.atualizado == nil || repo.atualizado.Ativo {
		t.Fatalf("estado persistido deveria estar inativo, veio %+v", repo.atualizado)
	}
}

func TestAtualizar_ReativaClienteQuandoAtivoTrue(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewService(repo, mockCep{})
	c := semearCliente(t, repo)
	c.Ativo = false // parte de inativo

	in := inputBase()
	in.Ativo = true
	out, err := svc.Atualizar(context.Background(), c.ID, in)
	if err != nil {
		t.Fatalf("Atualizar retornou erro: %v", err)
	}
	if !out.Ativo {
		t.Fatalf("esperava cliente ativo após Atualizar com Ativo=true")
	}
	if repo.atualizado == nil || !repo.atualizado.Ativo {
		t.Fatalf("estado persistido deveria estar ativo, veio %+v", repo.atualizado)
	}
}

func TestCriar_NasceAtivo(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewService(repo, mockCep{})

	out, err := svc.Criar(context.Background(), ports.CriarClienteInput{
		CPF: "529.982.247-25", Nome: "Maria", Email: "maria@x.com", Rua: "Rua A",
	})
	if err != nil {
		t.Fatalf("Criar retornou erro: %v", err)
	}
	if !out.Ativo {
		t.Fatalf("cliente recém-criado deveria nascer ativo")
	}
}
