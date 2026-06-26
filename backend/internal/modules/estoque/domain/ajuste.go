package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Ajuste representa o documento de lançamento manual de estoque (tela "Ajustar Estoque").
// É persistido em estoque.ajustes e imediatamente gera movimentações no razão.
type Ajuste struct {
	ID          uuid.UUID
	ProdutoID   uuid.UUID
	QtdEntrada  int
	QtdSaida    int
	Motivo      string
	Responsavel uuid.UUID
	CriadoEm   time.Time
}

// NovoAjuste valida e constrói um Ajuste. Retorna erro de domínio se inválido.
func NovoAjuste(produtoID, responsavelID uuid.UUID, qtdEntrada, qtdSaida int, motivo string) (*Ajuste, error) {
	a := &Ajuste{
		ID:          uuid.New(),
		ProdutoID:   produtoID,
		QtdEntrada:  qtdEntrada,
		QtdSaida:    qtdSaida,
		Motivo:      strings.TrimSpace(motivo),
		Responsavel: responsavelID,
		CriadoEm:   time.Now().UTC(),
	}
	return a, a.Validar()
}

// Validar verifica invariantes do ajuste.
func (a *Ajuste) Validar() error {
	if a.Motivo == "" {
		return ErrMotivoObrigatorio
	}
	if a.QtdEntrada < 0 || a.QtdSaida < 0 {
		return ErrAjusteQuantidadeNegativa
	}
	if a.QtdEntrada == 0 && a.QtdSaida == 0 {
		return ErrAjusteSemQuantidade
	}
	return nil
}
