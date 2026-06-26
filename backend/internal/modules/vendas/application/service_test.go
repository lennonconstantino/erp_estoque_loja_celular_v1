package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/application"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/ports"
)

// ─── Mocks ───────────────────────────────────────────────────────────────────

type mockRepo struct {
	criar         func(context.Context, *domain.Venda) error
	buscarPorID   func(context.Context, uuid.UUID) (*domain.Venda, error)
	listar        func(context.Context, *uuid.UUID, int, int) ([]domain.Venda, error)
	atualizarStatus func(context.Context, uuid.UUID, domain.StatusVenda) error
}

func (m *mockRepo) Criar(ctx context.Context, v *domain.Venda) error {
	return m.criar(ctx, v)
}
func (m *mockRepo) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Venda, error) {
	return m.buscarPorID(ctx, id)
}
func (m *mockRepo) Listar(ctx context.Context, cliID *uuid.UUID, limit, offset int) ([]domain.Venda, error) {
	return m.listar(ctx, cliID, limit, offset)
}
func (m *mockRepo) AtualizarStatus(ctx context.Context, id uuid.UUID, status domain.StatusVenda) error {
	return m.atualizarStatus(ctx, id, status)
}

type mockCatalogo struct {
	existeProduto       func(context.Context, uuid.UUID) error
	consultarSaldo      func(context.Context, uuid.UUID) (int, error)
}

func (m *mockCatalogo) ExisteProduto(ctx context.Context, id uuid.UUID) error {
	return m.existeProduto(ctx, id)
}
func (m *mockCatalogo) ConsultarSaldoProduto(ctx context.Context, id uuid.UUID) (int, error) {
	return m.consultarSaldo(ctx, id)
}

type mockEstoque struct {
	registrarSaida func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int) error
}

func (m *mockEstoque) RegistrarSaidaVenda(ctx context.Context, produtoID, vendaID, responsavelID uuid.UUID, qtd int) error {
	return m.registrarSaida(ctx, produtoID, vendaID, responsavelID, qtd)
}

type mockCliente struct {
	atualizar func(context.Context, uuid.UUID, time.Time) error
}

func (m *mockCliente) AtualizarUltimaVenda(ctx context.Context, cliID uuid.UUID, data time.Time) error {
	return m.atualizar(ctx, cliID, data)
}

type mockFiscal struct {
	emitir func(context.Context, *domain.Venda) (string, error)
}

func (m *mockFiscal) EmitirDocumento(ctx context.Context, v *domain.Venda) (string, error) {
	return m.emitir(ctx, v)
}

// ─── helper ──────────────────────────────────────────────────────────────────

func novaSvc(repo ports.VendaRepository, cat ports.CatalogoReader, est ports.EstoqueWriter, cli ports.ClienteWriter, fis ports.FiscalGateway) *application.Service {
	return application.NewService(repo, cat, est, cli, fis)
}

func okRepo() *mockRepo {
	return &mockRepo{
		criar:           func(_ context.Context, _ *domain.Venda) error { return nil },
		atualizarStatus: func(_ context.Context, _ uuid.UUID, _ domain.StatusVenda) error { return nil },
		listar:          func(_ context.Context, _ *uuid.UUID, _, _ int) ([]domain.Venda, error) { return nil, nil },
	}
}
func okCat() *mockCatalogo {
	return &mockCatalogo{
		existeProduto:  func(_ context.Context, _ uuid.UUID) error { return nil },
		consultarSaldo: func(_ context.Context, _ uuid.UUID) (int, error) { return 10, nil },
	}
}
func okEst() *mockEstoque {
	return &mockEstoque{registrarSaida: func(_ context.Context, _, _, _ uuid.UUID, _ int) error { return nil }}
}
func okCli() *mockCliente {
	return &mockCliente{atualizar: func(_ context.Context, _ uuid.UUID, _ time.Time) error { return nil }}
}
func okFiscal() *mockFiscal {
	return &mockFiscal{emitir: func(_ context.Context, _ *domain.Venda) (string, error) { return "CUP-0001", nil }}
}

// ─── CriarVenda ──────────────────────────────────────────────────────────────

func TestCriarVenda_Sucesso(t *testing.T) {
	svc := novaSvc(okRepo(), okCat(), okEst(), okCli(), okFiscal())
	v, err := svc.CriarVenda(context.Background(), ports.CriarVendaInput{
		ConsumidorFinal: true,
		FormaPgto:       domain.FormaDinheiro,
		DocFiscal:       domain.DocCupom,
		Itens:           []ports.CriarDetalheInput{{ProdutoID: uuid.New(), Quantidade: 1, PrecoUnitario: 50.0}},
	})
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if v.Status != domain.StatusRascunho {
		t.Errorf("status esperado RASCUNHO, obteve %s", v.Status)
	}
	if v.ValorTotal != 50.0 {
		t.Errorf("total esperado 50.0, obteve %.2f", v.ValorTotal)
	}
}

func TestCriarVenda_SemItens(t *testing.T) {
	svc := novaSvc(okRepo(), okCat(), okEst(), okCli(), okFiscal())
	_, err := svc.CriarVenda(context.Background(), ports.CriarVendaInput{
		ConsumidorFinal: true,
		FormaPgto:       domain.FormaDinheiro,
		DocFiscal:       domain.DocCupom,
	})
	if err != domain.ErrVendaVazia {
		t.Errorf("esperava ErrVendaVazia, obteve: %v", err)
	}
}

func TestCriarVenda_ProdutoNaoExiste(t *testing.T) {
	cat := okCat()
	cat.existeProduto = func(_ context.Context, _ uuid.UUID) error { return errors.New("produto não encontrado") }
	svc := novaSvc(okRepo(), cat, okEst(), okCli(), okFiscal())
	_, err := svc.CriarVenda(context.Background(), ports.CriarVendaInput{
		ConsumidorFinal: true,
		FormaPgto:       domain.FormaDinheiro,
		DocFiscal:       domain.DocCupom,
		Itens:           []ports.CriarDetalheInput{{ProdutoID: uuid.New(), Quantidade: 1, PrecoUnitario: 10.0}},
	})
	if err == nil {
		t.Error("esperava erro, obteve nil")
	}
}

func TestCriarVenda_DescontoMaiorQueTotal(t *testing.T) {
	svc := novaSvc(okRepo(), okCat(), okEst(), okCli(), okFiscal())
	_, err := svc.CriarVenda(context.Background(), ports.CriarVendaInput{
		ConsumidorFinal: true,
		FormaPgto:       domain.FormaDinheiro,
		DocFiscal:       domain.DocCupom,
		Desconto:        999.0,
		Itens:           []ports.CriarDetalheInput{{ProdutoID: uuid.New(), Quantidade: 1, PrecoUnitario: 10.0}},
	})
	if err != domain.ErrDescontoMaiorQueTotal {
		t.Errorf("esperava ErrDescontoMaiorQueTotal, obteve: %v", err)
	}
}

// ─── ConfirmarVenda ───────────────────────────────────────────────────────────

func vendaRascunho(cliID *uuid.UUID) *domain.Venda {
	v, _ := domain.NovaVenda(cliID, cliID == nil, domain.FormaDinheiro, domain.DocCupom, 0)
	v.Itens = []domain.DetalheVenda{{ProdutoID: uuid.New(), Quantidade: 2, PrecoUnitario: 30.0}}
	return v
}

func TestConfirmarVenda_Sucesso(t *testing.T) {
	v := vendaRascunho(nil)
	repo := okRepo()
	repo.buscarPorID = func(_ context.Context, _ uuid.UUID) (*domain.Venda, error) { return v, nil }
	svc := novaSvc(repo, okCat(), okEst(), okCli(), okFiscal())
	confirmado, err := svc.ConfirmarVenda(context.Background(), v.ID, uuid.New())
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if confirmado.Status != domain.StatusConfirmada {
		t.Errorf("status esperado CONFIRMADA, obteve %s", confirmado.Status)
	}
	if confirmado.DocFiscalNumero == "" {
		t.Error("esperava número fiscal preenchido")
	}
}

func TestConfirmarVenda_SaldoInsuficiente(t *testing.T) {
	v := vendaRascunho(nil)
	repo := okRepo()
	repo.buscarPorID = func(_ context.Context, _ uuid.UUID) (*domain.Venda, error) { return v, nil }
	cat := okCat()
	cat.consultarSaldo = func(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil }
	svc := novaSvc(repo, cat, okEst(), okCli(), okFiscal())
	_, err := svc.ConfirmarVenda(context.Background(), v.ID, uuid.New())
	if err != domain.ErrSaldoInsuficiente {
		t.Errorf("esperava ErrSaldoInsuficiente, obteve: %v", err)
	}
}

func TestConfirmarVenda_NaoEncontrada(t *testing.T) {
	repo := okRepo()
	repo.buscarPorID = func(_ context.Context, _ uuid.UUID) (*domain.Venda, error) {
		return nil, domain.ErrVendaNaoEncontrada
	}
	svc := novaSvc(repo, okCat(), okEst(), okCli(), okFiscal())
	_, err := svc.ConfirmarVenda(context.Background(), uuid.New(), uuid.New())
	if err != domain.ErrVendaNaoEncontrada {
		t.Errorf("esperava ErrVendaNaoEncontrada, obteve: %v", err)
	}
}

func TestConfirmarVenda_FiscalFalha_VendaAindaConfirmada(t *testing.T) {
	v := vendaRascunho(nil)
	repo := okRepo()
	repo.buscarPorID = func(_ context.Context, _ uuid.UUID) (*domain.Venda, error) { return v, nil }
	fiscal := &mockFiscal{emitir: func(_ context.Context, _ *domain.Venda) (string, error) {
		return "", errors.New("fiscal indisponível")
	}}
	svc := novaSvc(repo, okCat(), okEst(), okCli(), fiscal)
	confirmado, err := svc.ConfirmarVenda(context.Background(), v.ID, uuid.New())
	if err != nil {
		t.Fatalf("falha fiscal não deve reverter a venda; obteve: %v", err)
	}
	if confirmado.Status != domain.StatusConfirmada {
		t.Errorf("status esperado CONFIRMADA, obteve %s", confirmado.Status)
	}
}

func TestConfirmarVenda_ComCliente_AtualizaUltimaVenda(t *testing.T) {
	cliID := uuid.New()
	v := vendaRascunho(&cliID)
	repo := okRepo()
	repo.buscarPorID = func(_ context.Context, _ uuid.UUID) (*domain.Venda, error) { return v, nil }
	atualizado := false
	cli := &mockCliente{atualizar: func(_ context.Context, _ uuid.UUID, _ time.Time) error {
		atualizado = true
		return nil
	}}
	svc := novaSvc(repo, okCat(), okEst(), cli, okFiscal())
	if _, err := svc.ConfirmarVenda(context.Background(), v.ID, uuid.New()); err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if !atualizado {
		t.Error("AtualizarUltimaVenda não foi chamado")
	}
}

func TestListarVendas_Sucesso(t *testing.T) {
	repo := okRepo()
	repo.listar = func(_ context.Context, _ *uuid.UUID, _, _ int) ([]domain.Venda, error) {
		return []domain.Venda{{Status: domain.StatusConfirmada}}, nil
	}
	svc := novaSvc(repo, okCat(), okEst(), okCli(), okFiscal())
	vendas, err := svc.ListarVendas(context.Background(), nil, 10, 0)
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if len(vendas) != 1 {
		t.Errorf("esperava 1 venda, obteve %d", len(vendas))
	}
}
