// Package http expõe o módulo compras via HTTP.
package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

// Handler agrupa os handlers HTTP do módulo compras.
type Handler struct {
	svc ports.CompraService
}

// NewHandler cria o handler injetando o serviço.
func NewHandler(svc ports.CompraService) *Handler {
	return &Handler{svc: svc}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type detalheRequest struct {
	ProdutoID   string  `json:"produto_id"`
	Quantidade  int     `json:"quantidade"`
	PrecoCompra float64 `json:"preco_compra"`
	PrecoVenda  float64 `json:"preco_venda"`
}

type criarCompraRequest struct {
	FornecedorID string           `json:"fornecedor_id"`
	NF           string           `json:"nf"`
	DtCompra     string           `json:"dt_compra"` // formato "2006-01-02"
	Itens        []detalheRequest `json:"itens"`
}

type detalheResponse struct {
	ID          string  `json:"id"`
	ProdutoID   string  `json:"produto_id"`
	Quantidade  int     `json:"quantidade"`
	PrecoCompra float64 `json:"preco_compra"`
	PrecoVenda  float64 `json:"preco_venda"`
	Margem      float64 `json:"margem"`
}

type compraResponse struct {
	ID           string            `json:"id"`
	FornecedorID string            `json:"fornecedor_id"`
	NF           string            `json:"nf"`
	DtCompra     string            `json:"dt_compra"`
	ValorTotal   float64           `json:"valor_total"`
	Status       string            `json:"status"`
	Itens        []detalheResponse `json:"itens"`
	CriadoEm    time.Time         `json:"criado_em"`
	AtualizadoEm time.Time        `json:"atualizado_em"`
}

func toCompraResponse(c *domain.Compra) compraResponse {
	resp := compraResponse{
		ID:           c.ID.String(),
		FornecedorID: c.FornecedorID.String(),
		NF:           c.NF,
		DtCompra:     c.DtCompra.Format("2006-01-02"),
		ValorTotal:   c.ValorTotal,
		Status:       string(c.Status),
		CriadoEm:    c.CriadoEm,
		AtualizadoEm: c.AtualizadoEm,
	}
	for _, item := range c.Itens {
		resp.Itens = append(resp.Itens, detalheResponse{
			ID:          item.ID.String(),
			ProdutoID:   item.ProdutoID.String(),
			Quantidade:  item.Quantidade,
			PrecoCompra: item.PrecoCompra,
			PrecoVenda:  item.PrecoVenda,
			Margem:      item.Margem,
		})
	}
	if resp.Itens == nil {
		resp.Itens = []detalheResponse{}
	}
	return resp
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// CriarCompra cria uma nova compra em rascunho.
// POST /compras
func (h *Handler) CriarCompra(w http.ResponseWriter, r *http.Request) {
	var req criarCompraRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "REQUISICAO_INVALIDA", err.Error())
		return
	}

	fornID, err := uuid.Parse(req.FornecedorID)
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "FORNECEDOR_ID_INVALIDO", "fornecedor_id inválido")
		return
	}

	dtCompra, err := time.Parse("2006-01-02", req.DtCompra)
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "DATA_INVALIDA", "dt_compra deve estar no formato YYYY-MM-DD")
		return
	}

	in := ports.CriarCompraInput{
		FornecedorID: fornID,
		NF:           req.NF,
		DtCompra:     dtCompra,
	}
	for _, item := range req.Itens {
		prodID, err := uuid.Parse(item.ProdutoID)
		if err != nil {
			httpserver.Error(w, http.StatusBadRequest, "PRODUTO_ID_INVALIDO", "produto_id inválido")
			return
		}
		in.Itens = append(in.Itens, ports.CriarDetalheInput{
			ProdutoID:   prodID,
			Quantidade:  item.Quantidade,
			PrecoCompra: item.PrecoCompra,
			PrecoVenda:  item.PrecoVenda,
		})
	}

	c, err := h.svc.CriarCompra(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, toCompraResponse(c))
}

// ConfirmarCompra transiciona a compra para CONFIRMADA.
// POST /compras/{id}/confirmar
func (h *Handler) ConfirmarCompra(w http.ResponseWriter, r *http.Request) {
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

	c, err := h.svc.ConfirmarCompra(r.Context(), id, respID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toCompraResponse(c))
}

// ListarCompras retorna a lista de compras com paginação.
// GET /compras
func (h *Handler) ListarCompras(w http.ResponseWriter, r *http.Request) {
	var fornID *uuid.UUID
	if s := r.URL.Query().Get("fornecedor_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			httpserver.Error(w, http.StatusBadRequest, "FORNECEDOR_ID_INVALIDO", "fornecedor_id inválido")
			return
		}
		fornID = &id
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	compras, err := h.svc.ListarCompras(r.Context(), fornID, limit, offset)
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "ERRO_INTERNO", err.Error())
		return
	}

	resp := make([]compraResponse, len(compras))
	for i, c := range compras {
		c := c
		resp[i] = toCompraResponse(&c)
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": resp})
}

// BuscarCompra retorna uma compra completa pelo ID.
// GET /compras/{id}
func (h *Handler) BuscarCompra(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "ID_INVALIDO", "id inválido")
		return
	}

	c, err := h.svc.BuscarCompra(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toCompraResponse(c))
}

// ─── Tradução de erros de domínio ─────────────────────────────────────────────

func writeDomainError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrCompraNaoEncontrada:
		httpserver.Error(w, http.StatusNotFound, "NAO_ENCONTRADA", err.Error())
	case domain.ErrCompraJaConfirmada,
		domain.ErrCompraStatusInvalido,
		domain.ErrCompraVazia,
		domain.ErrFornecedorObrigatorio,
		domain.ErrDataCompraObrigatoria,
		domain.ErrQuantidadePositiva,
		domain.ErrPrecoNaoPositivo,
		domain.ErrPrecoInvalido:
		httpserver.Error(w, http.StatusUnprocessableEntity, "REGRA_VIOLADA", err.Error())
	default:
		httpserver.Error(w, http.StatusInternalServerError, "ERRO_INTERNO", err.Error())
	}
}
