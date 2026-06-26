// Package http é o adaptador de entrada do contexto IAM: traduz HTTP ↔ casos de uso.
package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam/ports"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

// Handler agrupa os endpoints do módulo IAM.
type Handler struct {
	svc ports.AuthService
}

// NewHandler cria o handler.
func NewHandler(svc ports.AuthService) *Handler {
	return &Handler{svc: svc}
}

// --- DTOs de transporte (JSON) -----------------------------------------------

type loginRequest struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type usuarioRequest struct {
	Nome   string   `json:"nome"`
	Email  string   `json:"email"`
	Senha  string   `json:"senha"`
	Ativo  *bool    `json:"ativo"`
	Papeis []string `json:"papeis"`
}

type usuarioResponse struct {
	ID        uuid.UUID  `json:"id"`
	Nome      string     `json:"nome"`
	Email     string     `json:"email"`
	Ativo     bool       `json:"ativo"`
	UltAcesso *time.Time `json:"ult_acesso,omitempty"`
	CriadoEm time.Time  `json:"criado_em"`
}

func toResponse(u *domain.Usuario) usuarioResponse {
	return usuarioResponse{
		ID:        u.ID,
		Nome:      u.Nome,
		Email:     u.Email,
		Ativo:     u.Ativo,
		UltAcesso: u.UltAcesso,
		CriadoEm:  u.CriadoEm,
	}
}

// --- Endpoints de autenticação -----------------------------------------------

// Login autentica o usuário. POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	out, err := h.svc.Login(r.Context(), ports.LoginInput{Email: req.Email, Senha: req.Senha})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, loginResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken})
}

// Refresh renova o access token. POST /auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	out, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, loginResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken})
}

// Logout revoga o refresh token. POST /auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	_ = h.svc.Logout(r.Context(), req.RefreshToken)
	w.WriteHeader(http.StatusNoContent)
}

// --- Endpoints de gerenciamento de usuários ----------------------------------

// ListarUsuarios lista usuários paginados. GET /usuarios
func (h *Handler) ListarUsuarios(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	us, err := h.svc.ListarUsuarios(r.Context(), limit, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]usuarioResponse, 0, len(us))
	for i := range us {
		out = append(out, toResponse(&us[i]))
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": out})
}

// CriarUsuario cria um usuário. POST /usuarios
func (h *Handler) CriarUsuario(w http.ResponseWriter, r *http.Request) {
	var req usuarioRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	u, err := h.svc.CriarUsuario(r.Context(), ports.CriarUsuarioInput{
		Nome:   req.Nome,
		Email:  req.Email,
		Senha:  req.Senha,
		Papeis: req.Papeis,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, toResponse(u))
}

// AtualizarUsuario altera um usuário. PATCH /usuarios/{id}
func (h *Handler) AtualizarUsuario(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id inválido")
		return
	}
	var req usuarioRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	u, err := h.svc.AtualizarUsuario(r.Context(), id, ports.AtualizarUsuarioInput{
		Nome:  req.Nome,
		Email: req.Email,
		Ativo: req.Ativo,
		Senha: req.Senha,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, toResponse(u))
}

// writeDomainError traduz erros de domínio em respostas HTTP.
func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrCredenciaisInvalidas):
		httpserver.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
	case errors.Is(err, domain.ErrUsuarioInativo):
		httpserver.Error(w, http.StatusForbidden, "USER_INACTIVE", err.Error())
	case errors.Is(err, domain.ErrTokenInvalido):
		httpserver.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", err.Error())
	case errors.Is(err, domain.ErrNaoEncontrado):
		httpserver.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrEmailJaCadastrado):
		httpserver.Error(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, domain.ErrNomeObrigatorio),
		errors.Is(err, domain.ErrEmailInvalido),
		errors.Is(err, domain.ErrSenhaFraca):
		httpserver.Error(w, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
	default:
		httpserver.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
