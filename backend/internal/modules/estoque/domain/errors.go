package domain

import "errors"

var (
	ErrProdutoNaoEncontrado     = errors.New("produto não encontrado")
	ErrMotivoObrigatorio        = errors.New("motivo do ajuste é obrigatório")
	ErrAjusteSemQuantidade      = errors.New("ajuste deve informar quantidade de entrada ou saída")
	ErrAjusteQuantidadeNegativa = errors.New("quantidade do ajuste não pode ser negativa")
	ErrSaldoInsuficiente        = errors.New("saldo insuficiente para o ajuste de saída")
)
