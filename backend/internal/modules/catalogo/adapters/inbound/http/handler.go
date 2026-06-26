// Package http é o adaptador de entrada do contexto catálogo: traduz HTTP
// (JSON) ↔ casos de uso (CategoriaService e ProdutoService).
package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

// Handler agrupa os endpoints do módulo.
type Handler struct {
	cats  ports.CategoriaService
	prods ports.ProdutoService
}

// NewHandler cria o handler a partir das portas de entrada.
func NewHandler(cats ports.CategoriaService, prods ports.ProdutoService) *Handler {
	return &Handler{cats: cats, prods: prods}
}

// ─── DTOs de Categoria ───────────────────────────────────────────────────────

type categoriaRequest struct {
	Descricao string `json:"descricao"`
}

type categoriaResponse struct {
	ID        uuid.UUID `json:"id"`
	Descricao string    `json:"descricao"`
}

func toCategoriaResponse(c *domain.Categoria) categoriaResponse {
	return categoriaResponse{ID: c.ID, Descricao: c.Descricao}
}

// ─── DTOs de Produto ─────────────────────────────────────────────────────────

type produtoRequest struct {
	CategoriaID   string  `json:"categoria_id"`
	Descricao     string  `json:"descricao"`
	PrecoCusto    float64 `json:"preco_custo"`
	PrecoVenda    float64 `json:"preco_venda"`
	EstoqueMinimo int     `json:"estoque_minimo"`
	GarantiaMeses int     `json:"garantia_meses"`
	Modelo        string  `json:"modelo"`
	Ativo         *bool   `json:"ativo,omitempty"`
}

type produtoResponse struct {
	ID            uuid.UUID `json:"id"`
	CategoriaID   uuid.UUID `json:"categoria_id"`
	Descricao     string    `json:"descricao"`
	PrecoCusto    float64   `json:"preco_custo"`
	PrecoVenda    float64   `json:"preco_venda"`
	MargemPct     float64   `json:"margem_pct"`
	EstoqueMinimo int       `json:"estoque_minimo"`
	EstoqueAtual  int       `json:"estoque_atual"`
	GarantiaMeses int       `json:"garantia_meses"`
	Modelo        string    `json:"modelo,omitempty"`
	Disponivel    bool      `json:"disponivel"`
	Ativo         bool      `json:"ativo"`
}

func toProdutoResponse(p *domain.Produto) produtoResponse {
	return produtoResponse{
		ID:            p.ID,
		CategoriaID:   p.CategoriaID,
		Descricao:     p.Descricao,
		PrecoCusto:    p.PrecoCusto,
		PrecoVenda:    p.PrecoVenda,
		MargemPct:     p.Margem(),
		EstoqueMinimo: p.EstoqueMinimo,
		EstoqueAtual:  p.EstoqueAtual,
		GarantiaMeses: p.GarantiaMeses,
		Modelo:        p.Modelo,
		Disponivel:    p.Disponivel,
		Ativo:         p.Ativo,
	}
}

// ─── Handlers de Categoria ───────────────────────────────────────────────────

// CriarCategoria cria uma categoria. POST /categorias/
func (h *Handler) CriarCategoria(w http.ResponseWriter, r *http.Request) {
	var req categoriaRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	c, err := h.cats.CriarCategoria(r.Context(), ports.CriarCategoriaInput{Descricao: req.Descricao})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, toCategoriaResponse(c))
}

// AtualizarCategoria altera uma categoria. PUT /categorias/{id}
func (h *Handler) AtualizarCategoria(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	var req categoriaRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	c, err := h.cats.AtualizarCategoria(r.Context(), id, ports.AtualizarCategoriaInput{Descricao: req.Descricao})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toCategoriaResponse(c))
}

// RemoverCategoria exclui uma categoria. DELETE /categorias/{id}
func (h *Handler) RemoverCategoria(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	if err := h.cats.RemoverCategoria(r.Context(), id); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// BuscarCategoriaPorID retorna uma categoria. GET /categorias/{id}
func (h *Handler) BuscarCategoriaPorID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	c, err := h.cats.BuscarCategoriaPorID(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toCategoriaResponse(c))
}

// ListarCategorias lista categorias. GET /categorias/?q=&limit=&offset=
func (h *Handler) ListarCategorias(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	cs, err := h.cats.ListarCategorias(r.Context(), q, limit, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]categoriaResponse, 0, len(cs))
	for i := range cs {
		out = append(out, toCategoriaResponse(&cs[i]))
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": out})
}

// ─── Handlers de Produto ─────────────────────────────────────────────────────

// CriarProduto cria um produto. POST /produtos/
func (h *Handler) CriarProduto(w http.ResponseWriter, r *http.Request) {
	var req produtoRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	catID, err := uuid.Parse(req.CategoriaID)
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "categoria_id inválido")
		return
	}
	p, err := h.prods.CriarProduto(r.Context(), ports.CriarProdutoInput{
		CategoriaID:   catID,
		Descricao:     req.Descricao,
		PrecoCusto:    req.PrecoCusto,
		PrecoVenda:    req.PrecoVenda,
		EstoqueMinimo: req.EstoqueMinimo,
		GarantiaMeses: req.GarantiaMeses,
		Modelo:        req.Modelo,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, toProdutoResponse(p))
}

// AtualizarProduto altera um produto. PUT /produtos/{id}
func (h *Handler) AtualizarProduto(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	var req produtoRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	catID, err := uuid.Parse(req.CategoriaID)
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "categoria_id inválido")
		return
	}
	p, err := h.prods.AtualizarProduto(r.Context(), id, ports.AtualizarProdutoInput{
		Descricao:     req.Descricao,
		CategoriaID:   catID,
		PrecoCusto:    req.PrecoCusto,
		PrecoVenda:    req.PrecoVenda,
		EstoqueMinimo: req.EstoqueMinimo,
		GarantiaMeses: req.GarantiaMeses,
		Modelo:        req.Modelo,
		Ativo:         req.Ativo,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toProdutoResponse(p))
}

// BuscarProdutoPorID retorna um produto. GET /produtos/{id}
func (h *Handler) BuscarProdutoPorID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	p, err := h.prods.BuscarProdutoPorID(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toProdutoResponse(p))
}

// ListarProdutos lista produtos. GET /produtos/?q=&categoria_id=&limit=&offset=
func (h *Handler) ListarProdutos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var catID *uuid.UUID
	if raw := r.URL.Query().Get("categoria_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "categoria_id inválido")
			return
		}
		catID = &id
	}

	ps, err := h.prods.ListarProdutos(r.Context(), q, catID, limit, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]produtoResponse, 0, len(ps))
	for i := range ps {
		out = append(out, toProdutoResponse(&ps[i]))
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": out})
}

// ─── Mapeamento de erros de domínio ──────────────────────────────────────────

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrCategoriaNaoEncontrada),
		errors.Is(err, domain.ErrProdutoNaoEncontrado):
		httpserver.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrCategoriaJaCadastrada):
		httpserver.Error(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, domain.ErrCategoriaComProdutos):
		httpserver.Error(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, domain.ErrDescricaoCatObrigatoria),
		errors.Is(err, domain.ErrDescricaoObrigatoria),
		errors.Is(err, domain.ErrCategoriaObrigatoria),
		errors.Is(err, domain.ErrPrecoNaoPositivo),
		errors.Is(err, domain.ErrPrecoInvalido),
		errors.Is(err, domain.ErrEstoqueMinNaoNeg),
		errors.Is(err, domain.ErrEstoqueAtualNaoNeg):
		httpserver.Error(w, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
	default:
		httpserver.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
