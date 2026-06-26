package domain

import (
	"time"

	"github.com/google/uuid"
)

// StatusVenda representa o ciclo de vida de uma venda.
type StatusVenda string

const (
	StatusRascunho   StatusVenda = "RASCUNHO"
	StatusConfirmada StatusVenda = "CONFIRMADA"
	StatusCancelada  StatusVenda = "CANCELADA"
)

// DocFiscal indica o tipo de documento fiscal emitido.
type DocFiscal string

const (
	DocCupom DocFiscal = "CUPOM"
	DocNF    DocFiscal = "NF"
)

// FormaPgto indica a forma de pagamento.
type FormaPgto string

const (
	FormaDinheiro FormaPgto = "DINHEIRO"
	FormaPIX      FormaPgto = "PIX"
	FormaDebito   FormaPgto = "DEBITO"
	FormaCredito  FormaPgto = "CREDITO"
	FormaOutro    FormaPgto = "OUTRO"
)

// DetalheVenda é um item dentro de uma venda.
type DetalheVenda struct {
	ID            uuid.UUID
	VendaID       uuid.UUID
	ProdutoID     uuid.UUID
	Quantidade    int
	PrecoUnitario float64
}

// Venda é o agregado raiz que representa uma saída de mercadoria.
type Venda struct {
	ID              uuid.UUID
	DtVenda         time.Time
	ValorTotal      float64  // total dos itens − desconto
	Desconto        float64
	FormaPgto       FormaPgto
	ClienteID       *uuid.UUID // nil quando ConsumidorFinal = true
	ConsumidorFinal bool
	DocFiscal       DocFiscal
	Status          StatusVenda
	DocFiscalNumero string // gerado pelo gateway; não persistido no banco
	Itens           []DetalheVenda
	CriadoEm        time.Time
	AtualizadoEm    time.Time
}

// NovaVenda cria uma venda em rascunho, validando as invariantes obrigatórias.
// Regra XOR: clienteID != nil ↔ consumidorFinal = false.
func NovaVenda(clienteID *uuid.UUID, consumidorFinal bool, formaPgto FormaPgto, docFiscal DocFiscal, desconto float64) (*Venda, error) {
	// Invariante XOR: exatamente uma das duas opções deve ser verdadeira
	if consumidorFinal == (clienteID != nil) {
		return nil, ErrClienteOuConsumidorFinal
	}
	if formaPgto == "" {
		return nil, ErrFormaPgtoObrigatoria
	}
	if desconto < 0 {
		return nil, ErrDescontoNaoNegativo
	}
	now := time.Now().UTC()
	return &Venda{
		ID:              uuid.New(),
		DtVenda:         now,
		Desconto:        desconto,
		FormaPgto:       formaPgto,
		ClienteID:       clienteID,
		ConsumidorFinal: consumidorFinal,
		DocFiscal:       docFiscal,
		Status:          StatusRascunho,
		CriadoEm:        now,
		AtualizadoEm:    now,
	}, nil
}

// NovoDetalheVenda cria e valida um item de venda.
func NovoDetalheVenda(vendaID, produtoID uuid.UUID, qtd int, precoUnitario float64) (*DetalheVenda, error) {
	if qtd <= 0 {
		return nil, ErrQuantidadePositiva
	}
	if precoUnitario <= 0 {
		return nil, ErrPrecoNaoPositivo
	}
	return &DetalheVenda{
		ID:            uuid.New(),
		VendaID:       vendaID,
		ProdutoID:     produtoID,
		Quantidade:    qtd,
		PrecoUnitario: precoUnitario,
	}, nil
}

// TotalItens retorna a soma de (quantidade × precoUnitario) sem desconto.
func (v *Venda) TotalItens() float64 {
	total := 0.0
	for _, item := range v.Itens {
		total += float64(item.Quantidade) * item.PrecoUnitario
	}
	return total
}

// CalcularTotal recalcula ValorTotal = TotalItens − Desconto, validando desconto.
func (v *Venda) CalcularTotal() error {
	total := v.TotalItens()
	if v.Desconto > total {
		return ErrDescontoMaiorQueTotal
	}
	v.ValorTotal = total - v.Desconto
	return nil
}

// Confirmar transiciona o status de RASCUNHO para CONFIRMADA.
func (v *Venda) Confirmar() error {
	switch v.Status {
	case StatusConfirmada:
		return ErrVendaJaConfirmada
	case StatusCancelada:
		return ErrVendaStatusInvalido
	}
	if len(v.Itens) == 0 {
		return ErrVendaVazia
	}
	v.Status = StatusConfirmada
	v.toca()
	return nil
}

func (v *Venda) toca() { v.AtualizadoEm = time.Now().UTC() }
