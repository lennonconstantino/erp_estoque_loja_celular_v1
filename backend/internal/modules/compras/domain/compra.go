// Package domain contém o núcleo do contexto compras: entidades, invariantes e erros.
// Sem dependências de banco ou HTTP.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// StatusCompra representa o ciclo de vida de uma compra.
type StatusCompra string

const (
	StatusRascunho   StatusCompra = "RASCUNHO"
	StatusConfirmada StatusCompra = "CONFIRMADA"
	StatusCancelada  StatusCompra = "CANCELADA"
)

// DetalheCompra é um item dentro de uma compra.
type DetalheCompra struct {
	ID          uuid.UUID
	CompraID    uuid.UUID
	ProdutoID   uuid.UUID
	Quantidade  int
	PrecoCompra float64
	PrecoVenda  float64
	Margem      float64 // percentual sobre o custo
}

// Compra é o agregado raiz que representa uma entrada de mercadoria.
type Compra struct {
	ID           uuid.UUID
	FornecedorID uuid.UUID
	NF           string
	DtCompra     time.Time
	ValorTotal   float64
	Status       StatusCompra
	Itens        []DetalheCompra
	CriadoEm    time.Time
	AtualizadoEm time.Time
}

// NovaCompra cria uma compra em rascunho, validando as invariantes obrigatórias.
func NovaCompra(fornecedorID uuid.UUID, nf string, dtCompra time.Time) (*Compra, error) {
	if fornecedorID == uuid.Nil {
		return nil, ErrFornecedorObrigatorio
	}
	if dtCompra.IsZero() {
		return nil, ErrDataCompraObrigatoria
	}
	return &Compra{
		ID:           uuid.New(),
		FornecedorID: fornecedorID,
		NF:           nf,
		DtCompra:     dtCompra.UTC(),
		Status:       StatusRascunho,
		CriadoEm:    time.Now().UTC(),
		AtualizadoEm: time.Now().UTC(),
	}, nil
}

// NovoDetalheCompra cria e valida um item de compra.
// Invariante: qtd > 0, precoCompra > 0, precoCompra < precoVenda.
func NovoDetalheCompra(compraID, produtoID uuid.UUID, qtd int, precoCompra, precoVenda float64) (*DetalheCompra, error) {
	if qtd <= 0 {
		return nil, ErrQuantidadePositiva
	}
	if precoCompra <= 0 || precoVenda <= 0 {
		return nil, ErrPrecoNaoPositivo
	}
	if precoCompra >= precoVenda {
		return nil, ErrPrecoInvalido
	}
	margem := (precoVenda - precoCompra) / precoCompra * 100
	return &DetalheCompra{
		ID:          uuid.New(),
		CompraID:    compraID,
		ProdutoID:   produtoID,
		Quantidade:  qtd,
		PrecoCompra: precoCompra,
		PrecoVenda:  precoVenda,
		Margem:      margem,
	}, nil
}

// CalcularTotal retorna a soma de (quantidade × custo) de todos os itens.
func (c *Compra) CalcularTotal() float64 {
	total := 0.0
	for _, item := range c.Itens {
		total += float64(item.Quantidade) * item.PrecoCompra
	}
	return total
}

// Confirmar transiciona o status de RASCUNHO para CONFIRMADA.
// Retorna erro se a compra não estiver em RASCUNHO ou não tiver itens.
func (c *Compra) Confirmar() error {
	switch c.Status {
	case StatusConfirmada:
		return ErrCompraJaConfirmada
	case StatusCancelada:
		return ErrCompraStatusInvalido
	}
	if len(c.Itens) == 0 {
		return ErrCompraVazia
	}
	c.Status = StatusConfirmada
	c.toca()
	return nil
}

func (c *Compra) toca() { c.AtualizadoEm = time.Now().UTC() }
