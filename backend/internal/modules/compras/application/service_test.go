package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/ports"
)

// ─── Mocks ───────────────────────────────────────────────────────────────────

type mockRepo struct {
	compras map[uuid.UUID]*domain.Compra
	criada  *domain.Compra
	errCriar error
	errBuscar error
	errStatus error
}

func newMockRepo() *mockRepo { return &mockRepo{compras: make(map[uuid.UUID]*domain.Compra)} }

func (m *mockRepo) Criar(_ context.Context, c *domain.Compra) error {
	if m.errCriar != nil {
		return m.errCriar
	}
	m.compras[c.ID] = c
	m.criada = c
	return nil
}

func (m *mockRepo) BuscarPorID(_ context.Context, id uuid.UUID) (*domain.Compra, error) {
	if m.errBuscar != nil {
		return nil, m.errBuscar
	}
	c, ok := m.compras[id]
	if !ok {
		return nil, domain.ErrCompraNaoEncontrada
	}
	copia := *c
	return &copia, nil
}

func (m *mockRepo) Listar(_ context.Context, _ *uuid.UUID, _, _ int) ([]domain.Compra, error) {
	var out []domain.Compra
	for _, c := range m.compras {
		out = append(out, *c)
	}
	return out, nil
}

func (m *mockRepo) AtualizarStatus(_ context.Context, id uuid.UUID, status domain.StatusCompra) error {
	if m.errStatus != nil {
		return m.errStatus
	}
	if c, ok := m.compras[id]; ok {
		c.Status = status
	}
	return nil
}

type mockCatalogo struct {
	errExiste error
}

func (m *mockCatalogo) ExisteProduto(_ context.Context, _ uuid.UUID) error { return m.errExiste }

type mockEstoque struct {
	chamadas int
	err      error
}

func (m *mockEstoque) RegistrarEntradaCompra(_ context.Context, _, _, _ uuid.UUID, _ int) error {
	m.chamadas++
	return m.err
}

type mockFornecedor struct {
	chamadas int
	err      error
}

func (m *mockFornecedor) AtualizarUltimaCompra(_ context.Context, _ uuid.UUID, _ time.Time) error {
	m.chamadas++
	return m.err
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func novoServico(repo *mockRepo, cat *mockCatalogo, est *mockEstoque, forn *mockFornecedor) *application.Service {
	return application.NewService(repo, cat, est, forn)
}

func inputValido() ports.CriarCompraInput {
	return ports.CriarCompraInput{
		FornecedorID: uuid.New(),
		NF:           "NF-123",
		DtCompra:     time.Now(),
		Itens: []ports.CriarDetalheInput{
			{ProdutoID: uuid.New(), Quantidade: 2, PrecoCompra: 10.0, PrecoVenda: 25.0},
		},
	}
}

// ─── CriarCompra ─────────────────────────────────────────────────────────────

func TestCriarCompra_Sucesso(t *testing.T) {
	repo := newMockRepo()
	svc := novoServico(repo, &mockCatalogo{}, &mockEstoque{}, &mockFornecedor{})

	c, err := svc.CriarCompra(context.Background(), inputValido())
	if err != nil {
		t.Fatalf("esperava sucesso, got: %v", err)
	}
	if c.Status != domain.StatusRascunho {
		t.Errorf("status deve ser RASCUNHO, got: %s", c.Status)
	}
	if len(c.Itens) != 1 {
		t.Errorf("esperava 1 item, got: %d", len(c.Itens))
	}
	if c.ValorTotal != 20.0 { // 2 × 10.0
		t.Errorf("valor total esperado 20.0, got: %f", c.ValorTotal)
	}
}

func TestCriarCompra_SemItens(t *testing.T) {
	svc := novoServico(newMockRepo(), &mockCatalogo{}, &mockEstoque{}, &mockFornecedor{})
	in := inputValido()
	in.Itens = nil
	_, err := svc.CriarCompra(context.Background(), in)
	if err != domain.ErrCompraVazia {
		t.Fatalf("esperava ErrCompraVazia, got: %v", err)
	}
}

func TestCriarCompra_SemFornecedor(t *testing.T) {
	svc := novoServico(newMockRepo(), &mockCatalogo{}, &mockEstoque{}, &mockFornecedor{})
	in := inputValido()
	in.FornecedorID = uuid.Nil
	_, err := svc.CriarCompra(context.Background(), in)
	if err != domain.ErrFornecedorObrigatorio {
		t.Fatalf("esperava ErrFornecedorObrigatorio, got: %v", err)
	}
}

func TestCriarCompra_ProdutoNaoEncontrado(t *testing.T) {
	errProd := errors.New("produto não encontrado")
	svc := novoServico(newMockRepo(), &mockCatalogo{errExiste: errProd}, &mockEstoque{}, &mockFornecedor{})
	_, err := svc.CriarCompra(context.Background(), inputValido())
	if err != errProd {
		t.Fatalf("esperava erro de produto, got: %v", err)
	}
}

func TestCriarCompra_ErroNaoPersistencia(t *testing.T) {
	repo := newMockRepo()
	repo.errCriar = errors.New("erro DB")
	svc := novoServico(repo, &mockCatalogo{}, &mockEstoque{}, &mockFornecedor{})
	_, err := svc.CriarCompra(context.Background(), inputValido())
	if err == nil {
		t.Fatal("esperava erro de persistência")
	}
}

func TestCriarCompra_ItemInvalido(t *testing.T) {
	svc := novoServico(newMockRepo(), &mockCatalogo{}, &mockEstoque{}, &mockFornecedor{})
	in := inputValido()
	in.Itens[0].Quantidade = 0
	_, err := svc.CriarCompra(context.Background(), in)
	if err != domain.ErrQuantidadePositiva {
		t.Fatalf("esperava ErrQuantidadePositiva, got: %v", err)
	}
}

// ─── ConfirmarCompra ─────────────────────────────────────────────────────────

func criarCompraNoRepo(t *testing.T, svc *application.Service, repo *mockRepo) uuid.UUID {
	t.Helper()
	c, err := svc.CriarCompra(context.Background(), inputValido())
	if err != nil {
		t.Fatalf("falha ao criar compra: %v", err)
	}
	return c.ID
}

func TestConfirmarCompra_Sucesso(t *testing.T) {
	repo := newMockRepo()
	estoque := &mockEstoque{}
	forn := &mockFornecedor{}
	svc := novoServico(repo, &mockCatalogo{}, estoque, forn)

	id := criarCompraNoRepo(t, svc, repo)
	c, err := svc.ConfirmarCompra(context.Background(), id, uuid.New())
	if err != nil {
		t.Fatalf("esperava sucesso, got: %v", err)
	}
	if c.Status != domain.StatusConfirmada {
		t.Errorf("status deve ser CONFIRMADA, got: %s", c.Status)
	}
	if estoque.chamadas != 1 {
		t.Errorf("estoque deve ter sido chamado 1 vez, got: %d", estoque.chamadas)
	}
	if forn.chamadas != 1 {
		t.Errorf("fornecedor deve ter sido chamado 1 vez, got: %d", forn.chamadas)
	}
}

func TestConfirmarCompra_NaoEncontrada(t *testing.T) {
	svc := novoServico(newMockRepo(), &mockCatalogo{}, &mockEstoque{}, &mockFornecedor{})
	_, err := svc.ConfirmarCompra(context.Background(), uuid.New(), uuid.New())
	if err != domain.ErrCompraNaoEncontrada {
		t.Fatalf("esperava ErrCompraNaoEncontrada, got: %v", err)
	}
}

func TestConfirmarCompra_JaConfirmada(t *testing.T) {
	repo := newMockRepo()
	svc := novoServico(repo, &mockCatalogo{}, &mockEstoque{}, &mockFornecedor{})
	id := criarCompraNoRepo(t, svc, repo)
	_, _ = svc.ConfirmarCompra(context.Background(), id, uuid.New())
	_, err := svc.ConfirmarCompra(context.Background(), id, uuid.New())
	if err != domain.ErrCompraJaConfirmada {
		t.Fatalf("esperava ErrCompraJaConfirmada, got: %v", err)
	}
}

func TestConfirmarCompra_ErroEstoque(t *testing.T) {
	repo := newMockRepo()
	estErr := errors.New("estoque indisponível")
	svc := novoServico(repo, &mockCatalogo{}, &mockEstoque{err: estErr}, &mockFornecedor{})
	id := criarCompraNoRepo(t, svc, repo)
	_, err := svc.ConfirmarCompra(context.Background(), id, uuid.New())
	if err != estErr {
		t.Fatalf("esperava erro de estoque, got: %v", err)
	}
}

// ─── ListarCompras / BuscarCompra ────────────────────────────────────────────

func TestListarCompras(t *testing.T) {
	repo := newMockRepo()
	svc := novoServico(repo, &mockCatalogo{}, &mockEstoque{}, &mockFornecedor{})
	_, _ = svc.CriarCompra(context.Background(), inputValido())
	_, _ = svc.CriarCompra(context.Background(), inputValido())

	compras, err := svc.ListarCompras(context.Background(), nil, 20, 0)
	if err != nil {
		t.Fatalf("esperava sucesso, got: %v", err)
	}
	if len(compras) != 2 {
		t.Errorf("esperava 2 compras, got: %d", len(compras))
	}
}

func TestBuscarCompra_Sucesso(t *testing.T) {
	repo := newMockRepo()
	svc := novoServico(repo, &mockCatalogo{}, &mockEstoque{}, &mockFornecedor{})
	id := criarCompraNoRepo(t, svc, repo)

	c, err := svc.BuscarCompra(context.Background(), id)
	if err != nil {
		t.Fatalf("esperava sucesso, got: %v", err)
	}
	if c.ID != id {
		t.Errorf("ID incorreto")
	}
}
