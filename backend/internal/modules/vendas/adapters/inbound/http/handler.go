// Package http expõe o módulo vendas via HTTP.
package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

// Handler agrupa os handlers HTTP do módulo vendas.
type Handler struct {
	svc ports.VendaService
}

// NewHandler cria o handler injetando o serviço.
func NewHandler(svc ports.VendaService) *Handler {
	return &Handler{svc: svc}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type detalheRequest struct {
	ProdutoID     string  `json:"produto_id"`
	Quantidade    int     `json:"quantidade"`
	PrecoUnitario float64 `json:"preco_unitario"`
}

type criarVendaRequest struct {
	ClienteID       string           `json:"cliente_id"`
	ConsumidorFinal bool             `json:"consumidor_final"`
	FormaPgto       string           `json:"forma_pgto"`
	DocFiscal       string           `json:"doc_fiscal"`
	Desconto        float64          `json:"desconto"`
	Itens           []detalheRequest `json:"itens"`
}

type detalheResponse struct {
	ID            string  `json:"id"`
	ProdutoID     string  `json:"produto_id"`
	Quantidade    int     `json:"quantidade"`
	PrecoUnitario float64 `json:"preco_unitario"`
}

type vendaResponse struct {
	ID              string            `json:"id"`
	DtVenda         time.Time         `json:"dt_venda"`
	ValorTotal      float64           `json:"valor_total"`
	Desconto        float64           `json:"desconto"`
	FormaPgto       string            `json:"forma_pgto"`
	ClienteID       *string           `json:"cliente_id"`
	ConsumidorFinal bool              `json:"consumidor_final"`
	DocFiscal       string            `json:"doc_fiscal"`
	Status          string            `json:"status"`
	DocFiscalNumero string            `json:"doc_fiscal_numero,omitempty"`
	Itens           []detalheResponse `json:"itens"`
	CriadoEm        time.Time         `json:"criado_em"`
	AtualizadoEm    time.Time         `json:"atualizado_em"`
}

func toVendaResponse(v *domain.Venda) vendaResponse {
	resp := vendaResponse{
		ID:              v.ID.String(),
		DtVenda:         v.DtVenda,
		ValorTotal:      v.ValorTotal,
		Desconto:        v.Desconto,
		FormaPgto:       string(v.FormaPgto),
		ConsumidorFinal: v.ConsumidorFinal,
		DocFiscal:       string(v.DocFiscal),
		Status:          string(v.Status),
		DocFiscalNumero: v.DocFiscalNumero,
		CriadoEm:        v.CriadoEm,
		AtualizadoEm:    v.AtualizadoEm,
	}
	if v.ClienteID != nil {
		s := v.ClienteID.String()
		resp.ClienteID = &s
	}
	for _, item := range v.Itens {
		resp.Itens = append(resp.Itens, detalheResponse{
			ID:            item.ID.String(),
			ProdutoID:     item.ProdutoID.String(),
			Quantidade:    item.Quantidade,
			PrecoUnitario: item.PrecoUnitario,
		})
	}
	if resp.Itens == nil {
		resp.Itens = []detalheResponse{}
	}
	return resp
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// CriarVenda cria uma nova venda em rascunho.
// POST /vendas
func (h *Handler) CriarVenda(w http.ResponseWriter, r *http.Request) {
	var req criarVendaRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "REQUISICAO_INVALIDA", err.Error())
		return
	}

	in := ports.CriarVendaInput{
		ConsumidorFinal: req.ConsumidorFinal,
		FormaPgto:       domain.FormaPgto(req.FormaPgto),
		DocFiscal:       domain.DocFiscal(req.DocFiscal),
		Desconto:        req.Desconto,
	}

	if req.ClienteID != "" {
		cliID, err := uuid.Parse(req.ClienteID)
		if err != nil {
			httpserver.Error(w, http.StatusBadRequest, "CLIENTE_ID_INVALIDO", "cliente_id inválido")
			return
		}
		in.ClienteID = &cliID
	}

	for _, item := range req.Itens {
		prodID, err := uuid.Parse(item.ProdutoID)
		if err != nil {
			httpserver.Error(w, http.StatusBadRequest, "PRODUTO_ID_INVALIDO", "produto_id inválido")
			return
		}
		in.Itens = append(in.Itens, ports.CriarDetalheInput{
			ProdutoID:     prodID,
			Quantidade:    item.Quantidade,
			PrecoUnitario: item.PrecoUnitario,
		})
	}

	v, err := h.svc.CriarVenda(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, toVendaResponse(v))
}

// ConfirmarVenda transiciona a venda para CONFIRMADA.
// POST /vendas/{id}/confirmar
func (h *Handler) ConfirmarVenda(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "ID_INVALIDO", "id inválido")
		return
	}

	claims := auth.FromContext(r.Context())
	respID, err := uuid.Parse(claims.Subject)
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "ERRO_INTERNO", "erro ao identificar responsável")
		return
	}

	v, err := h.svc.ConfirmarVenda(r.Context(), id, respID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toVendaResponse(v))
}

// ListarVendas retorna a lista de vendas com paginação.
// GET /vendas
func (h *Handler) ListarVendas(w http.ResponseWriter, r *http.Request) {
	var clienteID *uuid.UUID
	if s := r.URL.Query().Get("cliente_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			httpserver.Error(w, http.StatusBadRequest, "CLIENTE_ID_INVALIDO", "cliente_id inválido")
			return
		}
		clienteID = &id
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	vendas, err := h.svc.ListarVendas(r.Context(), clienteID, limit, offset)
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "ERRO_INTERNO", err.Error())
		return
	}

	resp := make([]vendaResponse, len(vendas))
	for i, v := range vendas {
		v := v
		resp[i] = toVendaResponse(&v)
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": resp})
}

// BuscarVenda retorna uma venda completa pelo ID.
// GET /vendas/{id}
func (h *Handler) BuscarVenda(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "ID_INVALIDO", "id inválido")
		return
	}

	v, err := h.svc.BuscarVenda(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toVendaResponse(v))
}

// ─── Tradução de erros de domínio ─────────────────────────────────────────────

func writeDomainError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrVendaNaoEncontrada:
		httpserver.Error(w, http.StatusNotFound, "NAO_ENCONTRADA", err.Error())
	case domain.ErrSaldoInsuficiente:
		httpserver.Error(w, http.StatusUnprocessableEntity, "SALDO_INSUFICIENTE", err.Error())
	case domain.ErrVendaJaConfirmada,
		domain.ErrVendaStatusInvalido,
		domain.ErrVendaVazia,
		domain.ErrClienteOuConsumidorFinal,
		domain.ErrFormaPgtoObrigatoria,
		domain.ErrQuantidadePositiva,
		domain.ErrPrecoNaoPositivo,
		domain.ErrDescontoNaoNegativo,
		domain.ErrDescontoMaiorQueTotal:
		httpserver.Error(w, http.StatusUnprocessableEntity, "REGRA_VIOLADA", err.Error())
	default:
		httpserver.Error(w, http.StatusInternalServerError, "ERRO_INTERNO", err.Error())
	}
}
