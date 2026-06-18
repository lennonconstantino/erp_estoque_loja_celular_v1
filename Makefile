# Makefile do projeto (backend + frontend + infra).
# Variáveis do backend (ex.: DATABASE_URL) vêm de backend/.env.
-include backend/.env
export

MIGRATE := migrate -path ./backend/migrations -database "$(DATABASE_URL)"
COMPOSE := docker compose --env-file backend/.env -f docker-compose.yml

# Sentinelas de desenvolvimento que NUNCA devem ir para produção (ver check-secrets).
INSECURE_JWT := __INSECURE_DEV_JWT_SECRET__
INSECURE_DB  := __INSECURE_DEV_DB_PASSWORD__

.PHONY: help \
	up down logs check-secrets \
	be-build be-run be-test be-vet be-fmt \
	migrate-up migrate-down migrate-create seed reset \
	fe-install fe-dev fe-build fe-lint \
	build test lint

help:          ## lista os alvos disponíveis
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

## ---- Infra (docker) ----------------------------------------------------
up:            ## sobe banco + migrations + api + frontend
	@if [ -z "$$JWT_SECRET" ] || [ "$$JWT_SECRET" = "$(INSECURE_JWT)" ]; then \
		echo "  ⚠ JWT_SECRET usando o default de dev — OK p/ local, NÃO use em produção (rode 'make check-secrets')"; \
	fi
	$(COMPOSE) up -d --build

down:          ## derruba os containers
	$(COMPOSE) down

logs:          ## segue os logs da api (use s=frontend, s=db, ... para outro serviço)
	$(COMPOSE) logs -f $(or $(s),api)

check-secrets: ## falha se JWT_SECRET/DB_PASSWORD ainda forem os defaults de dev (pré-deploy)
	@fail=0; \
	if [ -z "$$JWT_SECRET" ] || [ "$$JWT_SECRET" = "$(INSECURE_JWT)" ]; then \
		echo "  ✗ JWT_SECRET ausente ou no default de dev — defina um segredo real"; fail=1; \
	else echo "  ✓ JWT_SECRET definido"; fi; \
	if [ -z "$$DB_PASSWORD" ] || [ "$$DB_PASSWORD" = "$(INSECURE_DB)" ]; then \
		echo "  ✗ DB_PASSWORD ausente ou no default de dev — defina uma senha real"; fail=1; \
	else echo "  ✓ DB_PASSWORD definido"; fi; \
	if [ "$$fail" = "1" ]; then \
		echo "  Segredos inseguros — aceitável só em desenvolvimento. Exporte valores reais para produção."; \
		exit 1; \
	fi; \
	echo "  Segredos OK para produção."

## ---- Backend (Go) ------------------------------------------------------
be-build:      ## compila o binário em backend/bin/api
	cd backend && go build -o bin/api ./cmd/api

be-run:        ## roda a API localmente (:8080)
	cd backend && go run ./cmd/api

be-test:       ## roda a suíte de testes do backend
	cd backend && go test ./...

be-vet:        ## go vet
	cd backend && go vet ./...

be-fmt:        ## formata o código Go
	cd backend && gofmt -w .

## ---- Migrations (golang-migrate) --------------------------------------
migrate-up:    ## aplica todas as migrations
	$(MIGRATE) up

migrate-down:  ## reverte a última migration
	$(MIGRATE) down 1

migrate-create: ## make migrate-create name=add_xyz
	migrate create -ext sql -dir ./backend/migrations -seq $(name)

reset:         ## DROP total + recria (inclui seed)
	$(MIGRATE) drop -f
	$(MIGRATE) up

## ---- Frontend (React/Vite) --------------------------------------------
fe-install:    ## instala dependências (pnpm)
	cd frontend && pnpm install

fe-dev:        ## Vite dev server
	cd frontend && pnpm dev

fe-build:      ## build de produção (tsc + vite)
	cd frontend && pnpm build

fe-lint:       ## ESLint
	cd frontend && pnpm lint

## ---- Agregados ---------------------------------------------------------
build: be-build fe-build   ## compila backend e frontend
test: be-test              ## roda os testes (backend)
lint: be-vet fe-lint       ## checagens estáticas (backend + frontend)
