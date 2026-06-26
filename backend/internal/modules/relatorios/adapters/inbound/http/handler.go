// Package http expõe os endpoints de relatório via HTTP.
package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/relatorios/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

// Handler agrupa os handlers de relatório.
type Handler struct {
	svc ports.RelatorioService
}

// NewHandler cria um Handler com o serviço injetado.
func NewHandler(svc ports.RelatorioService) *Handler {
	return &Handler{svc: svc}
}

// ProdutosAbaixoDoMinimo godoc: GET /relatorios/produtos/abaixo-do-minimo
func (h *Handler) ProdutosAbaixoDoMinimo(w http.ResponseWriter, r *http.Request) {
	lista, err := h.svc.ProdutosAbaixoDoMinimo(r.Context())
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "ERRO_INTERNO", err.Error())
		return
	}
	type item struct {
		ID            string `json:"id"`
		Descricao     string `json:"descricao"`
		EstoqueAtual  int    `json:"estoque_atual"`
		EstoqueMinimo int    `json:"estoque_minimo"`
		Defasagem     int    `json:"defasagem"`
	}
	resp := make([]item, 0, len(lista))
	for _, p := range lista {
		resp = append(resp, item{
			ID:            p.ID.String(),
			Descricao:     p.Descricao,
			EstoqueAtual:  p.EstoqueAtual,
			EstoqueMinimo: p.EstoqueMinimo,
			Defasagem:     p.Defasagem,
		})
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": resp, "total": len(resp)})
}

// MaisVendidos godoc: GET /relatorios/produtos/mais-vendidos?de=&ate=&limite=
func (h *Handler) MaisVendidos(w http.ResponseWriter, r *http.Request) {
	de, ate, limite, ok := parsePeriodoLimite(w, r)
	if !ok {
		return
	}
	lista, err := h.svc.MaisVendidos(r.Context(), de, ate, limite)
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "ERRO_INTERNO", err.Error())
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": toProdutoVendidoResp(lista), "total": len(lista)})
}

// MenosVendidos godoc: GET /relatorios/produtos/menos-vendidos?de=&ate=&limite=
func (h *Handler) MenosVendidos(w http.ResponseWriter, r *http.Request) {
	de, ate, limite, ok := parsePeriodoLimite(w, r)
	if !ok {
		return
	}
	lista, err := h.svc.MenosVendidos(r.Context(), de, ate, limite)
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "ERRO_INTERNO", err.Error())
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": toProdutoVendidoResp(lista), "total": len(lista)})
}

// ResumoVendas godoc: GET /relatorios/vendas?de=&ate=
func (h *Handler) ResumoVendas(w http.ResponseWriter, r *http.Request) {
	de, ate, ok := parsePeriodo(w, r)
	if !ok {
		return
	}
	resumo, err := h.svc.ResumoVendas(r.Context(), de, ate)
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "ERRO_INTERNO", err.Error())
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{
		"total_vendas":  resumo.TotalVendas,
		"valor_total":   resumo.ValorTotal,
		"ticket_medio":  resumo.TicketMedio,
		"de":            resumo.De.Format(time.DateOnly),
		"ate":           resumo.Ate.Format(time.DateOnly),
	})
}

// ResumoCompras godoc: GET /relatorios/compras?de=&ate=
func (h *Handler) ResumoCompras(w http.ResponseWriter, r *http.Request) {
	de, ate, ok := parsePeriodo(w, r)
	if !ok {
		return
	}
	resumo, err := h.svc.ResumoCompras(r.Context(), de, ate)
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "ERRO_INTERNO", err.Error())
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{
		"total_compras": resumo.TotalCompras,
		"valor_total":   resumo.ValorTotal,
		"de":            resumo.De.Format(time.DateOnly),
		"ate":           resumo.Ate.Format(time.DateOnly),
	})
}

// --- helpers de parse ---

// parsePeriodo lê os query params `de` e `ate` (formato YYYY-MM-DD).
// Se ausentes, usa o mês corrente (1º dia a hoje).
func parsePeriodo(w http.ResponseWriter, r *http.Request) (de, ate time.Time, ok bool) {
	now := time.Now().UTC()
	deStr := r.URL.Query().Get("de")
	ateStr := r.URL.Query().Get("ate")

	if deStr == "" {
		de = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	} else {
		var err error
		de, err = time.Parse(time.DateOnly, deStr)
		if err != nil {
			httpserver.Error(w, http.StatusBadRequest, "PARAMETRO_INVALIDO", "parâmetro 'de' deve estar no formato YYYY-MM-DD")
			return time.Time{}, time.Time{}, false
		}
	}
	if ateStr == "" {
		ate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	} else {
		var err error
		ate, err = time.Parse(time.DateOnly, ateStr)
		if err != nil {
			httpserver.Error(w, http.StatusBadRequest, "PARAMETRO_INVALIDO", "parâmetro 'ate' deve estar no formato YYYY-MM-DD")
			return time.Time{}, time.Time{}, false
		}
		// inclui o dia inteiro
		ate = time.Date(ate.Year(), ate.Month(), ate.Day(), 23, 59, 59, 0, time.UTC)
	}
	return de, ate, true
}

func parsePeriodoLimite(w http.ResponseWriter, r *http.Request) (de, ate time.Time, limite int, ok bool) {
	de, ate, ok = parsePeriodo(w, r)
	if !ok {
		return
	}
	limite = 10
	if ls := r.URL.Query().Get("limite"); ls != "" {
		n, err := strconv.Atoi(ls)
		if err != nil || n <= 0 {
			httpserver.Error(w, http.StatusBadRequest, "PARAMETRO_INVALIDO", "parâmetro 'limite' deve ser um inteiro positivo")
			return time.Time{}, time.Time{}, 0, false
		}
		limite = n
	}
	return de, ate, limite, true
}

type produtoVendidoResp struct {
	ProdutoID    string  `json:"produto_id"`
	Descricao    string  `json:"descricao"`
	TotalVendido int     `json:"total_vendido"`
	TotalValor   float64 `json:"total_valor"`
}

func toProdutoVendidoResp(lista []domain.ProdutoVendido) []produtoVendidoResp {
	resp := make([]produtoVendidoResp, 0, len(lista))
	for _, p := range lista {
		resp = append(resp, produtoVendidoResp{
			ProdutoID:    p.ProdutoID.String(),
			Descricao:    p.Descricao,
			TotalVendido: p.TotalVendido,
			TotalValor:   p.TotalValor,
		})
	}
	return resp
}
