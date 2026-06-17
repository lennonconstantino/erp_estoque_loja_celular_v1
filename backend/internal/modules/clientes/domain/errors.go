package domain

import "errors"

// Erros de domínio do contexto clientes. Os adaptadores os traduzem para
// respostas HTTP adequadas.
var (
	ErrCPFInvalido     = errors.New("cpf inválido")
	ErrNomeObrigatorio = errors.New("nome é obrigatório")
	ErrEmailInvalido   = errors.New("email inválido")
	ErrUFInvalida      = errors.New("uf inválida")
	ErrNaoEncontrado   = errors.New("cliente não encontrado")
	ErrCPFJaCadastrado = errors.New("cpf já cadastrado")
)
