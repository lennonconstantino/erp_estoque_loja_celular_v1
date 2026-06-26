package domain_test

import (
	"testing"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/domain"
)

// cnpjValido é usado em todos os testes que precisam de um CNPJ correto.
// Dígitos verificadores calculados pelo algoritmo oficial.
const cnpjValido = "11222333000181"

func TestNovoFornecedor_Valido(t *testing.T) {
	f, err := domain.NovoFornecedor(cnpjValido, "Empresa LTDA", "Nome Fantasia", "contato@empresa.com", "11999999999", "João Silva")
	if err != nil {
		t.Fatalf("esperava fornecedor válido, got erro: %v", err)
	}
	if f.CNPJ != cnpjValido {
		t.Errorf("cnpj incorreto: %q", f.CNPJ)
	}
	if !f.Ativo {
		t.Error("novo fornecedor deve começar ativo")
	}
	if f.ID.String() == "" {
		t.Error("id deve ser preenchido")
	}
}

func TestNovoFornecedor_CNPJComMascara(t *testing.T) {
	f, err := domain.NovoFornecedor("11.222.333/0001-81", "Empresa LTDA", "Nome Fantasia", "a@b.com", "11999999999", "Contato")
	if err != nil {
		t.Fatalf("esperava aceitar cnpj com máscara: %v", err)
	}
	if f.CNPJ != cnpjValido {
		t.Errorf("cnpj não normalizado: %q", f.CNPJ)
	}
}

func TestNovoFornecedor_CNPJInvalido(t *testing.T) {
	casos := []string{"00000000000000", "12345678901234", "11111111111111", "123", ""}
	for _, c := range casos {
		_, err := domain.NovoFornecedor(c, "Empresa", "Fantasia", "a@b.com", "11999999999", "Contato")
		if err == nil {
			t.Errorf("cnpj %q deveria ser rejeitado", c)
		}
		if err != domain.ErrCNPJInvalido {
			t.Errorf("cnpj %q: esperava ErrCNPJInvalido, got %v", c, err)
		}
	}
}

func TestNovoFornecedor_RazaoSocialVazia(t *testing.T) {
	_, err := domain.NovoFornecedor(cnpjValido, "", "Fantasia", "a@b.com", "11999999999", "Contato")
	if err != domain.ErrRazaoSocialObrigatoria {
		t.Errorf("esperava ErrRazaoSocialObrigatoria, got %v", err)
	}
}

func TestNovoFornecedor_NomeFantasiaVazio(t *testing.T) {
	_, err := domain.NovoFornecedor(cnpjValido, "Empresa", "", "a@b.com", "11999999999", "Contato")
	if err != domain.ErrNomeFantasiaObrigatorio {
		t.Errorf("esperava ErrNomeFantasiaObrigatorio, got %v", err)
	}
}

func TestNovoFornecedor_EmailInvalido(t *testing.T) {
	casos := []string{"naoéemail", "@dominio.com", "sem-arroba", ""}
	for _, e := range casos {
		_, err := domain.NovoFornecedor(cnpjValido, "Empresa", "Fantasia", e, "11999999999", "Contato")
		if err != domain.ErrEmailInvalido {
			t.Errorf("email %q: esperava ErrEmailInvalido, got %v", e, err)
		}
	}
}

func TestNovoFornecedor_Telefone1Vazio(t *testing.T) {
	_, err := domain.NovoFornecedor(cnpjValido, "Empresa", "Fantasia", "a@b.com", "", "Contato")
	if err != domain.ErrTelefone1Obrigatorio {
		t.Errorf("esperava ErrTelefone1Obrigatorio, got %v", err)
	}
}

func TestNovoFornecedor_ComercialVazio(t *testing.T) {
	_, err := domain.NovoFornecedor(cnpjValido, "Empresa", "Fantasia", "a@b.com", "11999999999", "")
	if err != domain.ErrComercialObrigatorio {
		t.Errorf("esperava ErrComercialObrigatorio, got %v", err)
	}
}

func TestNovoFornecedor_EmailNormalizadoParaMinusculo(t *testing.T) {
	f, err := domain.NovoFornecedor(cnpjValido, "Empresa", "Fantasia", "CONTATO@EMPRESA.COM", "11999999999", "Contato")
	if err != nil {
		t.Fatal(err)
	}
	if f.Email != "contato@empresa.com" {
		t.Errorf("email não normalizado: %q", f.Email)
	}
}

func TestAplicarEndereco(t *testing.T) {
	f, _ := domain.NovoFornecedor(cnpjValido, "Empresa", "Fantasia", "a@b.com", "11999999999", "Contato")
	antes := f.AtualizadoEm

	f.AplicarEndereco(domain.Endereco{
		CEP:    "01310100",
		Rua:    "Av. Paulista",
		Bairro: "Bela Vista",
		Cidade: "São Paulo",
		UF:     "sp",
	})

	if f.Rua != "Av. Paulista" {
		t.Errorf("rua não preenchida: %q", f.Rua)
	}
	if f.UF != "SP" {
		t.Errorf("uf não convertida para maiúsculo: %q", f.UF)
	}
	if !f.AtualizadoEm.After(antes) {
		t.Error("AtualizadoEm não avançou após AplicarEndereco")
	}
}

func TestAtualizarDados_Valido(t *testing.T) {
	f, _ := domain.NovoFornecedor(cnpjValido, "Empresa LTDA", "Fantasia", "a@b.com", "11999999999", "Contato")
	antes := f.AtualizadoEm

	err := f.AtualizarDados("Empresa Nova LTDA", "Novo Fantasia", "novo@b.com", "11888888888", "11777777777", "Novo Contato", "Financeiro")
	if err != nil {
		t.Fatalf("esperava atualização válida, got erro: %v", err)
	}
	if f.RazaoSocial != "Empresa Nova LTDA" {
		t.Errorf("razão social não atualizada: %q", f.RazaoSocial)
	}
	if f.Financeiro != "Financeiro" {
		t.Errorf("financeiro não atualizado: %q", f.Financeiro)
	}
	if !f.AtualizadoEm.After(antes) {
		t.Error("AtualizadoEm não avançou após AtualizarDados")
	}
}

func TestAtualizarDados_UFInvalida(t *testing.T) {
	f, _ := domain.NovoFornecedor(cnpjValido, "Empresa", "Fantasia", "a@b.com", "11999999999", "Contato")
	f.UF = "SPP"
	if err := f.Validar(); err != domain.ErrUFInvalida {
		t.Errorf("esperava ErrUFInvalida, got %v", err)
	}
}

func TestValidarCNPJ(t *testing.T) {
	validos := []string{
		"11222333000181",
		"11.222.333/0001-81",
	}
	for _, c := range validos {
		if !domain.ValidarCNPJ(c) {
			t.Errorf("cnpj %q deveria ser válido", c)
		}
	}

	invalidos := []string{
		"00000000000000", // todos iguais
		"11111111111111", // todos iguais
		"12345678901234", // dígitos errados
		"1122233300018",  // tamanho errado
		"",
	}
	for _, c := range invalidos {
		if domain.ValidarCNPJ(c) {
			t.Errorf("cnpj %q deveria ser inválido", c)
		}
	}
}
