package domain_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/domain"
)

// ─── NovaVenda ────────────────────────────────────────────────────────────────

func TestNovaVenda_ConsumidorFinal(t *testing.T) {
	v, err := domain.NovaVenda(nil, true, domain.FormaDinheiro, domain.DocCupom, 0)
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if v.Status != domain.StatusRascunho {
		t.Errorf("status esperado RASCUNHO, obteve %s", v.Status)
	}
}

func TestNovaVenda_ComCliente(t *testing.T) {
	cliID := uuid.New()
	v, err := domain.NovaVenda(&cliID, false, domain.FormaPIX, domain.DocNF, 0)
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if v.ClienteID == nil || *v.ClienteID != cliID {
		t.Error("clienteID não foi setado corretamente")
	}
}

func TestNovaVenda_XORViolado_AmbosNil(t *testing.T) {
	_, err := domain.NovaVenda(nil, false, domain.FormaDinheiro, domain.DocCupom, 0)
	if err != domain.ErrClienteOuConsumidorFinal {
		t.Errorf("esperava ErrClienteOuConsumidorFinal, obteve: %v", err)
	}
}

func TestNovaVenda_XORViolado_AmbosSet(t *testing.T) {
	cliID := uuid.New()
	_, err := domain.NovaVenda(&cliID, true, domain.FormaDinheiro, domain.DocCupom, 0)
	if err != domain.ErrClienteOuConsumidorFinal {
		t.Errorf("esperava ErrClienteOuConsumidorFinal, obteve: %v", err)
	}
}

func TestNovaVenda_SemFormaPgto(t *testing.T) {
	_, err := domain.NovaVenda(nil, true, "", domain.DocCupom, 0)
	if err != domain.ErrFormaPgtoObrigatoria {
		t.Errorf("esperava ErrFormaPgtoObrigatoria, obteve: %v", err)
	}
}

func TestNovaVenda_DescontoNegativo(t *testing.T) {
	_, err := domain.NovaVenda(nil, true, domain.FormaDinheiro, domain.DocCupom, -1)
	if err != domain.ErrDescontoNaoNegativo {
		t.Errorf("esperava ErrDescontoNaoNegativo, obteve: %v", err)
	}
}

// ─── NovoDetalheVenda ─────────────────────────────────────────────────────────

func TestNovoDetalheVenda_Valido(t *testing.T) {
	d, err := domain.NovoDetalheVenda(uuid.New(), uuid.New(), 3, 29.90)
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if d.Quantidade != 3 {
		t.Errorf("quantidade esperada 3, obteve %d", d.Quantidade)
	}
}

func TestNovoDetalheVenda_QuantidadeZero(t *testing.T) {
	_, err := domain.NovoDetalheVenda(uuid.New(), uuid.New(), 0, 10.0)
	if err != domain.ErrQuantidadePositiva {
		t.Errorf("esperava ErrQuantidadePositiva, obteve: %v", err)
	}
}

func TestNovoDetalheVenda_PrecoZero(t *testing.T) {
	_, err := domain.NovoDetalheVenda(uuid.New(), uuid.New(), 1, 0)
	if err != domain.ErrPrecoNaoPositivo {
		t.Errorf("esperava ErrPrecoNaoPositivo, obteve: %v", err)
	}
}

// ─── CalcularTotal ────────────────────────────────────────────────────────────

func TestCalcularTotal_SemDesconto(t *testing.T) {
	v, _ := domain.NovaVenda(nil, true, domain.FormaDinheiro, domain.DocCupom, 0)
	v.Itens = []domain.DetalheVenda{
		{Quantidade: 2, PrecoUnitario: 50.0},
		{Quantidade: 1, PrecoUnitario: 30.0},
	}
	if err := v.CalcularTotal(); err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if v.ValorTotal != 130.0 {
		t.Errorf("total esperado 130.0, obteve %.2f", v.ValorTotal)
	}
}

func TestCalcularTotal_ComDesconto(t *testing.T) {
	v, _ := domain.NovaVenda(nil, true, domain.FormaDinheiro, domain.DocCupom, 20.0)
	v.Itens = []domain.DetalheVenda{{Quantidade: 1, PrecoUnitario: 100.0}}
	if err := v.CalcularTotal(); err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if v.ValorTotal != 80.0 {
		t.Errorf("total esperado 80.0, obteve %.2f", v.ValorTotal)
	}
}

func TestCalcularTotal_DescontoMaiorQueTotal(t *testing.T) {
	v, _ := domain.NovaVenda(nil, true, domain.FormaDinheiro, domain.DocCupom, 200.0)
	v.Itens = []domain.DetalheVenda{{Quantidade: 1, PrecoUnitario: 100.0}}
	if err := v.CalcularTotal(); err != domain.ErrDescontoMaiorQueTotal {
		t.Errorf("esperava ErrDescontoMaiorQueTotal, obteve: %v", err)
	}
}

// ─── Confirmar ────────────────────────────────────────────────────────────────

func TestConfirmar_Sucesso(t *testing.T) {
	v, _ := domain.NovaVenda(nil, true, domain.FormaDinheiro, domain.DocCupom, 0)
	v.Itens = []domain.DetalheVenda{{Quantidade: 1, PrecoUnitario: 10.0}}
	if err := v.Confirmar(); err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if v.Status != domain.StatusConfirmada {
		t.Errorf("status esperado CONFIRMADA, obteve %s", v.Status)
	}
}

func TestConfirmar_SemItens(t *testing.T) {
	v, _ := domain.NovaVenda(nil, true, domain.FormaDinheiro, domain.DocCupom, 0)
	if err := v.Confirmar(); err != domain.ErrVendaVazia {
		t.Errorf("esperava ErrVendaVazia, obteve: %v", err)
	}
}

func TestConfirmar_JaConfirmada(t *testing.T) {
	v, _ := domain.NovaVenda(nil, true, domain.FormaDinheiro, domain.DocCupom, 0)
	v.Status = domain.StatusConfirmada
	if err := v.Confirmar(); err != domain.ErrVendaJaConfirmada {
		t.Errorf("esperava ErrVendaJaConfirmada, obteve: %v", err)
	}
}

func TestConfirmar_Cancelada(t *testing.T) {
	v, _ := domain.NovaVenda(nil, true, domain.FormaDinheiro, domain.DocCupom, 0)
	v.Status = domain.StatusCancelada
	v.Itens = []domain.DetalheVenda{{Quantidade: 1, PrecoUnitario: 10.0}}
	if err := v.Confirmar(); err != domain.ErrVendaStatusInvalido {
		t.Errorf("esperava ErrVendaStatusInvalido, obteve: %v", err)
	}
}
