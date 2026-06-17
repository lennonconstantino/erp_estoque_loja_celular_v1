// Package domain contém o núcleo do contexto clientes: a entidade Cliente e
// suas invariantes. Não depende de banco, HTTP ou bibliotecas de infra.
package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	naoDigito = regexp.MustCompile(`\D`)
	emailRe   = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

// Cliente é a raiz de agregação do contexto.
type Cliente struct {
	ID           uuid.UUID
	CPF          string // somente dígitos (11)
	Nome         string
	Email        string
	Telefone     string
	CEP          string
	Rua          string
	Numero       string
	Complemento  string
	Bairro       string
	Cidade       string
	UF           string
	UltimaCompra *time.Time
	Ativo        bool
	CriadoEm     time.Time
	AtualizadoEm time.Time
}

// Endereco agrupa os campos de endereço (preenchidos por consulta de CEP).
type Endereco struct {
	CEP    string
	Rua    string
	Bairro string
	Cidade string
	UF     string
}

// NovoCliente cria um cliente válido. Normaliza o CPF e valida as invariantes
// obrigatórias (CPF, Nome, Email).
func NovoCliente(cpf, nome, email string) (*Cliente, error) {
	c := &Cliente{
		ID:           uuid.New(),
		CPF:          NormalizarCPF(cpf),
		Nome:         strings.TrimSpace(nome),
		Email:        strings.TrimSpace(email),
		Ativo:        true,
		CriadoEm:     time.Now().UTC(),
		AtualizadoEm: time.Now().UTC(),
	}
	if err := c.Validar(); err != nil {
		return nil, err
	}
	return c, nil
}

// AplicarEndereco preenche os campos de endereço a partir de uma consulta de CEP.
func (c *Cliente) AplicarEndereco(e Endereco) {
	if e.Rua != "" {
		c.Rua = e.Rua
	}
	if e.Bairro != "" {
		c.Bairro = e.Bairro
	}
	if e.Cidade != "" {
		c.Cidade = e.Cidade
	}
	if e.UF != "" {
		c.UF = strings.ToUpper(e.UF)
	}
	c.toca()
}

// AtualizarContato altera nome/email/telefone, revalidando as invariantes.
func (c *Cliente) AtualizarContato(nome, email, telefone string) error {
	c.Nome = strings.TrimSpace(nome)
	c.Email = strings.TrimSpace(email)
	c.Telefone = strings.TrimSpace(telefone)
	if err := c.Validar(); err != nil {
		return err
	}
	c.toca()
	return nil
}

// Validar verifica as invariantes do cliente.
func (c *Cliente) Validar() error {
	if !ValidarCPF(c.CPF) {
		return ErrCPFInvalido
	}
	if c.Nome == "" {
		return ErrNomeObrigatorio
	}
	if !emailRe.MatchString(c.Email) {
		return ErrEmailInvalido
	}
	if c.UF != "" && len(c.UF) != 2 {
		return ErrUFInvalida
	}
	return nil
}

func (c *Cliente) toca() { c.AtualizadoEm = time.Now().UTC() }

// NormalizarCPF remove tudo que não for dígito.
func NormalizarCPF(raw string) string {
	return naoDigito.ReplaceAllString(raw, "")
}

// ValidarCPF valida os dígitos verificadores do CPF (algoritmo oficial).
func ValidarCPF(cpf string) bool {
	cpf = NormalizarCPF(cpf)
	if len(cpf) != 11 {
		return false
	}
	// rejeita sequências de dígitos repetidos (ex.: 00000000000)
	repetido := true
	for i := 1; i < 11; i++ {
		if cpf[i] != cpf[0] {
			repetido = false
			break
		}
	}
	if repetido {
		return false
	}

	digito := func(n int) int {
		soma := 0
		for i := 0; i < n; i++ {
			soma += int(cpf[i]-'0') * (n + 1 - i)
		}
		r := (soma * 10) % 11
		if r == 10 {
			r = 0
		}
		return r
	}
	return digito(9) == int(cpf[9]-'0') && digito(10) == int(cpf[10]-'0')
}
