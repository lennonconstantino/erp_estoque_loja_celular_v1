package domain

import "errors"

var (
	ErrCNPJInvalido             = errors.New("cnpj inválido")
	ErrCNPJJaCadastrado         = errors.New("cnpj já cadastrado")
	ErrNaoEncontrado            = errors.New("fornecedor não encontrado")
	ErrRazaoSocialObrigatoria   = errors.New("razão social é obrigatória")
	ErrNomeFantasiaObrigatorio  = errors.New("nome fantasia é obrigatório")
	ErrEmailInvalido            = errors.New("email inválido")
	ErrTelefone1Obrigatorio     = errors.New("telefone 1 é obrigatório")
	ErrComercialObrigatorio     = errors.New("contato comercial é obrigatório")
	ErrUFInvalida               = errors.New("uf inválida")
)
