// Package http é o adaptador de entrada do contexto clientes: traduz HTTP
// (JSON) ↔ casos de uso (ports.ClienteService).
package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

// Handler agrupa os endpoints do módulo.
type Handler struct {
	svc ports.ClienteService
}

// NewHandler cria o handler a partir da porta de entrada.
func NewHandler(svc ports.ClienteService) *Handler {
	return &Handler{svc: svc}
}

// --- DTOs de transporte (JSON) -----------------------------------------------

type clienteRequest struct {
	CPF         string `json:"cpf"`
	Nome        string `json:"nome"`
	Email       string `json:"email"`
	Telefone    string `json:"telefone"`
	CEP         string `json:"cep"`
	Numero      string `json:"numero"`
	Complemento string `json:"complemento"`
	Rua         string `json:"rua"`
	Bairro      string `json:"bairro"`
	Cidade      string `json:"cidade"`
	UF          string `json:"uf"`
	// Ativo só é consumido na atualização (PUT); na criação é ignorado, pois todo
	// cliente nasce ativo (domain.NovoCliente). Declarado no DTO compartilhado para
	// não ser rejeitado como campo desconhecido pelo DecodeJSON.
	Ativo bool `json:"ativo"`
}

type clienteResponse struct {
	ID           uuid.UUID  `json:"id"`
	CPF          string     `json:"cpf"`
	Nome         string     `json:"nome"`
	Email        string     `json:"email"`
	Telefone     string     `json:"telefone,omitempty"`
	CEP          string     `json:"cep,omitempty"`
	Rua          string     `json:"rua,omitempty"`
	Numero       string     `json:"numero,omitempty"`
	Complemento  string     `json:"complemento,omitempty"`
	Bairro       string     `json:"bairro,omitempty"`
	Cidade       string     `json:"cidade,omitempty"`
	UF           string     `json:"uf,omitempty"`
	UltimaCompra *time.Time `json:"ultima_compra,omitempty"`
	Ativo        bool       `json:"ativo"`
}

func toResponse(c *domain.Cliente) clienteResponse {
	return clienteResponse{
		ID: c.ID, CPF: c.CPF, Nome: c.Nome, Email: c.Email, Telefone: c.Telefone,
		CEP: c.CEP, Rua: c.Rua, Numero: c.Numero, Complemento: c.Complemento,
		Bairro: c.Bairro, Cidade: c.Cidade, UF: c.UF,
		UltimaCompra: c.UltimaCompra, Ativo: c.Ativo,
	}
}

// --- Endpoints ---------------------------------------------------------------

// Criar cria um cliente. POST /
func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	var req clienteRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	c, err := h.svc.Criar(r.Context(), ports.CriarClienteInput{
		CPF: req.CPF, Nome: req.Nome, Email: req.Email, Telefone: req.Telefone,
		CEP: req.CEP, Numero: req.Numero, Complemento: req.Complemento,
		Rua: req.Rua, Bairro: req.Bairro, Cidade: req.Cidade, UF: req.UF,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, toResponse(c))
}

// Atualizar altera um cliente. PUT /{id}
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	var req clienteRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	c, err := h.svc.Atualizar(r.Context(), id, ports.AtualizarClienteInput{
		Nome: req.Nome, Email: req.Email, Telefone: req.Telefone,
		CEP: req.CEP, Numero: req.Numero, Complemento: req.Complemento,
		Rua: req.Rua, Bairro: req.Bairro, Cidade: req.Cidade, UF: req.UF,
		Ativo: req.Ativo,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toResponse(c))
}

// Remover exclui um cliente. DELETE /{id}
func (h *Handler) Remover(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	if err := h.svc.Remover(r.Context(), id); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// BuscarPorID retorna um cliente. GET /{id}
func (h *Handler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	c, err := h.svc.BuscarPorID(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toResponse(c))
}

// BuscarPorCPF consulta por CPF (usado antes de cadastrar). GET /by-cpf/{cpf}
func (h *Handler) BuscarPorCPF(w http.ResponseWriter, r *http.Request) {
	c, err := h.svc.BuscarPorCPF(r.Context(), chi.URLParam(r, "cpf"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toResponse(c))
}

// Listar lista/pesquisa clientes. GET /?q=&limit=&offset=
func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	cs, err := h.svc.Listar(r.Context(), q, limit, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]clienteResponse, 0, len(cs))
	for i := range cs {
		out = append(out, toResponse(&cs[i]))
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
	case errors.Is(err, domain.ErrCPFJaCadastrado):
		httpserver.Error(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, domain.ErrCPFInvalido),
		errors.Is(err, domain.ErrNomeObrigatorio),
		errors.Is(err, domain.ErrEmailInvalido),
		errors.Is(err, domain.ErrUFInvalida):
		httpserver.Error(w, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
	default:
		httpserver.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
