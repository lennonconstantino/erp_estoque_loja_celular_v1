package domain

import (
	"testing"
	"time"
)

func TestValidarCPF(t *testing.T) {
	casos := []struct {
		nome string
		cpf  string
		ok   bool
	}{
		{"válido formatado", "529.982.247-25", true},
		{"válido só dígitos", "52998224725", true},
		{"dígito verificador errado", "52998224724", false},
		{"todos repetidos", "11111111111", false},
		{"curto", "123", false},
		{"vazio", "", false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := ValidarCPF(c.cpf); got != c.ok {
				t.Fatalf("ValidarCPF(%q) = %v, quer %v", c.cpf, got, c.ok)
			}
		})
	}
}

func TestNovoCliente(t *testing.T) {
	if _, err := NovoCliente("529.982.247-25", "Maria", "maria@x.com"); err != nil {
		t.Fatalf("cliente válido retornou erro: %v", err)
	}
	if _, err := NovoCliente("529.982.247-25", "", "maria@x.com"); err != ErrNomeObrigatorio {
		t.Fatalf("esperava ErrNomeObrigatorio, veio %v", err)
	}
	if _, err := NovoCliente("529.982.247-25", "Maria", "invalido"); err != ErrEmailInvalido {
		t.Fatalf("esperava ErrEmailInvalido, veio %v", err)
	}
	if _, err := NovoCliente("00000000000", "Maria", "maria@x.com"); err != ErrCPFInvalido {
		t.Fatalf("esperava ErrCPFInvalido, veio %v", err)
	}
}

func TestDefinirAtivo(t *testing.T) {
	c, err := NovoCliente("529.982.247-25", "Maria", "maria@x.com")
	if err != nil {
		t.Fatalf("cliente válido retornou erro: %v", err)
	}
	if !c.Ativo {
		t.Fatalf("novo cliente deveria nascer ativo")
	}

	// inativar registra a mudança (toca AtualizadoEm). Usamos um valor no passado
	// para tornar o avanço determinístico, sem depender do relógio entre chamadas.
	passado := time.Now().UTC().Add(-time.Hour)
	c.AtualizadoEm = passado
	c.DefinirAtivo(false)
	if c.Ativo {
		t.Fatalf("esperava cliente inativo após DefinirAtivo(false)")
	}
	if !c.AtualizadoEm.After(passado) {
		t.Fatalf("DefinirAtivo(false) deveria atualizar AtualizadoEm")
	}

	// definir o mesmo status é no-op: não toca AtualizadoEm
	estavel := time.Now().UTC().Add(-time.Hour)
	c.AtualizadoEm = estavel
	c.DefinirAtivo(false)
	if c.AtualizadoEm != estavel {
		t.Fatalf("DefinirAtivo com status igual não deveria tocar AtualizadoEm")
	}

	// reativar volta a registrar a mudança
	c.DefinirAtivo(true)
	if !c.Ativo {
		t.Fatalf("esperava cliente ativo após DefinirAtivo(true)")
	}
}
