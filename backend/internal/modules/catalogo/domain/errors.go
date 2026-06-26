package domain

import "errors"

var (
	ErrCategoriaNaoEncontrada  = errors.New("categoria não encontrada")
	ErrProdutoNaoEncontrado    = errors.New("produto não encontrado")
	ErrDescricaoCatObrigatoria = errors.New("descrição da categoria é obrigatória")
	ErrDescricaoObrigatoria    = errors.New("descrição do produto é obrigatória")
	ErrCategoriaObrigatoria    = errors.New("categoria é obrigatória")
	ErrPrecoNaoPositivo        = errors.New("preços devem ser positivos")
	ErrPrecoInvalido           = errors.New("preço de custo deve ser menor que preço de venda")
	ErrEstoqueMinNaoNeg        = errors.New("estoque mínimo não pode ser negativo")
	ErrEstoqueAtualNaoNeg      = errors.New("estoque atual não pode ser negativo")
	ErrCategoriaJaCadastrada   = errors.New("categoria já cadastrada")
	ErrCategoriaComProdutos    = errors.New("categoria possui produtos cadastrados")
)
