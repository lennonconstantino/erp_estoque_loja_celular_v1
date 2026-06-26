package domain

import "errors"

var (
	ErrFornecedorObrigatorio = errors.New("fornecedor é obrigatório")
	ErrDataCompraObrigatoria = errors.New("data da compra é obrigatória")
	ErrCompraVazia           = errors.New("a compra deve ter pelo menos um item")
	ErrCompraJaConfirmada    = errors.New("compra já confirmada")
	ErrCompraStatusInvalido  = errors.New("operação inválida para o status atual da compra")
	ErrCompraNaoEncontrada   = errors.New("compra não encontrada")
	ErrQuantidadePositiva    = errors.New("quantidade deve ser maior que zero")
	ErrPrecoNaoPositivo      = errors.New("preço de compra e venda devem ser positivos")
	ErrPrecoInvalido         = errors.New("preço de compra deve ser menor que o preço de venda")
)
