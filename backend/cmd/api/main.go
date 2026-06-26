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

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/clientes"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/compras"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/iam"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/auth"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/config"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/database"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/httpserver"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/observability"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer pool.Close()

	// Observabilidade (OpenTelemetry): métricas sempre ligadas em /metrics;
	// tracing dormente até OTEL_EXPORTER_OTLP_ENDPOINT existir.
	obs, err := observability.Setup(ctx, observability.Config{
		ServiceName:  cfg.ServiceName,
		ServiceEnv:   cfg.AppEnv,
		OTLPEndpoint: cfg.OTLPEndpoint,
	})
	if err != nil {
		log.Fatalf("erro ao inicializar observabilidade: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := obs.Shutdown(shutdownCtx); err != nil {
			log.Printf("erro no shutdown da observabilidade: %v", err)
		}
	}()

	authMgr := auth.NewManager(cfg.JWTSecret, cfg.JWTAccessTTL)

	// Módulos (cada bounded context monta a si mesmo via DI).
	iamMod := iam.New(pool, authMgr, cfg.JWTRefreshTTL)
	clientesMod := clientes.New(pool, authMgr, cfg.CepAPIURL)
	fornecedoresMod := fornecedores.New(pool, authMgr, cfg.CepAPIURL)
	catalogoMod     := catalogo.New(pool, authMgr)
	estoqueMod      := estoque.New(pool, authMgr, catalogoMod.Writer())
	comprasMod      := compras.New(pool, authMgr, catalogoMod.Reader(), estoqueMod.Writer(), fornecedoresMod.Writer())

	r := httpserver.NewRouter()
	r.Handle("/metrics", obs.MetricsHandler)
	r.Route("/api/v1", func(api chi.Router) {
		api.Mount("/", iamMod.Router())
		api.Mount("/clientes", clientesMod.Router())
		api.Mount("/fornecedores", fornecedoresMod.Router())
		api.Mount("/categorias", catalogoMod.CategoriasRouter())
		api.Mount("/produtos", catalogoMod.ProdutosRouter())
		api.Mount("/estoque", estoqueMod.Router())
		api.Mount("/compras", comprasMod.Router())
		// ... demais módulos
	})

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Inicia o servidor e aguarda sinal de término para shutdown gracioso.
	go func() {
		log.Printf("API ouvindo em :%s (env=%s, tracing=%t)", cfg.AppPort, cfg.AppEnv, obs.TracingAtivo)
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
