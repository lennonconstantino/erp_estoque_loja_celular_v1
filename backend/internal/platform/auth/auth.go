// Package auth implementa emissão/validação de JWT (HS256) e o middleware de
// autorização baseada em permissões (RBAC) usado por todos os módulos.
package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

// Claims são as reivindicações do access token.
type Claims struct {
	Roles []string `json:"roles"`
	Perms []string `json:"perms"`
	jwt.RegisteredClaims
}

// Has indica se o token concede a permissão informada.
func (c *Claims) Has(perm string) bool {
	for _, p := range c.Perms {
		if p == perm {
			return true
		}
	}
	return false
}

type ctxKey struct{}

// FromContext recupera os Claims injetados pelo middleware Authenticate.
func FromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(ctxKey{}).(*Claims)
	return c
}

// Manager emite e valida tokens com um segredo HS256.
type Manager struct {
	secret    []byte
	accessTTL time.Duration
}

// NewManager cria o gerenciador de tokens.
func NewManager(secret string, accessTTL time.Duration) *Manager {
	return &Manager{secret: []byte(secret), accessTTL: accessTTL}
}

// GenerateAccess emite um access token para o usuário com os papéis e
// permissões informados. (Usado pelo módulo iam; útil também em testes.)
func (m *Manager) GenerateAccess(sub string, roles, perms []string) (string, error) {
	now := time.Now()
	claims := Claims{
		Roles: roles,
		Perms: perms,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) parse(token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// Authenticate valida o Bearer token e injeta os Claims no contexto.
func (m *Manager) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			httpserver.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "token ausente")
			return
		}
		claims, err := m.parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			httpserver.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "token inválido")
			return
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePerm exige que o token contenha a permissão informada.
func RequirePerm(perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := FromContext(r.Context())
			if claims == nil || !claims.Has(perm) {
				httpserver.Error(w, http.StatusForbidden, "PERM_DENIED", "permissão negada: "+perm)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
