# Changelog

Histórico de mudanças do ERP de estoque para loja de acessórios de celular.
Formato baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).

## [1.3.0] — 2026-07-01

Esteira de CI/CD (GitHub Actions como gate + auto-deploy nativo do Railway).

### Adicionado

- **`ci.yml`** (PR + push main): jobs `backend-test` (go vet + `go test -cover`, com
  `-race` só no main, + FF1), `frontend` (pnpm lint + build), `build-images`
  (docker build backend/frontend) e `integration` (sobe o stack e roda FF1+FF2 como
  gate + FF3 p99 informativa). Versões vêm do repo (`go.mod`, `.nvmrc`, Python 3.12).
- **`promote-production.yml`**: promoção manual com aprovação (GitHub Environment)
  que faz fast-forward `main`→`production`, com guarda anti-divergência.
- **`nightly-chaos.yml`**: FF4 (chaos) agendado, não-bloqueante.
- **`run_all.py`** aceita subconjunto de FFs (`ff1 ff2 ff3`); `requirements.txt`
  (httpx); `.nvmrc` (Node 20); alvos `make fitness` / `make fitness-static`.
- Documentação da esteira em [docs/setup/cicd.md](../setup/cicd.md).

## [1.2.0] — 2026-07-01

Fase 9 executada: aplicação **no ar** em produção (Railway + Supabase), com o
ciclo de negócio validado contra o ambiente real.

### Adicionado

- **Deploy em produção**: backend e frontend no Railway (projeto `erp-estoque`)
  contra Postgres no Supabase; migrations aplicadas via pre-deploy `/app/migrate up`;
  ciclo compra→venda→ajuste→relatório verificado em produção (saldo, cupom fiscal e
  ledger imutável conferindo). (`73f5aab`)
- **Retrospectiva de deploy** com diagramas Mermaid (topologia, sequência, mapa
  mental, runbook de troubleshooting): [docs/licoes-aprendidas.md](../licoes-aprendidas.md). (`73f5aab`)

### Corrigido

- **Frontend no Railway ficava `Failed`**: o nginx escutava fixo na porta 80, mas o
  healthcheck do Railway bate na porta injetada em `$PORT`. O `frontend/Dockerfile`
  passa a reescrever `listen 80;` para `$PORT` no boot (fallback 80 p/ local). (`c0d3df8`)

### Alterado

- **`DATABASE_URL` de produção** passa a usar a **Session pooler** do Supabase (IPv4).
  A conexão direta (`db.<ref>.supabase.co`) é IPv6-only e o egress do Railway é IPv4,
  causando `network is unreachable` no `migrate`. Docs de Supabase, Railway e o
  checklist de segurança atualizados. (`c0d3df8`)
- Removidas variáveis legadas `DB_*`/`APP_PORT` do serviço de backend (o código lê só
  `DATABASE_URL`).

## [1.1.0] — 2026-06-26

Hardening de produção, acessibilidade (WCAG) em todas as telas e polimento do
kit de UI, posteriores à entrega do MVP.

### Adicionado

- **Validação fail-closed da config de produção**: a API aborta o startup quando
  `JWT_SECRET` ou `DATABASE_URL` estão ausentes ou usam defaults inseguros de
  desenvolvimento, com testes unitários cobrindo a lógica de validação. (`06e9add`)
- Documentação de segurança das checagens de config em `CLAUDE.md` e nos guias de
  deploy. (`06e9add`)
- Suporte a **prefers-reduced-motion** (WCAG 2.3.3) e link "pular para o conteúdo"
  (skip-to-content). (`06e9add`)

### Alterado

- Acessibilidade em todas as páginas: `aria-label` em botões de ícone,
  `role="alert"` em mensagens de erro e associação correta de labels a campos de
  formulário. (`06e9add`)
- Modais migrados para **Radix Dialog**, com conformidade total de acessibilidade
  (foco, ESC, leitor de tela). (`06e9add`)
- Cores hardcoded substituídas por tokens semânticos do tema, garantindo
  consistência no dark/light mode; polimento geral dos componentes de UI. (`06e9add`)
- `.env.production.example` usa placeholder de projeto em vez de host Supabase
  hardcoded; README e design-system docs atualizados. (`06e9add`)

## [1.0.0] — 2026-06-26

Entrega completa do MVP: todos os bounded contexts implementados em Go, frontend
cobrindo todas as telas e infraestrutura de deploy preparada.

### Adicionado

#### Bounded contexts (backend)

- **IAM** — contexto completo de identidade e acesso: modelos de domínio
  `Usuario` e `RefreshToken`, erros padronizados, serviço de aplicação com login,
  refresh/logout e CRUD de usuários, adaptador HTTP REST protegido por RBAC,
  persistência PostgreSQL e configuração de TTL do refresh token. Seed do admin
  com hash bcrypt correto para `admin123`. (`ec455c1`)
- **Fornecedores** — domínio, serviço CRUD, persistência PostgreSQL e API HTTP
  autenticada com RBAC; página frontend com busca, paginação e consulta
  automática de endereço por CEP. (`d983a5d`)
- **Catálogo de produtos** — regras de negócio para `Categoria` e `Produto`,
  serviços, portas de repositório e adaptadores Postgres, endpoints HTTP com
  RBAC e telas de gestão (busca, paginação e CRUD em modais). (`db614fc`)
- **Estoque** — contexto de ajuste de estoque com domínio, serviço, repositórios
  PostgreSQL, rotas HTTP com RBAC e testes unitários de domínio; página
  `AjustesEstoquePage` com seleção de produto, criação de ajuste, histórico de
  movimentações e paginação. (`6ecd527`)
- **Compras** — contexto completo de compras com domínio, validações, serviço,
  repositórios e API HTTP; integrações cross-module com catálogo, estoque e
  fornecedores (checagem de produto, atualização de saldo e metadados do
  fornecedor); página `ComprasPage`. (`cddbc55`)
- **Vendas** — fluxo de venda completo (domínio, serviços, repositórios e API),
  páginas de PDV e listagem; emissão de documento fiscal resiliente com
  retry/circuit breaker; integrações cross-module: baixa atômica de estoque,
  validação de saldo no catálogo e atualização da data da última compra do
  cliente. (`a1fceb8`)
- **Relatórios** — contexto de relatórios (domínio, repositório, serviço e
  endpoints HTTP com RBAC); página `RelatoriosPage` com estoque abaixo do mínimo,
  produtos mais/menos vendidos e resumos de vendas/compras. (`39be0df`)

#### Frontend

- Kit de UI compartilhado (`PageShell`, `Button`, `DataTable`, `StatusBadge`,
  `Modal`, `Field`/`inputClasses`) com refatoração de todas as páginas para usá-lo;
  `ClientesPage` e rota `/clientes`. (`a00b794`)
- Suporte completo a **Dark/Light mode** com preferência persistida e fallback do
  sistema; primitivos padronizados `Sidebar`, `CommandPalette` (atalho ⌘K), `Tabs`
  e componentes polidos de formulário/tabela/modal; notificações via `sonner`;
  Tailwind com fontes Inter e JetBrains Mono. (`090d38e`)

#### Infraestrutura e deploy

- Infraestrutura de deploy em produção: `railway.json` para frontend e backend,
  Dockerfile do backend gerando binários da API e do migrate, healthcheck
  `/health` no nginx do frontend, arquivos de env de produção, runner de migração
  embutido e `supabase-setup.sh` para provisionamento do banco remoto. (`65467f6`)
- Middleware de **CORS** configurável via `ALLOWED_ORIGINS`; `eslint-plugin-react-hooks`
  no frontend com regras recomendadas. (`707db0e`)

### Alterado

- Setup de desenvolvimento via Docker: carregamento correto de arquivos de env,
  build arg `VITE_API_BASE_URL` no Dockerfile do frontend e `ALLOWED_ORIGINS` para
  a API; documentação passa a recomendar `make up`. (`0f49429`)
- Migration idempotente `000010_seed_demo` com dados de exemplo para todos os
  módulos. (`a00b794`)

### Corrigido

- Erros 404 na API por falta do prefixo `/api/v1` nas chamadas do frontend;
  rota de estoque ajustada para `/estoque/ajustes`. (`a00b794`)
- Avisos de lint do frontend com supressões para padrões intencionais de
  `useEffect` de data-fetching. (`707db0e`)

## [0.1.0] — 2026-06-25

Fundação do projeto: scaffolding do monólito hexagonal, infraestrutura Docker,
observabilidade, hardening de segredos e documentação base (anterior ao MVP).

### Adicionado

- Setup inicial do ERP de estoque: backend Go hexagonal, frontend React + Vite +
  TypeScript, orquestração Docker, migrations, Makefile e configuração. (`5145191`)
- Stack completa de observabilidade **OpenTelemetry**: instrumentação do servidor
  HTTP, métricas Prometheus em `/metrics`, `docker-compose.observability.yml` com
  Prometheus e Grafana, datasource provisionado e testes. Tracing inativo por
  padrão até `OTEL_EXPORTER_OTLP_ENDPOINT` ser definido. (`b2b3cb1`)
- Documentação completa do projeto: `mandates.md` (Definition of Done D1–D3 e
  protocolo do agente juiz), `todos.md` com checklists por fase, expansão do
  `docs/README.md`, overhaul do README raiz e template de release. (`9b85cb6`)
- Guias de setup em `docs/setup/` (backend, banco, frontend, Supabase, Railway),
  script de teardown do Docker, suíte de fitness functions, checklist de segurança
  e runbook de circuit breaker; alvo `check-secrets` no Makefile. (`7562391`)

### Alterado

- Bump de Go 1.23 → 1.25 no Dockerfile do backend e na documentação; badge do
  README atualizado e docs de observabilidade/teardown. (`23d3143`, `5d5583c`)
- Overhaul do `scripts/docker/teardown.sh`: flags `--recreate-volume`, `--obs` e
  `--help`, fallback para comandos Docker diretos, teardown da stack de
  observabilidade, recriação do volume Postgres e logging colorido. (`5d5583c`)
- Configs de build do frontend e Docker: `pnpm-workspace.yaml` com campo
  `packages`, preparação do pnpm 9 no Dockerfile e flag de env file no
  docker-compose via Makefile. (`2ee0be5`)

### Segurança

- Remoção de segredos hardcoded de Makefile, `.env.example`, config do backend e
  `docker-compose.yml`; checagens de env obrigatórias substituindo fallbacks para
  evitar credenciais de produção acidentais. (`b802ab8`)
- Tratamento de segredos no docker-compose via interpolação de env obrigatória,
  guia de secret scanning e itens de checklist pré-deploy. (`a55d305`)

[1.3.0]: #130--2026-07-01
[1.2.0]: #120--2026-07-01
[1.1.0]: #110--2026-06-26
[1.0.0]: #100--2026-06-26
[0.1.0]: #010--2026-06-25
