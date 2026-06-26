package domain

import (
	"time"

	"github.com/google/uuid"
)

// TipoMovimentacao classifica a origem de cada lançamento no razão.
type TipoMovimentacao string

const (
	TipoCompra        TipoMovimentacao = "COMPRA"
	TipoVenda         TipoMovimentacao = "VENDA"
	TipoAjusteEntrada TipoMovimentacao = "AJUSTE_ENTRADA"
	TipoAjusteSaida   TipoMovimentacao = "AJUSTE_SAIDA"
)

// Movimentacao é uma entrada imutável no razão de estoque (ledger append-only).
// Cada operação que altera saldo gera exatamente uma movimentação.
type Movimentacao struct {
	ID          uuid.UUID
	ProdutoID   uuid.UUID
	Tipo        TipoMovimentacao
	Quantidade  int
	SaldoAntes  int
	SaldoDepois int
	OrigemTipo  string
	OrigemID    *uuid.UUID
	Responsavel *uuid.UUID
	CriadoEm   time.Time
}

// NovaMovimentacao cria uma entrada no razão. Não valida: a service garante a
// consistência antes de chamar este construtor.
func NovaMovimentacao(
	produtoID uuid.UUID,
	tipo TipoMovimentacao,
	quantidade, saldoAntes, saldoDepois int,
	origemTipo string,
	origemID, responsavel *uuid.UUID,
) *Movimentacao {
	return &Movimentacao{
		ID:          uuid.New(),
		ProdutoID:   produtoID,
		Tipo:        tipo,
		Quantidade:  quantidade,
		SaldoAntes:  saldoAntes,
		SaldoDepois: saldoDepois,
		OrigemTipo:  origemTipo,
		OrigemID:    origemID,
		Responsavel: responsavel,
		CriadoEm:   time.Now().UTC(),
	}
}
