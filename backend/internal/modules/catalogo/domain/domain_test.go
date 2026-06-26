package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/domain"
)

// ─── Categoria ────────────────────────────────────────────────────────────────

func TestNovaCategoria_Valida(t *testing.T) {
	c, err := domain.NovaCategoria("Capas e Películas")
	if err != nil {
		t.Fatalf("esperava categoria válida, got %v", err)
	}
	if c.Descricao != "Capas e Películas" {
		t.Errorf("descricao errada: %q", c.Descricao)
	}
	if c.ID == uuid.Nil {
		t.Error("id deve ser preenchido")
	}
}

func TestNovaCategoria_DescricaoVazia(t *testing.T) {
	_, err := domain.NovaCategoria("")
	if err != domain.ErrDescricaoCatObrigatoria {
		t.Errorf("esperava ErrDescricaoCatObrigatoria, got %v", err)
	}
}

func TestNovaCategoria_DescricaoComEspacos(t *testing.T) {
	_, err := domain.NovaCategoria("   ")
	if err != domain.ErrDescricaoCatObrigatoria {
		t.Errorf("esperava ErrDescricaoCatObrigatoria com espaços, got %v", err)
	}
}

func TestCategoria_Atualizar_Valida(t *testing.T) {
	c, _ := domain.NovaCategoria("Original")
	if err := c.Atualizar("Nova Descrição"); err != nil {
		t.Fatalf("esperava atualização válida, got %v", err)
	}
	if c.Descricao != "Nova Descrição" {
		t.Errorf("descrição não atualizada: %q", c.Descricao)
	}
}

func TestCategoria_Atualizar_DescricaoVazia(t *testing.T) {
	c, _ := domain.NovaCategoria("Original")
	if err := c.Atualizar(""); err != domain.ErrDescricaoCatObrigatoria {
		t.Errorf("esperava ErrDescricaoCatObrigatoria, got %v", err)
	}
}

// ─── Produto ─────────────────────────────────────────────────────────────────

var catID = uuid.New()

func produtoValido(t *testing.T) *domain.Produto {
	t.Helper()
	p, err := domain.NovoProduto(catID, "Capa iPhone 15", 15.0, 49.90, 5, 12, "iPhone 15")
	if err != nil {
		t.Fatalf("NovoProduto: %v", err)
	}
	return p
}

func TestNovoProduto_Valido(t *testing.T) {
	p := produtoValido(t)
	if p.Descricao != "Capa iPhone 15" {
		t.Errorf("descricao errada: %q", p.Descricao)
	}
	if p.EstoqueAtual != 0 {
		t.Errorf("estoque inicial deve ser 0, got %d", p.EstoqueAtual)
	}
	if p.Disponivel {
		t.Error("novo produto deve começar indisponível")
	}
	if !p.Ativo {
		t.Error("novo produto deve começar ativo")
	}
}

func TestNovoProduto_CategoriaObrigatoria(t *testing.T) {
	_, err := domain.NovoProduto(uuid.Nil, "Produto", 10, 20, 0, 0, "")
	if err != domain.ErrCategoriaObrigatoria {
		t.Errorf("esperava ErrCategoriaObrigatoria, got %v", err)
	}
}

func TestNovoProduto_DescricaoObrigatoria(t *testing.T) {
	_, err := domain.NovoProduto(catID, "", 10, 20, 0, 0, "")
	if err != domain.ErrDescricaoObrigatoria {
		t.Errorf("esperava ErrDescricaoObrigatoria, got %v", err)
	}
}

func TestNovoProduto_PrecoCustoNaoPositivo(t *testing.T) {
	_, err := domain.NovoProduto(catID, "Produto", 0, 20, 0, 0, "")
	if err != domain.ErrPrecoNaoPositivo {
		t.Errorf("esperava ErrPrecoNaoPositivo, got %v", err)
	}
}

func TestNovoProduto_PrecoVendaNaoPositivo(t *testing.T) {
	_, err := domain.NovoProduto(catID, "Produto", 10, 0, 0, 0, "")
	if err != domain.ErrPrecoNaoPositivo {
		t.Errorf("esperava ErrPrecoNaoPositivo, got %v", err)
	}
}

func TestNovoProduto_PrecoInvalido(t *testing.T) {
	_, err := domain.NovoProduto(catID, "Produto", 50, 30, 0, 0, "")
	if err != domain.ErrPrecoInvalido {
		t.Errorf("esperava ErrPrecoInvalido (custo >= venda), got %v", err)
	}
}

func TestNovoProduto_PrecoIgual(t *testing.T) {
	_, err := domain.NovoProduto(catID, "Produto", 30, 30, 0, 0, "")
	if err != domain.ErrPrecoInvalido {
		t.Errorf("esperava ErrPrecoInvalido (custo == venda), got %v", err)
	}
}

func TestNovoProduto_EstoqueMinNegativo(t *testing.T) {
	_, err := domain.NovoProduto(catID, "Produto", 10, 20, -1, 0, "")
	if err != domain.ErrEstoqueMinNaoNeg {
		t.Errorf("esperava ErrEstoqueMinNaoNeg, got %v", err)
	}
}

func TestProduto_AtualizarDados_Valido(t *testing.T) {
	p := produtoValido(t)
	antes := p.AtualizadoEm
	time.Sleep(time.Millisecond)

	err := p.AtualizarDados("Nova Descrição", catID, 20.0, 59.90, 3, 6, "iPhone 16", true)
	if err != nil {
		t.Fatalf("AtualizarDados: %v", err)
	}
	if p.Descricao != "Nova Descrição" {
		t.Errorf("descrição não atualizada: %q", p.Descricao)
	}
	if p.PrecoCusto != 20.0 {
		t.Errorf("preco_custo não atualizado: %v", p.PrecoCusto)
	}
	if !p.AtualizadoEm.After(antes) {
		t.Error("AtualizadoEm não avançou")
	}
}

func TestProduto_AtualizarDados_PrecoInvalido(t *testing.T) {
	p := produtoValido(t)
	err := p.AtualizarDados("Produto", catID, 100, 50, 0, 0, "", true)
	if err != domain.ErrPrecoInvalido {
		t.Errorf("esperava ErrPrecoInvalido, got %v", err)
	}
}

func TestProduto_AtualizarDados_Desativar(t *testing.T) {
	p := produtoValido(t)
	if err := p.AtualizarDados(p.Descricao, catID, p.PrecoCusto, p.PrecoVenda, 0, 0, "", false); err != nil {
		t.Fatalf("AtualizarDados: %v", err)
	}
	if p.Ativo {
		t.Error("produto deveria estar inativo")
	}
}

func TestProduto_AtualizarSaldo_Positivo(t *testing.T) {
	p := produtoValido(t)
	antes := p.AtualizadoEm
	time.Sleep(time.Millisecond)

	if err := p.AtualizarSaldo(10); err != nil {
		t.Fatalf("AtualizarSaldo: %v", err)
	}
	if p.EstoqueAtual != 10 {
		t.Errorf("saldo incorreto: %d", p.EstoqueAtual)
	}
	if !p.Disponivel {
		t.Error("produto deve estar disponível com saldo > 0")
	}
	if !p.AtualizadoEm.After(antes) {
		t.Error("AtualizadoEm não avançou")
	}
}

func TestProduto_AtualizarSaldo_Zero(t *testing.T) {
	p := produtoValido(t)
	_ = p.AtualizarSaldo(10)
	_ = p.AtualizarSaldo(0)
	if p.Disponivel {
		t.Error("produto não deve estar disponível com saldo 0")
	}
}

func TestProduto_AtualizarSaldo_Negativo(t *testing.T) {
	p := produtoValido(t)
	if err := p.AtualizarSaldo(-1); err != domain.ErrEstoqueAtualNaoNeg {
		t.Errorf("esperava ErrEstoqueAtualNaoNeg, got %v", err)
	}
}

func TestProduto_Margem(t *testing.T) {
	p := produtoValido(t) // custo=15, venda=49.90
	margem := p.Margem()
	esperado := (49.90 - 15.0) / 15.0 * 100
	if margem < esperado-0.001 || margem > esperado+0.001 {
		t.Errorf("margem esperada %.4f, got %.4f", esperado, margem)
	}
}

func TestProduto_Margem_CustoZero(t *testing.T) {
	p := &domain.Produto{PrecoCusto: 0, PrecoVenda: 50}
	if p.Margem() != 0 {
		t.Error("margem com custo zero deve ser 0")
	}
}
