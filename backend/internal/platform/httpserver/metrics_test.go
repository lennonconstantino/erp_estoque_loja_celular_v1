package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("metrics"))
	})
}

func TestProtectMetrics(t *testing.T) {
	casos := []struct {
		nome   string
		token  string
		prod   bool
		header string
		quer   int
	}{
		{"dev sem token: liberado", "", false, "", http.StatusOK},
		{"prod sem token: fechado (404)", "", true, "", http.StatusNotFound},
		{"token + bearer correto: liberado", "s3cr3t", true, "Bearer s3cr3t", http.StatusOK},
		{"token + bearer errado: 401", "s3cr3t", true, "Bearer errado", http.StatusUnauthorized},
		{"token + sem header: 401", "s3cr3t", true, "", http.StatusUnauthorized},
		{"token + header sem prefixo Bearer: 401", "s3cr3t", true, "s3cr3t", http.StatusUnauthorized},
		{"token em dev também exige bearer", "s3cr3t", false, "", http.StatusUnauthorized},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			h := httpserver.ProtectMetrics(okHandler(), c.token, c.prod)
			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			if c.header != "" {
				req.Header.Set("Authorization", c.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != c.quer {
				t.Fatalf("status = %d, quer %d", rec.Code, c.quer)
			}
		})
	}
}
