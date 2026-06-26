package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/domain"
)

// --- mock ---

type mockRepo struct {
	abaixo   []domain.ProdutoAbaixoMinimo
	mais     []domain.ProdutoVendido
	menos    []domain.ProdutoVendido
	vendas   *domain.ResumoVendas
	compras  *domain.ResumoCompras
	err      error
}

func (m *mockRepo) ListarAbaixoDoMinimo(_ context.Context) ([]domain.ProdutoAbaixoMinimo, error) {
	return m.abaixo, m.err
}
func (m *mockRepo) ListarMaisVendidos(_ context.Context, _, _ time.Time, _ int) ([]domain.ProdutoVendido, error) {
	return m.mais, m.err
}
func (m *mockRepo) ListarMenosVendidos(_ context.Context, _, _ time.Time, _ int) ([]domain.ProdutoVendido, error) {
	return m.menos, m.err
}
func (m *mockRepo) AggregarVendas(_ context.Context, _, _ time.Time) (*domain.ResumoVendas, error) {
	return m.vendas, m.err
}
func (m *mockRepo) AggregarCompras(_ context.Context, _, _ time.Time) (*domain.ResumoCompras, error) {
	return m.compras, m.err
}

// --- testes ---

var (
	ctx     = context.Background()
	de      = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	ate     = time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	errFake = errors.New("falha simulada")
)

func TestProdutosAbaixoDoMinimo_Sucesso(t *testing.T) {
	lista := []domain.ProdutoAbaixoMinimo{
		{ID: uuid.New(), Descricao: "Capa iPhone", EstoqueAtual: 1, EstoqueMinimo: 5, Defasagem: 4},
	}
	svc := application.NewService(&mockRepo{abaixo: lista})
	got, err := svc.ProdutosAbaixoDoMinimo(ctx)
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("esperava 1 item, obteve %d", len(got))
	}
}

func TestProdutosAbaixoDoMinimo_Erro(t *testing.T) {
	svc := application.NewService(&mockRepo{err: errFake})
	_, err := svc.ProdutosAbaixoDoMinimo(ctx)
	if !errors.Is(err, errFake) {
		t.Fatalf("esperava errFake, obteve %v", err)
	}
}

func TestMaisVendidos_LimiteDefault(t *testing.T) {
	// limite 0 deve ser normalizado para 10 internamente
	svc := application.NewService(&mockRepo{mais: []domain.ProdutoVendido{}})
	got, err := svc.MaisVendidos(ctx, de, ate, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("esperava slice não-nil")
	}
}

func TestMaisVendidos_Erro(t *testing.T) {
	svc := application.NewService(&mockRepo{err: errFake})
	_, err := svc.MaisVendidos(ctx, de, ate, 5)
	if !errors.Is(err, errFake) {
		t.Fatalf("esperava errFake, obteve %v", err)
	}
}

func TestMenosVendidos_Sucesso(t *testing.T) {
	lista := []domain.ProdutoVendido{
		{ProdutoID: uuid.New(), Descricao: "Película", TotalVendido: 1, TotalValor: 15.0},
	}
	svc := application.NewService(&mockRepo{menos: lista})
	got, err := svc.MenosVendidos(ctx, de, ate, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("esperava 1, obteve %d", len(got))
	}
}

func TestMenosVendidos_LimiteDefault(t *testing.T) {
	svc := application.NewService(&mockRepo{menos: []domain.ProdutoVendido{}})
	_, err := svc.MenosVendidos(ctx, de, ate, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResumoVendas_Sucesso(t *testing.T) {
	resumo := &domain.ResumoVendas{TotalVendas: 3, ValorTotal: 300, TicketMedio: 100, De: de, Ate: ate}
	svc := application.NewService(&mockRepo{vendas: resumo})
	got, err := svc.ResumoVendas(ctx, de, ate)
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalVendas != 3 {
		t.Fatalf("esperava 3, obteve %d", got.TotalVendas)
	}
}

func TestResumoVendas_Erro(t *testing.T) {
	svc := application.NewService(&mockRepo{err: errFake})
	_, err := svc.ResumoVendas(ctx, de, ate)
	if !errors.Is(err, errFake) {
		t.Fatalf("esperava errFake, obteve %v", err)
	}
}

func TestResumoCompras_Sucesso(t *testing.T) {
	resumo := &domain.ResumoCompras{TotalCompras: 2, ValorTotal: 500, De: de, Ate: ate}
	svc := application.NewService(&mockRepo{compras: resumo})
	got, err := svc.ResumoCompras(ctx, de, ate)
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCompras != 2 {
		t.Fatalf("esperava 2, obteve %d", got.TotalCompras)
	}
}

func TestResumoCompras_Erro(t *testing.T) {
	svc := application.NewService(&mockRepo{err: errFake})
	_, err := svc.ResumoCompras(ctx, de, ate)
	if !errors.Is(err, errFake) {
		t.Fatalf("esperava errFake, obteve %v", err)
	}
}
