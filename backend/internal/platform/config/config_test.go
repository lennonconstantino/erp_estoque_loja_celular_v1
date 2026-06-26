package config

import (
	"strings"
	"testing"
)

// TestValidate_Producao_FailClosed garante que produção recusa subir quando os
// segredos obrigatórios estão ausentes ou usando os defaults inseguros de dev.
func TestValidate_Producao_FailClosed(t *testing.T) {
	casos := []struct {
		nome        string
		cfg         Config
		querEsperar []string // substrings que devem aparecer na mensagem de erro
	}{
		{
			nome:        "JWT_SECRET vazio",
			cfg:         Config{AppEnv: "production", JWTSecret: "", DatabaseURL: "postgres://u:p@db:5432/x?sslmode=require"},
			querEsperar: []string{"JWT_SECRET"},
		},
		{
			nome:        "JWT_SECRET com sentinela de dev",
			cfg:         Config{AppEnv: "production", JWTSecret: devJWTSecret, DatabaseURL: "postgres://u:p@db:5432/x?sslmode=require"},
			querEsperar: []string{"JWT_SECRET"},
		},
		{
			nome:        "DATABASE_URL vazia",
			cfg:         Config{AppEnv: "production", JWTSecret: "um-secret-forte", DatabaseURL: ""},
			querEsperar: []string{"DATABASE_URL"},
		},
		{
			nome:        "DATABASE_URL com default de dev",
			cfg:         Config{AppEnv: "production", JWTSecret: "um-secret-forte", DatabaseURL: devDatabaseURL},
			querEsperar: []string{"DATABASE_URL"},
		},
		{
			nome:        "ambos inseguros listam as duas variáveis",
			cfg:         Config{AppEnv: "production", JWTSecret: devJWTSecret, DatabaseURL: devDatabaseURL},
			querEsperar: []string{"JWT_SECRET", "DATABASE_URL"},
		},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			err := c.cfg.Validate()
			if err == nil {
				t.Fatalf("esperava erro de validação, obteve nil")
			}
			for _, sub := range c.querEsperar {
				if !strings.Contains(err.Error(), sub) {
					t.Errorf("mensagem de erro %q não menciona %q", err.Error(), sub)
				}
			}
		})
	}
}

// TestValidate_Producao_Valida aceita configuração de produção com segredos reais.
func TestValidate_Producao_Valida(t *testing.T) {
	cfg := Config{
		AppEnv:      "production",
		JWTSecret:   "valor-unico-e-longo-gerado-por-openssl-rand",
		DatabaseURL: "postgres://postgres:senha@db.exemplo.supabase.co:5432/postgres?sslmode=require",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("produção corretamente configurada não deveria falhar: %v", err)
	}
}

// TestValidate_Producao_CaseInsensitive cobre "Production"/"PRODUCTION".
func TestValidate_Producao_CaseInsensitive(t *testing.T) {
	for _, env := range []string{"production", "Production", "PRODUCTION"} {
		cfg := Config{AppEnv: env, JWTSecret: devJWTSecret, DatabaseURL: devDatabaseURL}
		if err := cfg.Validate(); err == nil {
			t.Errorf("APP_ENV=%q deveria ser tratado como produção e falhar", env)
		}
	}
}

// TestValidate_Desenvolvimento_Permissivo garante que defaults de dev são aceitos
// fora de produção (não devemos travar o fluxo local com `make be-run`).
func TestValidate_Desenvolvimento_Permissivo(t *testing.T) {
	for _, env := range []string{"development", "", "test", "staging"} {
		cfg := Config{AppEnv: env, JWTSecret: devJWTSecret, DatabaseURL: devDatabaseURL}
		if err := cfg.Validate(); err != nil {
			t.Errorf("APP_ENV=%q deveria permitir defaults de dev, mas falhou: %v", env, err)
		}
	}
}

// TestLoad_DefaultsDeDev confirma que Load() aplica os defaults de desenvolvimento
// quando nenhuma env relevante está setada (ambiente limpo).
func TestLoad_DefaultsDeDev(t *testing.T) {
	for _, k := range []string{"APP_ENV", "JWT_SECRET", "DATABASE_URL", "PORT", "APP_PORT"} {
		t.Setenv(k, "")
	}
	cfg := Load()
	if cfg.JWTSecret != devJWTSecret {
		t.Errorf("JWTSecret default = %q, esperava %q", cfg.JWTSecret, devJWTSecret)
	}
	if cfg.DatabaseURL != devDatabaseURL {
		t.Errorf("DatabaseURL default = %q, esperava %q", cfg.DatabaseURL, devDatabaseURL)
	}
	// Por padrão (sem APP_ENV=production) a validação deve passar.
	if err := cfg.Validate(); err != nil {
		t.Errorf("config default de dev não deveria falhar na validação: %v", err)
	}
}
