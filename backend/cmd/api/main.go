// Command api é o entrypoint do monólito modular. Monta a infraestrutura
// compartilhada (config, banco, auth) e registra os módulos sob /api/v1.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/config"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/database"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer pool.Close()

	authMgr := auth.NewManager(cfg.JWTSecret, cfg.JWTAccessTTL)

	// Módulos (cada bounded context monta a si mesmo via DI).
	clientesMod := clientes.New(pool, authMgr, cfg.CepAPIURL)

	r := httpserver.NewRouter()
	r.Route("/api/v1", func(api chi.Router) {
		api.Mount("/clientes", clientesMod.Router())
		// api.Mount("/fornecedores", fornecedoresMod.Router())
		// ... demais módulos
	})

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Inicia o servidor e aguarda sinal de término para shutdown gracioso.
	go func() {
		log.Printf("API ouvindo em :%s (env=%s)", cfg.AppPort, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("erro no servidor: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown forçado: %v", err)
	}
	log.Println("servidor encerrado")
}
