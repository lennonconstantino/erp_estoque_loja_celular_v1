package domain

import "errors"

var (
	ErrCredenciaisInvalidas = errors.New("credenciais inválidas")
	ErrUsuarioInativo       = errors.New("usuário inativo")
	ErrEmailJaCadastrado    = errors.New("email já cadastrado")
	ErrNaoEncontrado        = errors.New("usuário não encontrado")
	ErrTokenInvalido        = errors.New("token inválido ou expirado")
	ErrNomeObrigatorio      = errors.New("nome é obrigatório")
	ErrEmailInvalido        = errors.New("email inválido")
	ErrSenhaFraca           = errors.New("senha deve ter no mínimo 8 caracteres")
)
