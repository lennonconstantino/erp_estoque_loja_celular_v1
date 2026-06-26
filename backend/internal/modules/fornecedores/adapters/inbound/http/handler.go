// Package http é o adaptador de entrada do contexto fornecedores: traduz HTTP
// (JSON) ↔ casos de uso (ports.FornecedorService).
package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

// Handler agrupa os endpoints do módulo.
type Handler struct {
	svc ports.FornecedorService
}

// NewHandler cria o handler a partir da porta de entrada.
func NewHandler(svc ports.FornecedorService) *Handler {
	return &Handler{svc: svc}
}

// --- DTOs de transporte (JSON) -----------------------------------------------

type fornecedorRequest struct {
	CNPJ         string `json:"cnpj"`
	RazaoSocial  string `json:"razao_social"`
	NomeFantasia string `json:"nome_fantasia"`
	Email        string `json:"email"`
	Telefone1    string `json:"telefone1"`
	Telefone2    string `json:"telefone2,omitempty"`
	CEP          string `json:"cep,omitempty"`
	Numero       string `json:"numero,omitempty"`
	Complemento  string `json:"complemento,omitempty"`
	Rua          string `json:"rua,omitempty"`
	Bairro       string `json:"bairro,omitempty"`
	Cidade       string `json:"cidade,omitempty"`
	UF           string `json:"uf,omitempty"`
	Comercial    string `json:"comercial"`
	Financeiro   string `json:"financeiro,omitempty"`
	Ativo        *bool  `json:"ativo,omitempty"`
}

type fornecedorResponse struct {
	ID           uuid.UUID  `json:"id"`
	CNPJ         string     `json:"cnpj"`
	RazaoSocial  string     `json:"razao_social"`
	NomeFantasia string     `json:"nome_fantasia"`
	Email        string     `json:"email"`
	Telefone1    string     `json:"telefone1"`
	Telefone2    string     `json:"telefone2,omitempty"`
	CEP          string     `json:"cep,omitempty"`
	Rua          string     `json:"rua,omitempty"`
	Numero       string     `json:"numero,omitempty"`
	Complemento  string     `json:"complemento,omitempty"`
	Bairro       string     `json:"bairro,omitempty"`
	Cidade       string     `json:"cidade,omitempty"`
	UF           string     `json:"uf,omitempty"`
	Comercial    string     `json:"comercial"`
	Financeiro   string     `json:"financeiro,omitempty"`
	UltimaCompra *time.Time `json:"ultima_compra,omitempty"`
	Ativo        bool       `json:"ativo"`
}

func toResponse(f *domain.Fornecedor) fornecedorResponse {
	return fornecedorResponse{
		ID: f.ID, CNPJ: f.CNPJ,
		RazaoSocial: f.RazaoSocial, NomeFantasia: f.NomeFantasia,
		Email: f.Email, Telefone1: f.Telefone1, Telefone2: f.Telefone2,
		CEP: f.CEP, Rua: f.Rua, Numero: f.Numero, Complemento: f.Complemento,
		Bairro: f.Bairro, Cidade: f.Cidade, UF: f.UF,
		Comercial: f.Comercial, Financeiro: f.Financeiro,
		UltimaCompra: f.UltimaCompra, Ativo: f.Ativo,
	}
}

// --- Endpoints ---------------------------------------------------------------

// Criar cria um fornecedor. POST /
func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	var req fornecedorRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	f, err := h.svc.Criar(r.Context(), ports.CriarFornecedorInput{
		CNPJ: req.CNPJ, RazaoSocial: req.RazaoSocial, NomeFantasia: req.NomeFantasia,
		Email: req.Email, Telefone1: req.Telefone1, Telefone2: req.Telefone2,
		CEP: req.CEP, Numero: req.Numero, Complemento: req.Complemento,
		Rua: req.Rua, Bairro: req.Bairro, Cidade: req.Cidade, UF: req.UF,
		Comercial: req.Comercial, Financeiro: req.Financeiro,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, toResponse(f))
}

// Atualizar altera um fornecedor. PUT /{id}
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	var req fornecedorRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	f, err := h.svc.Atualizar(r.Context(), id, ports.AtualizarFornecedorInput{
		RazaoSocial: req.RazaoSocial, NomeFantasia: req.NomeFantasia,
		Email: req.Email, Telefone1: req.Telefone1, Telefone2: req.Telefone2,
		CEP: req.CEP, Numero: req.Numero, Complemento: req.Complemento,
		Rua: req.Rua, Bairro: req.Bairro, Cidade: req.Cidade, UF: req.UF,
		Comercial: req.Comercial, Financeiro: req.Financeiro,
		Ativo: req.Ativo,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toResponse(f))
}

// BuscarPorID retorna um fornecedor. GET /{id}
func (h *Handler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	f, err := h.svc.BuscarPorID(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toResponse(f))
}

// BuscarPorCNPJ consulta por CNPJ. GET /by-cnpj/{cnpj}
func (h *Handler) BuscarPorCNPJ(w http.ResponseWriter, r *http.Request) {
	f, err := h.svc.BuscarPorCNPJ(r.Context(), chi.URLParam(r, "cnpj"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toResponse(f))
}

// Listar lista/pesquisa fornecedores. GET /?q=&limit=&offset=
func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	fs, err := h.svc.Listar(r.Context(), q, limit, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]fornecedorResponse, 0, len(fs))
	for i := range fs {
		out = append(out, toResponse(&fs[i]))
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": out})
}

// ConsultarCEP faz proxy ao CepGateway. GET /cep/{cep}
func (h *Handler) ConsultarCEP(w http.ResponseWriter, r *http.Request) {
	end, err := h.svc.ConsultarCEP(r.Context(), chi.URLParam(r, "cep"))
	if err != nil {
		httpserver.Error(w, http.StatusBadGateway, "CEP_LOOKUP_FAILED", err.Error())
		return
	}
	httpserver.JSON(w, http.StatusOK, end)
}

// writeDomainError traduz erros de domínio em respostas HTTP.
func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNaoEncontrado):
		httpserver.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrCNPJJaCadastrado):
		httpserver.Error(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, domain.ErrCNPJInvalido),
		errors.Is(err, domain.ErrRazaoSocialObrigatoria),
		errors.Is(err, domain.ErrNomeFantasiaObrigatorio),
		errors.Is(err, domain.ErrEmailInvalido),
		errors.Is(err, domain.ErrTelefone1Obrigatorio),
		errors.Is(err, domain.ErrComercialObrigatorio),
		errors.Is(err, domain.ErrUFInvalida):
		httpserver.Error(w, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
	default:
		httpserver.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
