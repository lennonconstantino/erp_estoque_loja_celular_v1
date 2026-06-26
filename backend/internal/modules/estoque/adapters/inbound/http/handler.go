// Package http é o adaptador de entrada do contexto estoque: traduz HTTP
// (JSON) ↔ casos de uso (EstoqueService).
package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

// Handler agrupa os endpoints do módulo estoque.
type Handler struct {
	svc ports.EstoqueService
}

// NewHandler cria o handler a partir da porta de entrada.
func NewHandler(svc ports.EstoqueService) *Handler {
	return &Handler{svc: svc}
}

// ─── DTOs de Ajuste ──────────────────────────────────────────────────────────

type ajusteRequest struct {
	ProdutoID  string `json:"produto_id"`
	QtdEntrada int    `json:"qtd_entrada"`
	QtdSaida   int    `json:"qtd_saida"`
	Motivo     string `json:"motivo"`
}

type ajusteResponse struct {
	ID          uuid.UUID `json:"id"`
	ProdutoID   uuid.UUID `json:"produto_id"`
	QtdEntrada  int       `json:"qtd_entrada"`
	QtdSaida    int       `json:"qtd_saida"`
	Motivo      string    `json:"motivo"`
	Responsavel uuid.UUID `json:"responsavel_id"`
	CriadoEm   string    `json:"criado_em"`
}

func toAjusteResponse(a *domain.Ajuste) ajusteResponse {
	return ajusteResponse{
		ID:          a.ID,
		ProdutoID:   a.ProdutoID,
		QtdEntrada:  a.QtdEntrada,
		QtdSaida:    a.QtdSaida,
		Motivo:      a.Motivo,
		Responsavel: a.Responsavel,
		CriadoEm:   a.CriadoEm.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── DTOs de Movimentacao ────────────────────────────────────────────────────

type movimentacaoResponse struct {
	ID          uuid.UUID  `json:"id"`
	ProdutoID   uuid.UUID  `json:"produto_id"`
	Tipo        string     `json:"tipo"`
	Quantidade  int        `json:"quantidade"`
	SaldoAntes  int        `json:"saldo_antes"`
	SaldoDepois int        `json:"saldo_depois"`
	OrigemTipo  string     `json:"origem_tipo"`
	OrigemID    *uuid.UUID `json:"origem_id,omitempty"`
	CriadoEm   string     `json:"criado_em"`
}

func toMovimentacaoResponse(m *domain.Movimentacao) movimentacaoResponse {
	return movimentacaoResponse{
		ID:          m.ID,
		ProdutoID:   m.ProdutoID,
		Tipo:        string(m.Tipo),
		Quantidade:  m.Quantidade,
		SaldoAntes:  m.SaldoAntes,
		SaldoDepois: m.SaldoDepois,
		OrigemTipo:  m.OrigemTipo,
		OrigemID:    m.OrigemID,
		CriadoEm:   m.CriadoEm.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// LancarAjuste registra um ajuste manual. POST /estoque/ajustes
func (h *Handler) LancarAjuste(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	if claims == nil {
		httpserver.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "token ausente")
		return
	}
	responsavelID, err := uuid.Parse(claims.Subject)
	if err != nil {
		httpserver.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "claims inválidos")
		return
	}

	var req ajusteRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	produtoID, err := uuid.Parse(req.ProdutoID)
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "produto_id inválido")
		return
	}

	ajs, err := h.svc.LancarAjuste(r.Context(), ports.LancarAjusteInput{
		ProdutoID:     produtoID,
		QtdEntrada:    req.QtdEntrada,
		QtdSaida:      req.QtdSaida,
		Motivo:        req.Motivo,
		ResponsavelID: responsavelID,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, toAjusteResponse(ajs))
}

// ConsultarMovimentacoes retorna o razão de um produto. GET /estoque/{produtoId}
func (h *Handler) ConsultarMovimentacoes(w http.ResponseWriter, r *http.Request) {
	produtoID, err := uuid.Parse(chi.URLParam(r, "produtoId"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "produtoId inválido")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	movs, err := h.svc.ConsultarMovimentacoes(r.Context(), produtoID, limit, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]movimentacaoResponse, 0, len(movs))
	for i := range movs {
		out = append(out, toMovimentacaoResponse(&movs[i]))
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": out})
}

// ConsultarAjustes retorna os ajustes manuais de um produto. GET /estoque/{produtoId}/ajustes
func (h *Handler) ConsultarAjustes(w http.ResponseWriter, r *http.Request) {
	produtoID, err := uuid.Parse(chi.URLParam(r, "produtoId"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "produtoId inválido")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	ajs, err := h.svc.ConsultarAjustes(r.Context(), produtoID, limit, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]ajusteResponse, 0, len(ajs))
	for i := range ajs {
		out = append(out, toAjusteResponse(&ajs[i]))
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": out})
}

// ─── Mapeamento de erros de domínio ──────────────────────────────────────────

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrProdutoNaoEncontrado):
		httpserver.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrSaldoInsuficiente):
		httpserver.Error(w, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
	case errors.Is(err, domain.ErrMotivoObrigatorio),
		errors.Is(err, domain.ErrAjusteSemQuantidade),
		errors.Is(err, domain.ErrAjusteQuantidadeNegativa):
		httpserver.Error(w, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
	default:
		httpserver.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
