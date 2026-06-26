package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Produto é a raiz de agregação do subcontexto de produtos.
// Invariante central: PrecoCusto < PrecoVenda (espelhada no CHECK do banco).
type Produto struct {
	ID            uuid.UUID
	CategoriaID   uuid.UUID
	Descricao     string
	PrecoCusto    float64
	PrecoVenda    float64
	EstoqueMinimo int
	EstoqueAtual  int // saldo cacheado; fonte da verdade é estoque.movimentacoes
	GarantiaMeses int
	Modelo        string
	Disponivel    bool // true somente quando EstoqueAtual > 0
	Ativo         bool
	CriadoEm     time.Time
	AtualizadoEm time.Time
}

// NovoProduto cria um produto válido. EstoqueAtual começa em 0 e Disponivel=false;
// o saldo é gerenciado pelo módulo estoque via CatalogoWriter.
func NovoProduto(categoriaID uuid.UUID, descricao string, precoCusto, precoVenda float64, estoqueMin, garantia int, modelo string) (*Produto, error) {
	p := &Produto{
		ID:            uuid.New(),
		CategoriaID:   categoriaID,
		Descricao:     strings.TrimSpace(descricao),
		PrecoCusto:    precoCusto,
		PrecoVenda:    precoVenda,
		EstoqueMinimo: estoqueMin,
		EstoqueAtual:  0,
		GarantiaMeses: garantia,
		Modelo:        strings.TrimSpace(modelo),
		Disponivel:    false,
		Ativo:         true,
		CriadoEm:     time.Now().UTC(),
		AtualizadoEm: time.Now().UTC(),
	}
	if err := p.Validar(); err != nil {
		return nil, err
	}
	return p, nil
}

// AtualizarDados altera os dados cadastrais do produto, revalidando invariantes.
func (p *Produto) AtualizarDados(descricao string, categoriaID uuid.UUID, precoCusto, precoVenda float64, estoqueMin, garantia int, modelo string, ativo bool) error {
	p.Descricao = strings.TrimSpace(descricao)
	p.CategoriaID = categoriaID
	p.PrecoCusto = precoCusto
	p.PrecoVenda = precoVenda
	p.EstoqueMinimo = estoqueMin
	p.GarantiaMeses = garantia
	p.Modelo = strings.TrimSpace(modelo)
	p.Ativo = ativo
	if err := p.Validar(); err != nil {
		return err
	}
	p.toca()
	return nil
}

// AtualizarSaldo aplica um novo saldo e recalcula Disponivel.
// Chamado pelo módulo estoque via porta CatalogoWriter.
func (p *Produto) AtualizarSaldo(novoSaldo int) error {
	if novoSaldo < 0 {
		return ErrEstoqueAtualNaoNeg
	}
	p.EstoqueAtual = novoSaldo
	p.Disponivel = novoSaldo > 0
	p.toca()
	return nil
}

// Margem retorna a margem de lucro percentual sobre o custo.
func (p *Produto) Margem() float64 {
	if p.PrecoCusto == 0 {
		return 0
	}
	return (p.PrecoVenda - p.PrecoCusto) / p.PrecoCusto * 100
}

// Validar verifica as invariantes do produto.
func (p *Produto) Validar() error {
	if p.CategoriaID == uuid.Nil {
		return ErrCategoriaObrigatoria
	}
	if p.Descricao == "" {
		return ErrDescricaoObrigatoria
	}
	if p.PrecoCusto <= 0 || p.PrecoVenda <= 0 {
		return ErrPrecoNaoPositivo
	}
	if p.PrecoCusto >= p.PrecoVenda {
		return ErrPrecoInvalido
	}
	if p.EstoqueMinimo < 0 {
		return ErrEstoqueMinNaoNeg
	}
	if p.EstoqueAtual < 0 {
		return ErrEstoqueAtualNaoNeg
	}
	return nil
}

func (p *Produto) toca() { p.AtualizadoEm = time.Now().UTC() }
