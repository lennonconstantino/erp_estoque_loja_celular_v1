package domain_test

import (
	"testing"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/domain"
)

func TestNovoUsuario_Valido(t *testing.T) {
	u, err := domain.NovoUsuario("Ana Costa", "ana@loja.local", "$2a$10$hash")
	if err != nil {
		t.Fatalf("esperava usuário válido, got erro: %v", err)
	}
	if u.Nome != "Ana Costa" {
		t.Errorf("nome incorreto: %q", u.Nome)
	}
	if u.Email != "ana@loja.local" {
		t.Errorf("email incorreto: %q", u.Email)
	}
	if !u.Ativo {
		t.Error("novo usuário deve começar ativo")
	}
	if u.ID.String() == "" {
		t.Error("id deve ser preenchido")
	}
}

func TestNovoUsuario_NomeVazio(t *testing.T) {
	_, err := domain.NovoUsuario("", "ana@loja.local", "hash")
	if err == nil {
		t.Fatal("esperava erro por nome vazio")
	}
	if err != domain.ErrNomeObrigatorio {
		t.Errorf("erro esperado ErrNomeObrigatorio, got %v", err)
	}
}

func TestNovoUsuario_EmailInvalido(t *testing.T) {
	casos := []string{"naoéemail", "@dominio.com", "sem-arroba", ""}
	for _, e := range casos {
		_, err := domain.NovoUsuario("Ana", e, "hash")
		if err == nil {
			t.Errorf("email %q deveria ser rejeitado", e)
		}
		if err != domain.ErrEmailInvalido {
			t.Errorf("email %q: erro esperado ErrEmailInvalido, got %v", e, err)
		}
	}
}

func TestNovoUsuario_EmailNormalizadoParaMinusculo(t *testing.T) {
	u, err := domain.NovoUsuario("Ana", "Ana@LOJA.LOCAL", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "ana@loja.local" {
		t.Errorf("email não foi normalizado: %q", u.Email)
	}
}

func TestNovoUsuario_NomeComEspacos(t *testing.T) {
	u, err := domain.NovoUsuario("  Ana Costa  ", "ana@loja.local", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if u.Nome != "Ana Costa" {
		t.Errorf("espaços do nome não foram removidos: %q", u.Nome)
	}
}

func TestAtualizarDados_Valido(t *testing.T) {
	u, _ := domain.NovoUsuario("Ana", "ana@loja.local", "hash")
	antes := u.AtualizadoEm

	if err := u.AtualizarDados("Ana Costa", "ana.costa@loja.local"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if u.Nome != "Ana Costa" {
		t.Errorf("nome não foi atualizado: %q", u.Nome)
	}
	if u.Email != "ana.costa@loja.local" {
		t.Errorf("email não foi atualizado: %q", u.Email)
	}
	if !u.AtualizadoEm.After(antes) {
		t.Error("AtualizadoEm não avançou após atualização")
	}
}

func TestAtualizarDados_EmailInvalido(t *testing.T) {
	u, _ := domain.NovoUsuario("Ana", "ana@loja.local", "hash")
	if err := u.AtualizarDados("Ana", "invalido"); err == nil {
		t.Fatal("esperava erro por email inválido")
	}
}

func TestAtualizarDados_NomeVazio(t *testing.T) {
	u, _ := domain.NovoUsuario("Ana", "ana@loja.local", "hash")
	if err := u.AtualizarDados("", "ana@loja.local"); err == nil {
		t.Fatal("esperava erro por nome vazio")
	}
}
