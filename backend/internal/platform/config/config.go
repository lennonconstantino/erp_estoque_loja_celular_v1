// Package config carrega a configuração da aplicação a partir de variáveis
// de ambiente (com suporte a um arquivo .env opcional).
package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config agrega todos os parâmetros de execução do serviço.
type Config struct {
	AppEnv       string
	AppPort      string
	DatabaseURL  string
	JWTSecret    string
	JWTAccessTTL time.Duration
	CepAPIURL    string
}

// Load lê o .env (se existir) e monta a Config a partir do ambiente,
// aplicando defaults para desenvolvimento.
func Load() *Config {
	_ = godotenv.Load() // ignora ausência do arquivo (produção usa env real)

	return &Config{
		AppEnv:       getenv("APP_ENV", "development"),
		AppPort:      getenv("APP_PORT", "8080"),
		DatabaseURL:  getenv("DATABASE_URL", "postgres://erp:erp_secret@localhost:5432/erp_estoque?sslmode=disable"),
		JWTSecret:    getenv("JWT_SECRET", "troque-este-segredo-em-producao"),
		JWTAccessTTL: getdur("JWT_ACCESS_TTL", 15*time.Minute),
		CepAPIURL:    getenv("CEP_API_URL", "https://viacep.com.br/ws"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getdur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
