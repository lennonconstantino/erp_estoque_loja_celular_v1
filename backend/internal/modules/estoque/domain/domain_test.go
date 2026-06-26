package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/domain"
)

// ─── Ajuste ──────────────────────────────────────────────────────────────────

func TestNovoAjuste_Valido_Entrada(t *testing.T) {
	prodID := uuid.New()
	respID := uuid.New()
	a, err := domain.NovoAjuste(prodID, respID, 10, 0, "inventário físico")
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if a.QtdEntrada != 10 || a.QtdSaida != 0 {
		t.Errorf("quantidades incorretas: entrada=%d saida=%d", a.QtdEntrada, a.QtdSaida)
	}
	if a.ProdutoID != prodID {
		t.Error("ProdutoID não propagado")
	}
	if a.Responsavel != respID {
		t.Error("Responsavel não propagado")
	}
}

func TestNovoAjuste_Valido_Saida(t *testing.T) {
	a, err := domain.NovoAjuste(uuid.New(), uuid.New(), 0, 5, "quebra/perda")
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if a.QtdSaida != 5 {
		t.Errorf("qtd_saida esperada 5, obteve %d", a.QtdSaida)
	}
}

func TestNovoAjuste_Valido_EntradaESaida(t *testing.T) {
	a, err := domain.NovoAjuste(uuid.New(), uuid.New(), 10, 3, "recontagem")
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if a.QtdEntrada != 10 || a.QtdSaida != 3 {
		t.Errorf("quantidades incorretas: entrada=%d saida=%d", a.QtdEntrada, a.QtdSaida)
	}
}

func TestNovoAjuste_MotivoVazio(t *testing.T) {
	_, err := domain.NovoAjuste(uuid.New(), uuid.New(), 10, 0, "")
	if !errors.Is(err, domain.ErrMotivoObrigatorio) {
		t.Errorf("esperava ErrMotivoObrigatorio, obteve: %v", err)
	}
}

func TestNovoAjuste_MotivoSomenteEspacos(t *testing.T) {
	_, err := domain.NovoAjuste(uuid.New(), uuid.New(), 10, 0, "   ")
	if !errors.Is(err, domain.ErrMotivoObrigatorio) {
		t.Errorf("esperava ErrMotivoObrigatorio, obteve: %v", err)
	}
}

func TestNovoAjuste_SemQuantidade(t *testing.T) {
	_, err := domain.NovoAjuste(uuid.New(), uuid.New(), 0, 0, "motivo qualquer")
	if !errors.Is(err, domain.ErrAjusteSemQuantidade) {
		t.Errorf("esperava ErrAjusteSemQuantidade, obteve: %v", err)
	}
}

func TestNovoAjuste_QuantidadeEntradaNegativa(t *testing.T) {
	_, err := domain.NovoAjuste(uuid.New(), uuid.New(), -1, 0, "motivo")
	if !errors.Is(err, domain.ErrAjusteQuantidadeNegativa) {
		t.Errorf("esperava ErrAjusteQuantidadeNegativa, obteve: %v", err)
	}
}

func TestNovoAjuste_QuantidadeSaidaNegativa(t *testing.T) {
	_, err := domain.NovoAjuste(uuid.New(), uuid.New(), 0, -5, "motivo")
	if !errors.Is(err, domain.ErrAjusteQuantidadeNegativa) {
		t.Errorf("esperava ErrAjusteQuantidadeNegativa, obteve: %v", err)
	}
}

func TestAjuste_MotivoComEspacosTrimado(t *testing.T) {
	a, err := domain.NovoAjuste(uuid.New(), uuid.New(), 1, 0, "  inventário  ")
	if err != nil {
		t.Fatalf("não esperava erro, obteve: %v", err)
	}
	if a.Motivo != "inventário" {
		t.Errorf("motivo deveria ser trimado, obteve: %q", a.Motivo)
	}
}

// ─── Movimentacao ────────────────────────────────────────────────────────────

func TestNovaMovimentacao_AjusteEntrada(t *testing.T) {
	prodID := uuid.New()
	origID := uuid.New()
	respID := uuid.New()
	m := domain.NovaMovimentacao(prodID, domain.TipoAjusteEntrada, 5, 3, 8, "AJUSTE", &origID, &respID)
	if m.ID == uuid.Nil {
		t.Error("ID não gerado")
	}
	if m.ProdutoID != prodID {
		t.Error("ProdutoID incorreto")
	}
	if m.Tipo != domain.TipoAjusteEntrada {
		t.Errorf("tipo incorreto: %s", m.Tipo)
	}
	if m.Quantidade != 5 || m.SaldoAntes != 3 || m.SaldoDepois != 8 {
		t.Errorf("valores incorretos: qtd=%d ant=%d dep=%d", m.Quantidade, m.SaldoAntes, m.SaldoDepois)
	}
	if m.OrigemTipo != "AJUSTE" {
		t.Errorf("origem_tipo incorreto: %s", m.OrigemTipo)
	}
	if m.OrigemID == nil || *m.OrigemID != origID {
		t.Error("OrigemID incorreto")
	}
	if m.CriadoEm.IsZero() {
		t.Error("CriadoEm não preenchido")
	}
}

func TestNovaMovimentacao_AjusteSaida(t *testing.T) {
	m := domain.NovaMovimentacao(uuid.New(), domain.TipoAjusteSaida, 2, 10, 8, "AJUSTE", nil, nil)
	if m.Tipo != domain.TipoAjusteSaida {
		t.Errorf("tipo incorreto: %s", m.Tipo)
	}
	if m.OrigemID != nil || m.Responsavel != nil {
		t.Error("campos opcionais deveriam ser nil")
	}
}

func TestNovaMovimentacao_TiposConst(t *testing.T) {
	tipos := []domain.TipoMovimentacao{
		domain.TipoCompra,
		domain.TipoVenda,
		domain.TipoAjusteEntrada,
		domain.TipoAjusteSaida,
	}
	for _, tipo := range tipos {
		if tipo == "" {
			t.Errorf("tipo vazio: %v", tipo)
		}
	}
}
