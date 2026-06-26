// Package config carrega a configuração da aplicação a partir de variáveis
// de ambiente (com suporte a um arquivo .env opcional).
package config

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config agrega todos os parâmetros de execução do serviço.
type Config struct {
	AppEnv       string
	AppPort      string
	DatabaseURL  string
	JWTSecret      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration
	CepAPIURL      string
	AllowedOrigins []string // origens permitidas no CORS (ALLOWED_ORIGINS, separadas por vírgula)

	// Observabilidade (OpenTelemetry). OTLPEndpoint vazio mantém o tracing
	// dormente; preencher OTEL_EXPORTER_OTLP_ENDPOINT liga a exportação de
	// spans para um OTel Collector (fase de microsserviços).
	ServiceName  string
	OTLPEndpoint string
}

// Load lê o .env (se existir) e monta a Config a partir do ambiente,
// aplicando defaults para desenvolvimento.
func Load() *Config {
	_ = godotenv.Load() // ignora ausência do arquivo (produção usa env real)

	return &Config{
		AppEnv:       getenv("APP_ENV", "development"),
		// PORT é injetado por plataformas como o Railway e tem precedência;
		// APP_PORT é o override local; 8080 é o fallback de desenvolvimento.
		AppPort:      getenv("PORT", getenv("APP_PORT", "8080")),
		DatabaseURL:  getenv("DATABASE_URL", "postgres://erp@localhost:5432/erp_estoque?sslmode=disable"),
		JWTSecret:     getenv("JWT_SECRET", "__INSECURE_DEV_JWT_SECRET__"),
		JWTAccessTTL:  getdur("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL: getdur("JWT_REFRESH_TTL", 720*time.Hour),
		CepAPIURL:     getenv("CEP_API_URL", "https://viacep.com.br/ws"),
		AllowedOrigins: strings.Split(getenv("ALLOWED_ORIGINS", "http://localhost:5173"), ","),

		ServiceName:  getenv("OTEL_SERVICE_NAME", "erp-api"),
		OTLPEndpoint: getenv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
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
