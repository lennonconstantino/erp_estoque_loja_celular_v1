package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/domain"
)

var (
	fornID  = uuid.New()
	prodID  = uuid.New()
	compraID = uuid.New()
	dtHoje  = time.Now()
)

// ─── NovaCompra ──────────────────────────────────────────────────────────────

func TestNovaCompra_Valida(t *testing.T) {
	c, err := domain.NovaCompra(fornID, "NF-001", dtHoje)
	if err != nil {
		t.Fatalf("esperava sucesso, got: %v", err)
	}
	if c.Status != domain.StatusRascunho {
		t.Errorf("status inicial deve ser RASCUNHO, got: %s", c.Status)
	}
	if c.FornecedorID != fornID {
		t.Errorf("fornecedorID incorreto")
	}
	if c.NF != "NF-001" {
		t.Errorf("NF incorreta")
	}
}

func TestNovaCompra_SemFornecedor(t *testing.T) {
	_, err := domain.NovaCompra(uuid.Nil, "NF-001", dtHoje)
	if err != domain.ErrFornecedorObrigatorio {
		t.Fatalf("esperava ErrFornecedorObrigatorio, got: %v", err)
	}
}

func TestNovaCompra_SemData(t *testing.T) {
	_, err := domain.NovaCompra(fornID, "", time.Time{})
	if err != domain.ErrDataCompraObrigatoria {
		t.Fatalf("esperava ErrDataCompraObrigatoria, got: %v", err)
	}
}

func TestNovaCompra_NFVaziaPermitida(t *testing.T) {
	c, err := domain.NovaCompra(fornID, "", dtHoje)
	if err != nil {
		t.Fatalf("NF vazia deve ser permitida, got: %v", err)
	}
	if c.NF != "" {
		t.Errorf("NF deve ser vazia")
	}
}

// ─── NovoDetalheCompra ───────────────────────────────────────────────────────

func TestNovoDetalheCompra_Valido(t *testing.T) {
	d, err := domain.NovoDetalheCompra(compraID, prodID, 5, 10.0, 25.0)
	if err != nil {
		t.Fatalf("esperava sucesso, got: %v", err)
	}
	if d.Quantidade != 5 {
		t.Errorf("quantidade incorreta")
	}
	if d.PrecoCompra != 10.0 {
		t.Errorf("preço compra incorreto")
	}
	if d.PrecoVenda != 25.0 {
		t.Errorf("preço venda incorreto")
	}
	// margem = (25-10)/10*100 = 150%
	if d.Margem < 149.9 || d.Margem > 150.1 {
		t.Errorf("margem incorreta: %f", d.Margem)
	}
}

func TestNovoDetalheCompra_QuantidadeZero(t *testing.T) {
	_, err := domain.NovoDetalheCompra(compraID, prodID, 0, 10.0, 25.0)
	if err != domain.ErrQuantidadePositiva {
		t.Fatalf("esperava ErrQuantidadePositiva, got: %v", err)
	}
}

func TestNovoDetalheCompra_QuantidadeNegativa(t *testing.T) {
	_, err := domain.NovoDetalheCompra(compraID, prodID, -1, 10.0, 25.0)
	if err != domain.ErrQuantidadePositiva {
		t.Fatalf("esperava ErrQuantidadePositiva, got: %v", err)
	}
}

func TestNovoDetalheCompra_PrecoCompraZero(t *testing.T) {
	_, err := domain.NovoDetalheCompra(compraID, prodID, 1, 0.0, 25.0)
	if err != domain.ErrPrecoNaoPositivo {
		t.Fatalf("esperava ErrPrecoNaoPositivo, got: %v", err)
	}
}

func TestNovoDetalheCompra_PrecoVendaZero(t *testing.T) {
	_, err := domain.NovoDetalheCompra(compraID, prodID, 1, 10.0, 0.0)
	if err != domain.ErrPrecoNaoPositivo {
		t.Fatalf("esperava ErrPrecoNaoPositivo, got: %v", err)
	}
}

func TestNovoDetalheCompra_PrecoCompraIgualVenda(t *testing.T) {
	_, err := domain.NovoDetalheCompra(compraID, prodID, 1, 10.0, 10.0)
	if err != domain.ErrPrecoInvalido {
		t.Fatalf("esperava ErrPrecoInvalido, got: %v", err)
	}
}

func TestNovoDetalheCompra_PrecoCompraMaiorQueVenda(t *testing.T) {
	_, err := domain.NovoDetalheCompra(compraID, prodID, 1, 30.0, 10.0)
	if err != domain.ErrPrecoInvalido {
		t.Fatalf("esperava ErrPrecoInvalido, got: %v", err)
	}
}

// ─── CalcularTotal ───────────────────────────────────────────────────────────

func TestCalcularTotal(t *testing.T) {
	c, _ := domain.NovaCompra(fornID, "", dtHoje)
	d1, _ := domain.NovoDetalheCompra(c.ID, prodID, 2, 10.0, 20.0)
	d2, _ := domain.NovoDetalheCompra(c.ID, uuid.New(), 3, 5.0, 10.0)
	c.Itens = []domain.DetalheCompra{*d1, *d2}
	// total = 2*10 + 3*5 = 35
	total := c.CalcularTotal()
	if total < 34.9 || total > 35.1 {
		t.Errorf("total esperado 35.0, got %f", total)
	}
}

func TestCalcularTotal_SemItens(t *testing.T) {
	c, _ := domain.NovaCompra(fornID, "", dtHoje)
	if c.CalcularTotal() != 0 {
		t.Errorf("total sem itens deve ser 0")
	}
}

// ─── Confirmar ───────────────────────────────────────────────────────────────

func TestConfirmar_Sucesso(t *testing.T) {
	c, _ := domain.NovaCompra(fornID, "", dtHoje)
	d, _ := domain.NovoDetalheCompra(c.ID, prodID, 1, 5.0, 10.0)
	c.Itens = []domain.DetalheCompra{*d}
	if err := c.Confirmar(); err != nil {
		t.Fatalf("esperava sucesso, got: %v", err)
	}
	if c.Status != domain.StatusConfirmada {
		t.Errorf("status deve ser CONFIRMADA, got: %s", c.Status)
	}
}

func TestConfirmar_SemItens(t *testing.T) {
	c, _ := domain.NovaCompra(fornID, "", dtHoje)
	if err := c.Confirmar(); err != domain.ErrCompraVazia {
		t.Fatalf("esperava ErrCompraVazia, got: %v", err)
	}
}

func TestConfirmar_JaConfirmada(t *testing.T) {
	c, _ := domain.NovaCompra(fornID, "", dtHoje)
	d, _ := domain.NovoDetalheCompra(c.ID, prodID, 1, 5.0, 10.0)
	c.Itens = []domain.DetalheCompra{*d}
	_ = c.Confirmar()
	if err := c.Confirmar(); err != domain.ErrCompraJaConfirmada {
		t.Fatalf("esperava ErrCompraJaConfirmada, got: %v", err)
	}
}

func TestConfirmar_Cancelada(t *testing.T) {
	c, _ := domain.NovaCompra(fornID, "", dtHoje)
	d, _ := domain.NovoDetalheCompra(c.ID, prodID, 1, 5.0, 10.0)
	c.Itens = []domain.DetalheCompra{*d}
	c.Status = domain.StatusCancelada
	if err := c.Confirmar(); err != domain.ErrCompraStatusInvalido {
		t.Fatalf("esperava ErrCompraStatusInvalido, got: %v", err)
	}
}
