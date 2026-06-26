// Package domain contém as entidades e invariantes do contexto IAM.
// Não depende de banco, HTTP ou bibliotecas de infra.
package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// Usuario é a raiz de agregação do contexto IAM.
type Usuario struct {
	ID           uuid.UUID
	Nome         string
	Email        string
	SenhaHash    string
	Ativo        bool
	UltAcesso    *time.Time
	CriadoEm    time.Time
	AtualizadoEm time.Time
}

// RefreshToken representa um token de renovação armazenado como hash.
type RefreshToken struct {
	ID        uuid.UUID
	UsuarioID uuid.UUID
	TokenHash string
	ExpiraEm  time.Time
	Revogado  bool
	CriadoEm time.Time
}

// NovoUsuario cria um Usuario válido com a senha já hasheada.
func NovoUsuario(nome, email, senhaHash string) (*Usuario, error) {
	u := &Usuario{
		ID:           uuid.New(),
		Nome:         strings.TrimSpace(nome),
		Email:        strings.TrimSpace(strings.ToLower(email)),
		SenhaHash:    senhaHash,
		Ativo:        true,
		CriadoEm:    time.Now().UTC(),
		AtualizadoEm: time.Now().UTC(),
	}
	if err := u.Validar(); err != nil {
		return nil, err
	}
	return u, nil
}

// Validar verifica as invariantes do usuário.
func (u *Usuario) Validar() error {
	if u.Nome == "" {
		return ErrNomeObrigatorio
	}
	if !emailRe.MatchString(u.Email) {
		return ErrEmailInvalido
	}
	return nil
}

// AtualizarDados altera nome e email, revalidando as invariantes.
func (u *Usuario) AtualizarDados(nome, email string) error {
	u.Nome = strings.TrimSpace(nome)
	u.Email = strings.TrimSpace(strings.ToLower(email))
	if err := u.Validar(); err != nil {
		return err
	}
	u.AtualizadoEm = time.Now().UTC()
	return nil
}
