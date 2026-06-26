# Lista de Verificação — Implementação do ERP

Estado atual: migrations 000001–000009 prontas, plataforma (auth JWT, httpserver,
resilience, config, database) pronta, módulo `clientes` totalmente implementado,
camada `lib/` do frontend pronta, `LoginPage` e `DashboardPage` existem.

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

- [ ] Criar `internal/modules/vendas/domain/` — entidades `Venda`, `DetalheVenda` (XOR cliente/consumidor), erros (`SaldoInsuficiente`)
- [ ] Criar `ports/inbound.go` — `VendaService` (CriarVenda, AdicionarItem, ConfirmarVenda, ListarVendas, BuscarVenda)
- [ ] Criar `ports/outbound.go` — `VendaRepository`, `CatalogoReader`, `EstoqueWriter`, `FiscalGateway`
- [ ] Criar `adapters/outbound/fiscal/` — `FiscalGateway` (cupom/NF via API externa)
  - [ ] Implementar `EmitirCupom(venda)` e `EmitirNF(venda)` com resilience Policy (retry + circuit breaker)
- [ ] Criar `application/service.go` — caso de uso `ConfirmarVenda` em transação: valida saldo via `CatalogoReader`, baixa saldo via `EstoqueWriter` (`UPDATE WHERE estoque_a_pro >= qtd`), persiste venda, emite documento via `FiscalGateway`, atualiza `dt_ult_comp_cli`
- [ ] Garantir que saldo negativo é impossível (constraint no banco + lock pessimista)
- [ ] Criar `adapters/inbound/http/handler.go` + `router.go` — RBAC: `vendas:read`, `vendas:write`
- [ ] Criar `module.go` e montar em `main.go`
- [ ] Testar: venda simultânea com saldo = 1 unidade — apenas uma deve ser confirmada

### Frontend

- [ ] Criar `pages/VendasPage.tsx` — listagem de vendas
- [ ] Criar `pages/NovaVendaPage.tsx` (PDV):
  - [ ] Buscar cliente por CPF (preencher automaticamente ou abrir cadastro rápido)
  - [ ] Adicionar itens com validação de saldo em tempo real
  - [ ] Aplicar desconto
  - [ ] Confirmar venda e exibir documento fiscal emitido
- [ ] Integrar com API (`GET /vendas`, `POST /vendas`, `POST /vendas/{id}/confirmar`)

---

## Fase 7 — Relatórios

> Necessário para o critério de aceitação do cliente.

### Backend

- [ ] Criar endpoints de relatório (podem ser `GET` específicos em módulos existentes ou num módulo `relatorios` dedicado):
  - [ ] `GET /relatorios/produtos/abaixo-do-minimo` — RBAC: `relatorios:read`
  - [ ] `GET /relatorios/produtos/mais-vendidos?periodo=` — RBAC: `relatorios:read`
  - [ ] `GET /relatorios/produtos/menos-vendidos?periodo=` — RBAC: `relatorios:read`
  - [ ] `GET /relatorios/vendas?de=&ate=` — resumo por período
  - [ ] `GET /relatorios/compras?de=&ate=` — resumo por período
- [ ] Registrar rotas em `main.go`

### Frontend

- [ ] Criar `pages/RelatoriosPage.tsx` — navegação entre relatórios
- [ ] Exibir tabela de produtos abaixo do mínimo (destacar em vermelho)
- [ ] Exibir tabela de mais/menos vendidos por período
- [ ] Exibir resumo de vendas e compras por período
- [ ] Integrar com API de relatórios

---

## Fase 8 — Qualidade e Segurança

- [ ] Rodar `make be-vet` sem erros
- [ ] Rodar `make be-test` — todos os testes passam (domínio já tem `cliente_test.go`; adicionar testes para os novos domínios)
- [ ] Rodar `pnpm tsc --noEmit` no frontend sem erros
- [ ] Rodar `pnpm lint` no frontend sem erros
- [ ] Revisar checklist de segurança em `docs/reference/checklist.md`
- [ ] Verificar que `JWT_SECRET` é longo e aleatório (não está no `.env.example`)
- [ ] Verificar CORS: `ALLOWED_ORIGINS` lista apenas as origens legítimas em produção
- [ ] Confirmar que `admin@loja.local` / `admin123` está trocado em produção

---

## Fase 9 — Deploy (Railway + Supabase)

- [ ] Criar projeto no Supabase; obter `DATABASE_URL` (porta 5432, `sslmode=require`)
- [ ] Criar projeto no Railway com dois serviços: `erp-estoque-backend` e `erp-estoque-frontend`
- [ ] Configurar variáveis de ambiente no Railway para o backend (`DATABASE_URL`, `JWT_SECRET`, `ALLOWED_ORIGINS`, etc.)
- [ ] Configurar variável `VITE_API_BASE_URL` no serviço de frontend **antes** do primeiro build
- [ ] Definir pre-deploy command do backend: `/app/migrate up`
- [ ] Fazer deploy; verificar `/health` do backend
- [ ] Verificar migrations aplicadas no Supabase (tables nos schemas corretos)
- [ ] Testar ciclo completo em produção: login → compra → venda → ajuste → relatório

---

## Critério de Aceitação (conforme brief do cliente)

- [ ] Receber mercadoria → registrar compra → saldo atualizado automaticamente
- [ ] Atender cliente → registrar venda → saldo baixado → documento fiscal emitido
- [ ] Identificar ruptura → relatório de produtos abaixo do mínimo
- [ ] Lançar ajuste com motivo → registro imutável no histórico
- [ ] Emitir relatório de vendas e compras do período
- [ ] Operação de um dia completo sem planilha, sem saldo negativo, sem erro fiscal
