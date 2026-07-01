package httpserver

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// ProtectMetrics controla o acesso ao endpoint /metrics conforme o ambiente.
// O /metrics fica no mesmo servidor HTTP público da API, então sem proteção ele
// vazaria métricas internas (rotas, latências, uso de memória) para qualquer um.
//
//   - token != "": exige "Authorization: Bearer <token>" (comparação em tempo
//     constante). É o modo de produção com scraping autenticado — ex.: o Grafana
//     Alloy raspando via rede privada do Railway (ver observability/alloy).
//   - token == "" e prod == false: libera. Conveniência de desenvolvimento — o
//     Prometheus local (docker-compose.observability.yml) raspa sem credencial.
//   - token == "" e prod == true: fail-safe. Responde 404 para nunca expor
//     métricas publicamente sem proteção em produção.
func ProtectMetrics(next http.Handler, token string, prod bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case token != "":
			if !bearerAutorizado(r, token) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		case prod:
			// Sem token em produção: endpoint tratado como inexistente.
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// bearerAutorizado valida o header Authorization no formato "Bearer <token>"
// com comparação resistente a timing attacks.
func bearerAutorizado(r *http.Request, token string) bool {
	const prefixo = "Bearer "
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, prefixo) {
		return false
	}
	recebido := strings.TrimPrefix(h, prefixo)
	return subtle.ConstantTimeCompare([]byte(recebido), []byte(token)) == 1
}
