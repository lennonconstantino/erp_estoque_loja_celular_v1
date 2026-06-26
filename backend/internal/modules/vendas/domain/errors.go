// Package domain contém o núcleo do contexto vendas: entidades, invariantes e erros.
package domain

import "errors"

var (
	ErrVendaNaoEncontrada      = errors.New("venda não encontrada")
	ErrVendaJaConfirmada       = errors.New("venda já confirmada")
	ErrVendaStatusInvalido     = errors.New("operação inválida para o status atual da venda")
	ErrVendaVazia              = errors.New("a venda deve ter pelo menos um item")
	ErrSaldoInsuficiente       = errors.New("saldo insuficiente")
	ErrClienteOuConsumidorFinal = errors.New("informe o cliente ou marque como consumidor final")
	ErrFormaPgtoObrigatoria    = errors.New("forma de pagamento é obrigatória")
	ErrQuantidadePositiva      = errors.New("quantidade deve ser maior que zero")
	ErrPrecoNaoPositivo        = errors.New("preço unitário deve ser positivo")
	ErrDescontoNaoNegativo     = errors.New("desconto não pode ser negativo")
	ErrDescontoMaiorQueTotal   = errors.New("desconto não pode ser maior que o total dos itens")
)
