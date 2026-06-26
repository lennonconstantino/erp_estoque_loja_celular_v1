// Package domain contém o núcleo do contexto fornecedores: a entidade Fornecedor
// e suas invariantes. Não depende de banco, HTTP ou bibliotecas de infra.
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

// Fornecedor é a raiz de agregação do contexto.
type Fornecedor struct {
	ID           uuid.UUID
	CNPJ         string     // somente dígitos (14)
	RazaoSocial  string
	NomeFantasia string
	Email        string
	Telefone1    string
	Telefone2    string
	CEP          string
	Rua          string
	Numero       string
	Complemento  string
	Bairro       string
	Cidade       string
	UF           string
	Comercial    string // contato comercial (obrigatório)
	Financeiro   string // contato financeiro (opcional)
	UltimaCompra *time.Time
	Ativo        bool
	CriadoEm    time.Time
	AtualizadoEm time.Time
}

// Endereco agrupa os campos de endereço preenchidos via CEP.
type Endereco struct {
	CEP    string
	Rua    string
	Bairro string
	Cidade string
	UF     string
}

// NovoFornecedor cria um fornecedor válido. Normaliza o CNPJ e valida as
// invariantes obrigatórias.
func NovoFornecedor(cnpj, razao, nomeFantasia, email, tel1, comercial string) (*Fornecedor, error) {
	f := &Fornecedor{
		ID:           uuid.New(),
		CNPJ:         NormalizarDigitos(cnpj),
		RazaoSocial:  strings.TrimSpace(razao),
		NomeFantasia: strings.TrimSpace(nomeFantasia),
		Email:        strings.ToLower(strings.TrimSpace(email)),
		Telefone1:    strings.TrimSpace(tel1),
		Comercial:    strings.TrimSpace(comercial),
		Ativo:        true,
		CriadoEm:    time.Now().UTC(),
		AtualizadoEm: time.Now().UTC(),
	}
	if err := f.Validar(); err != nil {
		return nil, err
	}
	return f, nil
}

// AplicarEndereco preenche os campos de endereço a partir de uma consulta de CEP.
func (f *Fornecedor) AplicarEndereco(e Endereco) {
	if e.Rua != "" {
		f.Rua = e.Rua
	}
	if e.Bairro != "" {
		f.Bairro = e.Bairro
	}
	if e.Cidade != "" {
		f.Cidade = e.Cidade
	}
	if e.UF != "" {
		f.UF = strings.ToUpper(e.UF)
	}
	f.toca()
}

// AtualizarDados altera campos editáveis, revalidando as invariantes.
func (f *Fornecedor) AtualizarDados(razao, nomeFantasia, email, tel1, tel2, comercial, financeiro string) error {
	f.RazaoSocial = strings.TrimSpace(razao)
	f.NomeFantasia = strings.TrimSpace(nomeFantasia)
	f.Email = strings.ToLower(strings.TrimSpace(email))
	f.Telefone1 = strings.TrimSpace(tel1)
	f.Telefone2 = strings.TrimSpace(tel2)
	f.Comercial = strings.TrimSpace(comercial)
	f.Financeiro = strings.TrimSpace(financeiro)
	if err := f.Validar(); err != nil {
		return err
	}
	f.toca()
	return nil
}

// Validar verifica as invariantes do fornecedor.
func (f *Fornecedor) Validar() error {
	if !ValidarCNPJ(f.CNPJ) {
		return ErrCNPJInvalido
	}
	if f.RazaoSocial == "" {
		return ErrRazaoSocialObrigatoria
	}
	if f.NomeFantasia == "" {
		return ErrNomeFantasiaObrigatorio
	}
	if !emailRe.MatchString(f.Email) {
		return ErrEmailInvalido
	}
	if f.Telefone1 == "" {
		return ErrTelefone1Obrigatorio
	}
	if f.Comercial == "" {
		return ErrComercialObrigatorio
	}
	if f.UF != "" && len(f.UF) != 2 {
		return ErrUFInvalida
	}
	return nil
}

func (f *Fornecedor) toca() { f.AtualizadoEm = time.Now().UTC() }

// NormalizarDigitos remove tudo que não for dígito.
func NormalizarDigitos(raw string) string {
	return naoDigito.ReplaceAllString(raw, "")
}

// ValidarCNPJ valida os dígitos verificadores do CNPJ (algoritmo oficial).
func ValidarCNPJ(cnpj string) bool {
	cnpj = NormalizarDigitos(cnpj)
	if len(cnpj) != 14 {
		return false
	}
	// rejeita sequências de dígitos repetidos (ex.: 00000000000000)
	repetido := true
	for i := 1; i < 14; i++ {
		if cnpj[i] != cnpj[0] {
			repetido = false
			break
		}
	}
	if repetido {
		return false
	}

	calcDigito := func(n int) int {
		pesos := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
		inicio := len(pesos) - n
		soma := 0
		for i := 0; i < n; i++ {
			soma += int(cnpj[i]-'0') * pesos[inicio+i]
		}
		r := soma % 11
		if r < 2 {
			return 0
		}
		return 11 - r
	}

	d1 := calcDigito(12)
	if d1 != int(cnpj[12]-'0') {
		return false
	}
	d2 := calcDigito(13)
	return d2 == int(cnpj[13]-'0')
}
