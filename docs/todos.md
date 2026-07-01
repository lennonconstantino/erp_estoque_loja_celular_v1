# Lista de Verificação — Implementação do ERP

Estado atual: migrations 000001–000010 prontas (inclui `seed_demo`), plataforma
(auth JWT, httpserver, resilience, config, database) pronta, **todos os módulos
de negócio implementados** (iam, clientes, fornecedores, catálogo, estoque,
compras, vendas, relatórios). Frontend completo: todas as telas existem e
compartilham o **kit de UI** em `@/components/ui` (Fases 1–8 e 10 concluídas;
resta a Fase 9 — deploy).

> **Ordem lógica:** backend antes do frontend — o frontend consome a API.
> A partir da fase 3 (cadastros) é possível trabalhar em paralelo.

> **Definition of Done por tarefa ([docs/mandates.md](mandates.md)):** cada tarefa
> só conclui com **D1** testes unitários no mesmo passo (`go test -cover`,
> `domain/` ≥ 80%), **D2** este checklist atualizado com `[x]`, e **D3** veredito
> `CONFORME` de um subagent juiz independente contra a spec da fase. O `curl`/PASS
> de cada fase é o piso, não o teto.

---

## Fase 0 — Infraestrutura local

- [x] Copiar `backend/.env.example` → `backend/.env` e preencher variáveis
- [x] Confirmar que `docker compose up -d` sobe db + migrate + api sem erros
- [x] Verificar `GET /health` → 200
- [x] Confirmar seed: login `admin@loja.local` / `admin123` funciona via `POST /api/v1/auth/login`

---

## Fase 1 — Módulo IAM (autenticação e usuários)

> Portão de entrada de todo o sistema. Sem isso o frontend não funciona.

### Backend

- [x] Criar `internal/modules/iam/domain/` — entidades `Usuario`, `Papel`, `Permissao` + erros de domínio
- [x] Criar `internal/modules/iam/ports/inbound.go` — `AuthService` (Login, Refresh, Logout, CriarUsuario, ListarUsuarios)
- [x] Criar `internal/modules/iam/ports/outbound.go` — `UsuarioRepository`, `TokenStore`
- [x] Criar `internal/modules/iam/application/service.go` — casos de uso (bcrypt, emissão de JWT via `platform/auth`, rotação de refresh token)
- [x] Criar `internal/modules/iam/adapters/outbound/postgres/usuario_repo.go` — CRUD em `iam.usuarios`, papéis, permissões, refresh tokens
- [x] Criar `internal/modules/iam/adapters/inbound/http/handler.go` — handlers: `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`
- [x] Criar `internal/modules/iam/adapters/inbound/http/handler.go` — handlers de usuários: `GET /usuarios`, `POST /usuarios`, `PATCH /usuarios/{id}` (somente ADMIN)
- [x] Criar `internal/modules/iam/adapters/inbound/http/router.go` — rotas + RBAC
- [x] Criar `internal/modules/iam/module.go` — DI: instancia repo, service, router
- [x] Montar módulo em `cmd/api/main.go` sob `/api/v1`
- [x] Testar: login retorna `access_token` + `refresh_token`; refresh rotaciona; logout revoga

### Frontend

- [x] Implementar fluxo de login em `LoginPage.tsx` (chamar `POST /api/v1/auth/login`, salvar tokens via `lib/auth.ts`)
- [x] Implementar proteção de rotas em `App.tsx` — redireciona para login se não autenticado
- [x] Implementar renovação automática de token (já esperada em `lib/api.ts`) e logout no menu

---

## Fase 2 — Módulo Fornecedores

> Necessário antes de Compras. Segue o mesmo molde de `clientes`.

### Backend

- [x] Criar `internal/modules/fornecedores/domain/` — entidade `Fornecedor` (CNPJ), erros
- [x] Criar `ports/inbound.go` — `FornecedorService`
- [x] Criar `ports/outbound.go` — `FornecedorRepository`, `CepGateway` (reusar interface de clientes)
- [x] Criar `application/service.go` — CRUD + busca de CEP
- [x] Criar `adapters/outbound/postgres/fornecedor_repo.go`
- [x] Reusar `adapters/outbound/cep/viacep.go` (criado em fornecedores/adapters/outbound/cep com domínio próprio)
- [x] Criar `adapters/inbound/http/handler.go` + `router.go` — RBAC: `fornecedores:read`, `fornecedores:write`
- [x] Criar `module.go` e montar em `main.go`
- [x] Testar CRUD de fornecedor com CNPJ único e preenchimento de CEP

### Frontend

- [x] Criar `pages/FornecedoresPage.tsx` — listagem com paginação
- [x] Criar formulário de cadastro/edição (CNPJ + contatos + CEP automático)
- [x] Integrar com API (`GET /fornecedores`, `POST /fornecedores`, `PUT /fornecedores/{id}`)

---

## Fase 3 — Módulo Catálogo (Categorias e Produtos)

> Dependência de Compras e Vendas. Produtos têm saldo materializado.

### Backend

- [x] Criar `internal/modules/catalogo/domain/` — entidades `Categoria`, `Produto` (invariante: `custo < venda`), erros
- [x] Criar `ports/inbound.go` — `CategoriaService`, `ProdutoService`
- [x] Criar `ports/outbound.go` — `CategoriaRepository`, `ProdutoRepository`, `EstoqueReader` (porta de saída, implementada por `estoque`)
- [x] Criar `application/service.go` — CRUD + cálculo de margem
- [x] Criar `adapters/outbound/postgres/` — repos de categoria e produto
- [x] Criar `adapters/inbound/http/handler.go` + `router.go` — RBAC: `catalogo:read`, `catalogo:write`
- [x] Criar `module.go` e montar em `main.go`
- [x] Expor interface `CatalogoReader` (usada por compras/vendas) e `CatalogoWriter` (usada por estoque)
- [x] Testar: `CHECK p_custo < p_venda` rejeita produto inválido; saldo começa em 0

### Frontend

- [x] Criar `pages/CategoriasPage.tsx` — listagem + modal de cadastro/edição
- [x] Criar `pages/ProdutosPage.tsx` — listagem com saldo atual e indicador de mínimo
- [x] Formulário de produto: custo, venda, margem calculada em tempo real, estoque mínimo
- [x] Integrar com API (`GET /categorias`, `GET /produtos`, `POST`, `PUT`)

---

## Fase 4 — Módulo Estoque (Razão e Ajustes)

> Fonte da verdade de saldo. Implementar antes de compras e vendas.

### Backend

- [x] Criar `internal/modules/estoque/domain/` — entidades `Movimentacao` (append-only), `Ajuste`, tipos `COMPRA/VENDA/AJUSTE_ENTRADA/AJUSTE_SAIDA`
- [x] Criar `ports/inbound.go` — `EstoqueService` (LancarAjuste, ConsultarMovimentacoes, ConsultarAjustes)
- [x] Criar `ports/outbound.go` — `MovimentacaoRepository`, `AjusteRepository`, `CatalogoWriter` (atualiza `estoque_a_pro` + `disp_pro`)
- [x] Criar `application/service.go` — lógica de ajuste append-only, atualização de saldo via `CatalogoWriter`
- [x] Criar `adapters/outbound/postgres/movimentacao_repo.go` + `ajuste_repo.go`
- [x] Criar `adapters/inbound/http/handler.go` + `router.go` — `POST /estoque/ajustes` (RBAC: `estoque:write`); `GET /estoque/{produtoId}` (RBAC: `estoque:read`)
- [x] Criar `module.go` e montar em `main.go`; injetar `CatalogoWriter` do módulo catálogo
- [x] Verificar trigger no banco que bloqueia UPDATE/DELETE em `estoque.ajustes` e `estoque.movimentacoes`
- [x] Testar: ajuste registra `saldo_ant`/`saldo_atu`; tentativa de editar ajuste retorna erro

### Frontend

- [x] Criar `pages/AjustesEstoquePage.tsx` — formulário de ajuste (produto, quantidade, motivo, tipo entrada/saída)
- [x] Exibir histórico de ajustes (somente leitura, sem botão de editar/excluir)
- [x] Integrar com API (`POST /estoque/ajustes`, `GET /estoque/{produtoId}`)

---

## Fase 5 — Módulo Compras

> Entrada de mercadoria: cria `compra_master` + `detalhe_compras` + movimentação de estoque.

### Backend

- [x] Criar `internal/modules/compras/domain/` — entidades `Compra`, `DetalheCompra`, erros
- [x] Criar `ports/inbound.go` — `CompraService` (CriarCompra, ConfirmarCompra, ListarCompras, BuscarCompra)
- [x] Criar `ports/outbound.go` — `CompraRepository`, `CatalogoReader`, `EstoqueWriter`, `FornecedorWriter`
- [x] Criar `application/service.go` — caso de uso `ConfirmarCompra`: valida itens via `CatalogoReader`, persiste compra, emite movimentação via `EstoqueWriter`, atualiza `dt_ult_comp_for` via `FornecedorWriter`
- [x] Criar `adapters/outbound/postgres/compra_repo.go`
- [x] Criar `adapters/inbound/http/handler.go` + `router.go` — RBAC: `compras:read`, `compras:write`
- [x] Criar `module.go` e montar em `main.go`
- [x] Testar: confirmar compra aumenta saldo do produto; movimentação registrada com `saldo_ant`/`saldo_atu`

### Frontend

- [x] Criar `pages/ComprasPage.tsx` — listagem de compras
- [x] Criar formulário "Nova Compra": selecionar fornecedor + NF + adicionar itens (produto, qtd, custo, venda)
- [x] Confirmar compra: exibir resumo e botão de confirmação
- [x] Integrar com API (`GET /compras`, `POST /compras`, `POST /compras/{id}/confirmar`)

---

## Fase 6 — Módulo Vendas (PDV)

> Saída de estoque + documento fiscal. Fluxo mais crítico do sistema.

### Backend

- [x] Criar `internal/modules/vendas/domain/` — entidades `Venda`, `DetalheVenda` (XOR cliente/consumidor), erros (`SaldoInsuficiente`)
- [x] Criar `ports/inbound.go` — `VendaService` (CriarVenda, ConfirmarVenda, ListarVendas, BuscarVenda)
- [x] Criar `ports/outbound.go` — `VendaRepository`, `CatalogoReader`, `EstoqueWriter`, `ClienteWriter`, `FiscalGateway`
- [x] Criar `adapters/outbound/fiscal/` — `FiscalGateway` (cupom/NF via API externa)
  - [x] Implementar `emitirCupom` e `emitirNF` com resilience Policy (retry + circuit breaker)
- [x] Criar `application/service.go` — caso de uso `ConfirmarVenda`: valida saldo via `CatalogoReader`, baixa saldo via `EstoqueWriter` (atômico `UPDATE WHERE estoque_a_pro >= qtd`), persiste venda, emite documento via `FiscalGateway`, atualiza `dt_ult_comp_cli`
- [x] Garantir que saldo negativo é impossível (constraint no banco + lock pessimista via `DecrementarSaldo`)
- [x] Criar `adapters/inbound/http/handler.go` + `router.go` — RBAC: `vendas:read`, `vendas:write`
- [x] Criar `module.go` e montar em `main.go`
- [x] Testar: domain/ 100%, application/ 74.5% — guard atômico implementado em catalogo.DecrementarSaldo

### Frontend

- [x] Criar `pages/VendasPage.tsx` — listagem de vendas com status, confirmar e detalhe
- [x] Criar `pages/NovaVendaPage.tsx` (PDV):
  - [x] Buscar cliente por CPF (preencher automaticamente ou abrir cadastro rápido)
  - [x] Adicionar itens com validação de saldo em tempo real
  - [x] Aplicar desconto
  - [x] Confirmar venda e exibir documento fiscal emitido
- [x] Integrar com API (`GET /vendas`, `POST /vendas`, `POST /vendas/{id}/confirmar`)

---

## Fase 7 — Relatórios

> Necessário para o critério de aceitação do cliente.

### Backend

- [x] Criar endpoints de relatório (módulo `relatorios` dedicado):
  - [x] `GET /relatorios/produtos/abaixo-do-minimo` — RBAC: `relatorios:read`
  - [x] `GET /relatorios/produtos/mais-vendidos?de=&ate=&limite=` — RBAC: `relatorios:read`
  - [x] `GET /relatorios/produtos/menos-vendidos?de=&ate=&limite=` — RBAC: `relatorios:read`
  - [x] `GET /relatorios/vendas?de=&ate=` — resumo por período
  - [x] `GET /relatorios/compras?de=&ate=` — resumo por período
- [x] Registrar rotas em `main.go`

### Frontend

- [x] Criar `pages/RelatoriosPage.tsx` — navegação entre relatórios (abas)
- [x] Exibir tabela de produtos abaixo do mínimo (destacar em vermelho)
- [x] Exibir tabela de mais/menos vendidos por período
- [x] Exibir resumo de vendas e compras por período (cards de métricas)
- [x] Integrar com API de relatórios

---

## Fase 8 — Qualidade e Segurança

- [x] Rodar `make be-vet` sem erros
- [x] Rodar `make be-test` — todos os testes passam
- [x] Rodar `pnpm tsc --noEmit` no frontend sem erros
- [x] Rodar `pnpm lint` no frontend sem erros (instalado `eslint-plugin-react-hooks`, desabilitada a regra `set-state-in-effect` que gera falso positivo no padrão de data-fetching)
- [x] Revisar checklist de segurança em `docs/reference/checklist.md`
- [x] Verificar que `JWT_SECRET` é longo e aleatório (`.env.example` tem `JWT_SECRET=` vazio; padrão de dev é `__INSECURE_DEV_JWT_SECRET__`, claramente marcado)
- [x] Verificar CORS: implementado middleware `cors()` em `platform/httpserver`; `ALLOWED_ORIGINS` configurável via env (padrão dev: `http://localhost:5173`)
- [x] Confirmar que `admin@loja.local` / `admin123` está documentado para troca em produção (checklist de deploy em `docs/reference/checklist.md`)

---

## Fase 9 — Deploy (Railway + Supabase)

> **Estado:** **NO AR** desde 2026-06-30. Backend e frontend rodando no Railway
> (projeto `erp-estoque`, env `production`) contra Postgres no Supabase. Passo a
> passo em [docs/setup/railway-deployment.md](setup/railway-deployment.md).
>
> - Backend:  https://erp-estoque-backend-production.up.railway.app  (`/health` → 200)
> - Frontend: https://erp-estoque-frontend-production.up.railway.app (`/health` → ok)
>
> **Duas correções necessárias descobertas na subida** (ver §Lições em
> railway-deployment.md): (1) `DATABASE_URL` precisa ser a do **Session Pooler**
> do Supabase (IPv4) — a conexão direta `db.<ref>.supabase.co` é IPv6-only e o
> egress do Railway é IPv4; (2) `frontend/Dockerfile` faz o nginx escutar em
> `$PORT` (Railway sonda o healthcheck nessa porta, não na 80).

### Preparação no repositório (pronto)

- [x] Runner de migrations `cmd/migrate` (golang-migrate embarcado; `up`/`down`/`version`/`force`)
- [x] Backend lê `PORT` do Railway (fallback `APP_PORT` → `8080`)
- [x] `backend/Dockerfile` compila e copia `/app/api` **e** `/app/migrate` (+ `migrations/`)
- [x] `frontend/nginx.conf` com `GET /health` para o healthcheck
- [x] `backend/railway.json` e `frontend/railway.json` (builder Dockerfile, healthcheck, pre-deploy `/app/migrate up`)
- [x] Templates `backend/.env.production.example` e `frontend/.env.production.example` (Supabase + CORS + JWT)
- [x] `.gitignore` cobre `.segredo`, `.env` e `.env.*` (mantendo os `*.example`)
- [x] Validação local: build dos dois binários, imagem Docker do backend, ciclo `up`/`down`/`up` num Postgres limpo
- [x] Fix: `000010_seed_demo.down.sql` agora reverte o ledger (DISABLE/ENABLE trigger pelo owner)

### Execução do deploy (concluída em 2026-06-30)

- [x] Projeto Supabase criado; `DATABASE_URL` via **Session Pooler** (IPv4, `aws-1-us-east-2.pooler.supabase.com:5432`, usuário `postgres.<ref>`, `sslmode=require`)
- [x] Projeto Railway `erp-estoque` com dois serviços: `erp-estoque-backend` e `erp-estoque-frontend`
- [x] Variáveis do backend no Railway (`DATABASE_URL` do pooler, `JWT_SECRET`, `ALLOWED_ORIGINS`, `JWT_*_TTL`, `CEP_API_URL`, `APP_ENV=production`); removidas vars legadas `DB_*`/`APP_PORT` não usadas pelo código
- [x] `VITE_API_BASE_URL` setado no frontend (build-arg) apontando p/ o domínio do backend antes do build
- [x] Pre-deploy `/app/migrate up` confirmado nos logs (`migrate: … versão mais recente`)
- [x] Deploy dos dois serviços; `/health` do backend → `{"status":"ok"}` e do frontend → `ok`
- [x] Migrations aplicadas no Supabase confirmadas (versão 10; schemas iam/clientes/fornecedores/catalogo/estoque/compras/vendas; seed admin presente)
- [x] Ciclo em produção verificado: login admin → 200; CORS preflight/real com `Access-Control-Allow-Origin` correto; leitura de todos os módulos (produtos, categorias, clientes, fornecedores, vendas, compras, usuários) e relatórios → 200

> Pendência de durabilidade: o serviço de frontend faz auto-deploy do branch `main`
> no GitHub. O fix do `frontend/Dockerfile` (nginx em `$PORT`) foi deployado via
> `railway up` a partir dos arquivos locais; **precisa ser commitado em `main`**
> para que um futuro push não reintroduza a versão quebrada.

---

## Fase 10 — Polimento de UI, consistência e dados de demonstração

> Trabalho pós-MVP: dados de demonstração, correção de carga e padronização
> visual de todas as telas. Não altera o backend de negócio.

### Dados de demonstração e roteamento

- [x] Criar migration `000010_seed_demo` (idempotente): fornecedores, clientes, produtos com saldo, compras e vendas CONFIRMADA
- [x] Criar `pages/ClientesPage.tsx` e registrar rota `/clientes` em `App.tsx`
- [x] Corrigir rota de ajuste de estoque para `/estoque/ajustes` (estava `/estoque`, caía no catch-all)

### Correção de carga de dados (bug "Erro desconhecido")

- [x] Adicionar o prefixo `/api/v1` às chamadas de Categorias, Produtos, Vendas, Estoque, Relatórios e NovaVenda (faltava → 404 → "Erro desconhecido")
- [x] Confirmar via API: categorias, produtos, vendas e relatórios retornam dados

### Kit de UI compartilhado (`@/components/ui`)

- [x] Criar `PageShell` (casca: cabeçalho + `<main>`), `Button`/`buttonClasses`, `StatusBadge`, `DataTable<T>`, `Modal`, `Field`/`inputClasses`
- [x] Refatorar todas as telas para o kit (Categorias, Produtos, Vendas, Estoque, Relatórios, Compras, Clientes, Fornecedores) — padrão único (cinza-900, tabelas com borda, sem azul/índigo ad-hoc)

### Ordenação de tabelas

- [x] Implementar ordenação por coluna no `DataTable` (`sortAccessor`; ciclo asc → desc → sem ordenação; comparação numérica/`localeCompare` pt-BR/data)
- [x] Habilitar colunas ordenáveis em todas as telas com dados tabulados
- [ ] (Opcional) Ordenação global no servidor via parâmetros `sort`/`order` nos endpoints — hoje a ordenação é client-side sobre a página carregada

### Verificação

- [x] `pnpm tsc --noEmit` e `pnpm lint` sem erros
- [x] Rebuild do container `frontend` e telas validadas no browser

### Refino visual técnico + Dark/Light mode

> Segunda passada de padronização: tema escuro/claro, primitivos novos e
> consolidação dos formulários no estilo pill/técnico.

- [x] **Tema Dark/Light** via `@/lib/theme` (`ThemeProvider` + `useTheme`), classe
  `.dark` no `<html>`, tokens HSL em `index.css` e `ThemeToggle` no `PageShell`
- [x] Migrar telas para **tokens semânticos** (`bg-card`/`text-foreground`/`text-destructive`…)
  removendo cores cruas (ex.: `text-red-600` em Categorias; tendência negativa do Dashboard)
- [x] Novos primitivos no kit: `Tabs<T>` (pill), `Sidebar`, `CommandPalette`/`Command`/`Dialog`
  (⌘K), `ThemeToggle`, `Toaster`/`toast` (`sonner`)
- [x] `PageShell` reestruturado: `Sidebar` fixa + cabeçalho com paleta de comandos e toggle de tema
- [x] Helpers de formulário compacto no kit (`inputClassesCompact`/`compactLabelClass`) e adoção
  nas grades densas de itens (Compras, NovaVenda/PDV) — fim das classes ad-hoc duplicadas
- [x] Padronização final das páginas restantes (Categorias: erros/paginação no padrão; Produtos,
  Vendas, Compras, Fornecedores conferidos)
- [x] **Correção de build:** `cn` ausente em Compras, `Modal` aceitar `max-w-4xl`, imports
  não usados (NovaVenda, Relatórios), `icon: any` → `LucideIcon` no Dashboard
- [x] Verificação: `pnpm tsc --noEmit`, `pnpm lint` e `pnpm build` sem erros
- [ ] Pendências de a11y/UX e toast-tema documentadas em
  [docs/reference/design-system.md](reference/design-system.md#pendências-e-próximos-passos-ui) — agendar

---

## Critério de Aceitação (conforme brief do cliente)

- [ ] Receber mercadoria → registrar compra → saldo atualizado automaticamente
- [ ] Atender cliente → registrar venda → saldo baixado → documento fiscal emitido
- [ ] Identificar ruptura → relatório de produtos abaixo do mínimo
- [ ] Lançar ajuste com motivo → registro imutável no histórico
- [ ] Emitir relatório de vendas e compras do período
- [ ] Operação de um dia completo sem planilha, sem saldo negativo, sem erro fiscal

---

## Hardening de segurança — Secrets & Config (review focado)

> Revisão focada com a skill `security-best-practices` (specs Go + React). Escopo:
> Secrets & Config (GO-CONFIG-001, REACT-CONFIG-001).

- [x] **F1+F2 — fail-closed em produção:** `Config.Validate()` recusa subir quando
  `JWT_SECRET`/`DATABASE_URL` estão vazios ou no default inseguro de dev; chamada em
  `cmd/api/main.go`. Sentinelas extraídas como consts (`devJWTSecret`/`devDatabaseURL`).
- [x] **D1 — testes:** `config_test.go` cobre prod inseguro/seguro, case-insensitive e
  dev permissivo. `go test -cover ./internal/platform/config/...` = **84.2%** (≥ 80%).
- [x] **F3 — info disclosure:** host real do Supabase trocado por `SEU_PROJECT_REF` em
  `backend/.env.production.example`.
- [x] Verificação: `go vet ./...`, `go build ./...` e suíte completa (`go test ./...`) OK.
- [x] Confirmado (review): `.env`/`.env.production` reais nunca versionados nem no
  histórico; sem logging de segredos; bundle do frontend expõe só `VITE_API_BASE_URL`.
- [x] **D3 — juiz independente:** veredito **CONFORME** (sem buracos no fail-closed; cobertura 84.2%; vet/build/test OK; sem regressões).
- [x] **Docs atualizadas:** `backend/CLAUDE.md` (§Config), `docs/reference/security.md` e
  `docs/setup/railway-deployment.md` documentam o fail-closed de produção.
