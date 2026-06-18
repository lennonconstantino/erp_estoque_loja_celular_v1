# PROMPT.md — ERP Estoque Loja de Acessórios para Celular

> Documento de especificação executável. Leia §0 inteiro antes de escrever
> qualquer linha de código. Implemente uma fase de cada vez na ordem indicada.
> Todo critério de aceitação deve dar `PASS` antes de avançar.

---

## §0 — Arquitetura (LEIA, NÃO IMPLEMENTE)

### Contexto de negócio

ERP web para loja de acessórios de celular (capas, películas, carregadores,
cabos). Equipe de ~5 pessoas: `ADMIN`, `VENDEDOR`, `ESTOQUISTA`. A operação
hoje usa planilhas — o sistema substitui isso com controle de estoque em tempo
real, ciclo compra→venda→fiscal completo e relatórios operacionais.

**Critério de aceitação do cliente:** operação de um dia completo (compras,
vendas, ajustes) executada sem planilha, sem saldo negativo, sem erro fiscal.

### Stack

| Camada | Tecnologia |
|--------|------------|
| Backend | Go, `go-chi/chi`, `pgx/v5`, `golang-migrate`, `golang-jwt/jwt`, `bcrypt` |
| Banco | PostgreSQL 16 (local: docker-compose; produção: Supabase porta 5432) |
| Frontend | React 18 + Vite + TypeScript strict, React Router, Tailwind, shadcn/ui, **pnpm** |
| Deploy | Railway (backend + frontend) + Supabase (PostgreSQL) |

### Leis arquiteturais (L — nunca violar)

```
L1  Um módulo NUNCA importa o pacote interno de outro módulo.
    Comunicação cross-module: declara porta em ports/outbound.go e recebe
    a implementação por injeção no main.

L2  Dependências apontam sempre para DENTRO do módulo:
    adapters → application → ports → domain
    O domain não importa nada de infra (sem pgx, sem net/http).

L3  Sem FK entre schemas no banco. UUIDs de outros contextos são "soltos".
    Integridade cross-context é responsabilidade da aplicação.

L4  Toda rota protegida usa authMgr.Authenticate + auth.RequirePerm("recurso:acao").
    Nenhuma rota protegida sem RBAC.

L5  Saldo de produto NUNCA fica negativo. Toda baixa usa UPDATE ... WHERE
    estoque_a_pro >= $qtd dentro de uma transação.

L6  estoque.movimentacoes e estoque.ajustes são append-only.
    UPDATE/DELETE bloqueados por trigger (já na migration 000008).

L7  Frontend fala APENAS com o backend Go via JSON + Bearer JWT.
    Sem acesso direto ao banco, sem @supabase/supabase-js, sem SSR/Next.js.

L8  Código, comentários e nomes de variáveis/funções em PORTUGUÊS.
```

### Módulos e dependências entre contextos

```
iam          (nenhuma dep. de outro módulo)
clientes     → CepGateway (ViaCEP)
fornecedores → CepGateway (ViaCEP)
catalogo     → EstoqueReader   (implementado por estoque)
estoque      → CatalogoWriter  (implementado por catalogo)
compras      → EstoqueWriter   (implementado por estoque)
             → CatalogoReader  (implementado por catalogo)
vendas       → EstoqueWriter, CatalogoReader, FiscalGateway
relatorios   → leitura direta no banco (sem deps de outros módulos)
```

### Razão de estoque (ledger + saldo materializado)

- `estoque.movimentacoes` = fonte da verdade (append-only, tipos: COMPRA,
  VENDA, AJUSTE_ENTRADA, AJUSTE_SAIDA). Cada linha guarda `saldo_ant`/`saldo_atu`.
- `catalogo.produtos.estoque_a_pro` = cache. Atualizado via `CatalogoWriter`
  a cada movimentação.
- `disp_pro = (estoque_a_pro > 0)` — recalculado junto.

### Fluxo de autenticação

```
POST /api/v1/auth/login  →  valida bcrypt → emite JWT (15min) + refresh (30d)
POST /api/v1/auth/refresh →  confere hash + validade → novo access + rotaciona refresh
POST /api/v1/auth/logout  →  revoga refresh (campo revogado=true)

Claims do JWT: sub, roles, perms, exp
Permissões já embutidas no token — SEM consulta ao banco por request.
```

### RBAC — papéis e permissões

| Papel | Permissões |
|-------|------------|
| `ADMIN` | todas (`iam:admin` incluso) |
| `VENDEDOR` | `vendas:*`, `clientes:*`, `catalogo:read`, `estoque:read`, `relatorios:read` |
| `ESTOQUISTA` | `compras:*`, `estoque:*`, `catalogo:*`, `fornecedores:*`, `relatorios:read` |

### Tratamento de erros (backend)

Envelope: `{"error":{"code":"...","message":"..."}}`.

| HTTP | Quando |
|------|--------|
| 401 | token ausente/expirado/inválido |
| 403 | sem permissão |
| 404 | não encontrado |
| 409 | invariante violada (CPF duplicado, saldo insuficiente) |
| 422 | payload inválido |
| 502 | falha em gateway externo (CEP, fiscal) |
| 500 | erro interno |

### Estado atual do repositório

**Pronto — não reimplementar:**

- Migrations `000001`–`000009` (todos os schemas e seed)
- `backend/internal/platform/` completo: `auth/`, `config/`, `database/`,
  `httpserver/`, `resilience/`
- Módulo `clientes` completo (referência canônica do padrão hexagonal)
- `frontend/src/lib/`: `api.ts`, `auth.ts`, `env.ts`, `http.ts`, `utils.ts`
- `frontend/src/pages/LoginPage.tsx`, `DashboardPage.tsx`

**Falta implementar (nesta ordem):** módulos `iam`, `fornecedores`, `catalogo`,
`estoque`, `compras`, `vendas`, `relatorios` no backend; páginas/componentes
no frontend; configuração de deploy.

---

## §0.5 — Definition of Done por Tarefa (D — obrigatório)

> **Mandate de execução.** Toda tarefa proveniente de um plano (uma fase F0–F9 ou
> um item de `docs/todos.md`) só é concluída quando D1, D2 e D3 derem `PASS`. O
> `curl`/PASS de aceitação é o **piso**, não o teto. Detalhe vinculante,
> protocolo do juiz e templates em **[docs/mandates.md](docs/mandates.md)**.

```
D1  TESTES NO MESMO PASSO  — testes unitários entregues junto da lógica de
    domínio/caso de uso (não na Fase 8). Verificar:
        cd backend && go test -cover ./internal/modules/<dominio>/...
    Metas: domain/ ≥ 80%, application/ ≥ 70%. Frontend: tsc --noEmit + lint.

D2  CHECKLIST VIVO        — derive o checklist da tarefa antes de codar e marque
    [x] em docs/todos.md os itens da fase à medida que conclui e verifica.

D3  AGENTE JUIZ           — antes do PASS da fase, um subagent juiz INDEPENDENTE
    julga o diff contra esta spec + Leis L1–L8 + Regras R1–R12 e emite
    CONFORME | NÃO CONFORME. NÃO CONFORME bloqueia o avanço de fase.

Ordem:  implementar → D1 → D2 → critério curl/PASS da fase → D3 (CONFORME)
```

---

## §1 — Regras de Execução (R — siga sempre)

```
R1  Leia o arquivo existente ANTES de editar. Nunca sobrescreva sem ler.

R2  Use internal/modules/clientes/ como referência canônica ao criar módulos.
    Replique a estrutura: domain/ ports/ application/ adapters/ module.go.

R3  Confirme conformidade em tempo de compilação em application/service.go:
    var _ ports.XService = (*Service)(nil)

R4  Erros são sentinelas de domínio (domain/errors.go).
    O handler os traduz para HTTP em writeDomainError. Nunca propague
    status HTTP fora do adaptador HTTP.

R5  Após criar ou editar qualquer arquivo .go, compile:
    cd backend && go build ./...
    Se falhar, pare e corrija antes de criar o próximo arquivo.

R6  Sem stubs, sem TODO, sem panic("não implementado"). Toda função
    entregue deve funcionar corretamente.

R7  Handlers usam DTOs próprios (xyzRequest / xyzResponse).
    Nunca serialize a entidade de domínio diretamente.

R8  Adaptadores outbound que chamam APIs externas (CEP, fiscal) DEVEM
    embrulhar a chamada em uma resilience.Policy (Retry + CB + Bulkhead).
    Veja clientes/adapters/outbound/cep/viacep.go como referência.

R9  Novo módulo = nova linha em cmd/api/main.go.
    Descomente o mount correspondente ou adicione-o.

R10 Frontend: use APENAS pnpm. Sem npm install / yarn add.
    Sem axios, lodash, moment, @supabase/supabase-js, Next.js.
    HTTP via @/lib/api. Validação de env via @/lib/env.ts.

R11 TypeScript strict: sem `any`. Use `unknown` + narrowing quando necessário.

R12 Critério de aceitação de cada fase = comando shell que imprime PASS.
    Execute-o ao concluir a fase. Só avance se der PASS.
```

---

## §2 — Estrutura de Diretórios Esperada

```
. (raiz)
├── PROMPT.md
├── CLAUDE.md
├── Makefile
├── docker-compose.yml
├── docs/
├── backend/
│   ├── cmd/
│   │   ├── api/main.go          ← monta todos os módulos
│   │   └── migrate/main.go
│   ├── internal/
│   │   ├── platform/            ← PRONTO (auth, config, database, httpserver, resilience)
│   │   └── modules/
│   │       ├── clientes/        ← PRONTO (referência canônica)
│   │       ├── iam/             ← F1
│   │       ├── fornecedores/    ← F2
│   │       ├── catalogo/        ← F3
│   │       ├── estoque/         ← F4
│   │       ├── compras/         ← F5
│   │       ├── vendas/          ← F6
│   │       └── relatorios/      ← F7
│   └── migrations/              ← PRONTAS (000001–000009)
└── frontend/
    └── src/
        ├── lib/                 ← PRONTO (api, auth, env, http, utils)
        ├── components/
        │   └── ui/              ← primitivos shadcn
        ├── pages/
        │   ├── LoginPage.tsx    ← PRONTO (completar lógica de submit)
        │   ├── DashboardPage.tsx
        │   ├── ClientesPage.tsx      ← F1-FE
        │   ├── FornecedoresPage.tsx  ← F2-FE
        │   ├── CategoriasPage.tsx    ← F3-FE
        │   ├── ProdutosPage.tsx      ← F3-FE
        │   ├── EstoquePage.tsx       ← F4-FE
        │   ├── ComprasPage.tsx       ← F5-FE
        │   ├── VendasPage.tsx        ← F6-FE
        │   └── RelatoriosPage.tsx    ← F7-FE
        └── App.tsx                   ← rotas + proteção JWT
```

### Anatomia obrigatória de cada módulo backend

```
internal/modules/<dominio>/
├── domain/
│   ├── <entidade>.go      # struct + New() + Validar() + invariantes
│   └── errors.go          # var ErrXxx = errors.New("...")
├── ports/
│   ├── inbound.go         # interface XService + structs Input
│   └── outbound.go        # interface XRepository + gateways externos
├── application/
│   ├── service.go         # Service implementa ports.XService
│   └── dto.go             # tipos de entrada/saída internos (opcional)
├── adapters/
│   ├── inbound/http/
│   │   ├── handler.go     # Handler{svc ports.XService} + writeDomainError
│   │   └── router.go      # NewRouter(h, authMgr) com RBAC
│   └── outbound/
│       └── postgres/
│           └── x_repo.go  # implementa ports.XRepository
└── module.go              # New(...) → instancia repo, service, handler, router
```

---

## §3 — Fase 0: Ambiente Local

Objetivo: Docker Compose sobe, migrations aplicadas, login funciona.

**Tarefas:**

1. Copiar `backend/.env.example` → `backend/.env`; preencher `DATABASE_URL`,
   `JWT_SECRET` (mín. 32 chars aleatórios), `CEP_API_URL=https://viacep.com.br`
2. `docker compose up -d`
3. Verificar health

**Critério de aceitação:**

```bash
curl -s http://localhost:8080/health | grep -q "ok" && echo "PASS F0.health"
# POST /auth/login validado ao final de F1
```

---

## §4 — Fase 1: Módulo IAM (autenticação + usuários)

Objetivo: login, refresh, logout e gestão de usuários funcionando.

### Backend — F1

**Arquivos a criar:**

`internal/modules/iam/domain/usuario.go`
```go
// Entidade Usuario com hash de senha (bcrypt). Nunca exponha SenhaHash.
// New() valida email e papel obrigatórios.
```

`internal/modules/iam/domain/errors.go`
```go
var (
    ErrCredenciaisInvalidas = errors.New("email ou senha inválidos")
    ErrUsuarioNaoEncontrado = errors.New("usuário não encontrado")
    ErrEmailJaCadastrado    = errors.New("email já cadastrado")
    ErrRefreshInvalido      = errors.New("refresh token inválido ou expirado")
)
```

`internal/modules/iam/ports/inbound.go`
```go
type AuthService interface {
    Login(ctx context.Context, email, senha string) (LoginOutput, error)
    Refresh(ctx context.Context, refreshToken string) (LoginOutput, error)
    Logout(ctx context.Context, refreshToken string) error
}
type UsuarioService interface {
    Criar(ctx context.Context, input CriarUsuarioInput) (*domain.Usuario, error)
    Listar(ctx context.Context) ([]domain.Usuario, error)
    AlterarSenha(ctx context.Context, id uuid.UUID, novaSenha string) error
}
type LoginOutput struct {
    AccessToken  string
    RefreshToken string
}
```

`internal/modules/iam/ports/outbound.go`
```go
type UsuarioRepository interface {
    Criar(ctx context.Context, u *domain.Usuario) error
    BuscarPorEmail(ctx context.Context, email string) (*domain.Usuario, error)
    BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Usuario, error)
    Listar(ctx context.Context) ([]domain.Usuario, error)
    AtualizarSenhaHash(ctx context.Context, id uuid.UUID, hash string) error
    SalvarRefreshToken(ctx context.Context, usuarioID uuid.UUID, tokenHash string, expiracao time.Time) error
    BuscarRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenRecord, error)
    RevogarRefreshToken(ctx context.Context, tokenHash string) error
    RotacionarRefreshToken(ctx context.Context, antigoHash, novoHash string, novaExp time.Time) error
}
type RefreshTokenRecord struct {
    UsuarioID uuid.UUID
    Expiracao time.Time
    Revogado  bool
}
```

`internal/modules/iam/application/service.go`
- `Login`: busca usuário por email, `bcrypt.CompareHashAndPassword`, resolve
  papéis→permissões, emite access JWT via `authMgr.GenerateAccess`,
  gera refresh token opaco (UUID), salva hash no banco.
- `Refresh`: busca refresh pelo hash, valida expiração e `!Revogado`, rotaciona.
- `Logout`: revoga refresh token.
- `var _ ports.AuthService = (*Service)(nil)`

`internal/modules/iam/adapters/outbound/postgres/usuario_repo.go`
- Queries em `iam.usuarios`, `iam.papeis`, `iam.permissoes`, `iam.usuario_papeis`,
  `iam.papel_permissoes`, `iam.refresh_tokens`.

`internal/modules/iam/adapters/inbound/http/handler.go`
- `POST /auth/login` (público), `POST /auth/refresh` (público),
  `POST /auth/logout` (autenticado)
- `GET /usuarios`, `POST /usuarios`, `PATCH /usuarios/{id}/senha`
  (todos: RequirePerm("iam:admin"))

`internal/modules/iam/module.go`
```go
func New(pool *pgxpool.Pool, authMgr *auth.Manager) *Module
func (m *Module) AuthRouter() chi.Router
func (m *Module) UsuarioRouter() chi.Router
```

`cmd/api/main.go` — adicionar:
```go
iamMod := iam.New(pool, authMgr)
// no r.Route("/api/v1"):
api.Mount("/auth",     iamMod.AuthRouter())
api.Mount("/usuarios", iamMod.UsuarioRouter())
```

**Critério de aceitação F1-BE:**

```bash
cd backend && go build ./... && echo "PASS F1.build"

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@loja.local","senha":"admin123"}' \
  | jq -r '.access_token')
[ -n "$TOKEN" ] && [ "$TOKEN" != "null" ] && echo "PASS F1.login"

curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/usuarios | grep -q "items" && echo "PASS F1.usuarios"
```

### Frontend — F1-FE

`frontend/src/pages/LoginPage.tsx` — completar lógica de submit:
- Chamar `api.post('/auth/login', {email, senha})`
- Salvar tokens via funções de `@/lib/auth.ts`
- Redirecionar para `/dashboard`
- Exibir erro em caso de credenciais inválidas

`frontend/src/App.tsx` — implementar:
- Rota pública: `/login`
- Rotas protegidas: todas as demais (redirecionar para `/login` se sem token)
- Componente `RotaProtegida` que verifica `auth.getAccessToken()`

```bash
cd frontend && pnpm tsc --noEmit && echo "PASS F1.ts"
```

---

## §5 — Fase 2: Módulo Fornecedores

Objetivo: CRUD de fornecedores com validação de CNPJ e CEP automático.
Segue exatamente o molde do módulo `clientes`.

### Backend — F2

**Domínio:**
- `Fornecedor`: CNPJ (14 dígitos, único), RazaoSocial, NomeFantasia, Email,
  Telefone1 (obrigatórios), Telefone2, ContatoComercial, ContatoFinanceiro,
  CEP + endereço, `DtUltCompFor *time.Time`.
- Invariante CNPJ: validação de dígitos verificadores.
- Erros: `ErrCNPJInvalido`, `ErrCNPJJaCadastrado`, `ErrNaoEncontrado`,
  `ErrRazaoSocialObrigatoria`.

**Portas de saída:** `FornecedorRepository`, `CepGateway` (interface local idêntica
à de `clientes` — declare em `ports/outbound.go`; injete `cep.NewGateway` no `module.go`).

**Endpoints:** `GET /fornecedores`, `GET /fornecedores/{id}`,
`POST /fornecedores`, `PUT /fornecedores/{id}`, `DELETE /fornecedores/{id}`,
`GET /fornecedores/cep/{cep}`.
RBAC: `fornecedores:read` / `fornecedores:write`.

**Critério de aceitação F2-BE:**

```bash
cd backend && go build ./... && echo "PASS F2.build"

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@loja.local","senha":"admin123"}' | jq -r '.access_token')

curl -s -X POST http://localhost:8080/api/v1/fornecedores \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"cnpj":"11222333000181","razao_social":"Dist. Teste LTDA",
       "nome_fantasia":"Dist. Teste","email":"contato@dist.com",
       "telefone1":"11999990000","contato_comercial":"João"}' \
  | grep -q "id" && echo "PASS F2.criar"
```

### Frontend — F2-FE

`FornecedoresPage.tsx`:
- Tabela com CNPJ, Razão Social, Contato, ações (editar, excluir)
- Modal de cadastro/edição com preenchimento automático de endereço via CEP

```bash
cd frontend && pnpm tsc --noEmit && echo "PASS F2.ts"
```

---

## §6 — Fase 3: Módulo Catálogo (Categorias + Produtos)

Objetivo: CRUD de categorias e produtos; expor `CatalogoReader` e
`CatalogoWriter` para injeção nos módulos seguintes.

### Backend — F3

**Domínio Categoria:** `Descricao` único. Erro: `ErrDescricaoJaCadastrada`.

**Domínio Produto:**
- Campos: `CategoriaID`, `Descricao`, `Custo`, `Venda`, `EstoqueA` (não
  editável pelo cadastro), `EstoqueMin`, `DispPro bool`, `Ativo bool`.
- Invariantes: `Custo < Venda` (erro `ErrMargemInvalida`); `EstoqueA >= 0`.
- Margem % = `(Venda-Custo)/Custo*100` — campo derivado na response.

**Interfaces que este módulo expõe para injeção em outros módulos:**

```go
// Declare em catalogo/ports/outbound.go (ou em arquivo público do módulo).
// Os módulos que as consomem declaram interfaces locais idênticas —
// Go satisfaz por duck typing; o main injeta a implementação de catalogo.

type CatalogoReader interface {
    BuscarProduto(ctx context.Context, id uuid.UUID) (*Produto, error)
    ValidarSaldo(ctx context.Context, produtoID uuid.UUID, qtd int) error
}
type CatalogoWriter interface {
    AtualizarSaldo(ctx context.Context, produtoID uuid.UUID, novoSaldo int) error
    AtualizarDisponibilidade(ctx context.Context, produtoID uuid.UUID, disp bool) error
}
```

`module.go` deve expor:
```go
func (m *Module) Reader() CatalogoReader
func (m *Module) Writer() CatalogoWriter
```

**Endpoints:**
- `GET|POST /categorias`, `PUT|DELETE /categorias/{id}` — `catalogo:read/write`
- `GET|POST /produtos`, `GET|PUT|DELETE /produtos/{id}` — `catalogo:read/write`

**Critério de aceitação F3-BE:**

```bash
cd backend && go build ./... && echo "PASS F3.build"

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@loja.local","senha":"admin123"}' | jq -r '.access_token')

CAT=$(curl -s -X POST http://localhost:8080/api/v1/categorias \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"descricao":"Cabos"}' | jq -r '.id')
[ -n "$CAT" ] && [ "$CAT" != "null" ] && echo "PASS F3.categoria"

curl -s -X POST http://localhost:8080/api/v1/produtos \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"descricao\":\"Cabo USB-C\",\"categoria_id\":\"$CAT\",
       \"custo\":5.00,\"venda\":15.00,\"estoque_min\":10}" \
  | grep -q "id" && echo "PASS F3.produto"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  http://localhost:8080/api/v1/produtos \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"descricao\":\"Inválido\",\"categoria_id\":\"$CAT\",
       \"custo\":20.00,\"venda\":10.00}")
[ "$STATUS" = "422" ] && echo "PASS F3.invariante"
```

### Frontend — F3-FE

`CategoriasPage.tsx`: tabela + modal CRUD.
`ProdutosPage.tsx`: tabela com saldo atual, margem % calculada, destaque
vermelho quando `estoque_a < estoque_min`.

```bash
cd frontend && pnpm tsc --noEmit && echo "PASS F3.ts"
```

---

## §7 — Fase 4: Módulo Estoque (Razão + Ajustes)

Objetivo: razão append-only, ajustes com motivo, porta `EstoqueWriter` para
uso por compras e vendas.

### Backend — F4

**Domínio:**

```go
type TipoMovimentacao string
const (
    Compra        TipoMovimentacao = "COMPRA"
    Venda         TipoMovimentacao = "VENDA"
    AjusteEntrada TipoMovimentacao = "AJUSTE_ENTRADA"
    AjusteSaida   TipoMovimentacao = "AJUSTE_SAIDA"
)

type Movimentacao struct {
    ID            uuid.UUID
    ProdutoID     uuid.UUID
    Tipo          TipoMovimentacao
    Quantidade    int
    SaldoAnterior int
    SaldoAtual    int
    ReferenciaID  uuid.UUID
    CriadoEm     time.Time
}

type Ajuste struct {
    ID         uuid.UUID
    ProdutoID  uuid.UUID
    Tipo       TipoMovimentacao
    Quantidade int
    Motivo     string
    UsuarioID  uuid.UUID
    CriadoEm  time.Time
}
```

**Porta inbound:**
```go
type EstoqueService interface {
    RegistrarMovimentacao(ctx context.Context, input RegistrarMovInput) (*Movimentacao, error)
    LancarAjuste(ctx context.Context, input LancarAjusteInput) (*Ajuste, error)
    ListarMovimentacoes(ctx context.Context, produtoID uuid.UUID) ([]Movimentacao, error)
    ListarAjustes(ctx context.Context, produtoID uuid.UUID) ([]Ajuste, error)
    SaldoAtual(ctx context.Context, produtoID uuid.UUID) (int, error)
}
```

**Interface que este módulo expõe para compras e vendas:**
```go
// EstoqueWriter — implementada por estoque, injetada em compras e vendas.
type EstoqueWriter interface {
    RegistrarMovimentacao(ctx context.Context, input RegistrarMovInput) (*Movimentacao, error)
}
```

**Implementação de `RegistrarMovimentacao` (em transação):**
1. Busca saldo atual do produto.
2. Calcula novo saldo (`ant ± qtd`).
3. Valida `novoSaldo >= 0` — `ErrSaldoInsuficiente` se falhar.
4. Persiste `Movimentacao`.
5. Chama `CatalogoWriter.AtualizarSaldo` + `AtualizarDisponibilidade`.

**`module.go`** deve expor:
```go
func (m *Module) Writer() EstoqueWriter
```

**Endpoints:**
- `POST /estoque/ajustes` — `estoque:write`
- `GET /estoque/ajustes?produto_id=` — `estoque:read`
- `GET /estoque/movimentacoes?produto_id=` — `estoque:read`

**Critério de aceitação F4-BE:**

```bash
cd backend && go build ./... && echo "PASS F4.build"

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@loja.local","senha":"admin123"}' | jq -r '.access_token')

PROD_ID="<uuid do produto criado em F3>"

curl -s -X POST http://localhost:8080/api/v1/estoque/ajustes \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"produto_id\":\"$PROD_ID\",\"tipo\":\"AJUSTE_ENTRADA\",
       \"quantidade\":50,\"motivo\":\"Estoque inicial\"}" \
  | grep -q "id" && echo "PASS F4.ajuste"

curl -s "http://localhost:8080/api/v1/estoque/movimentacoes?produto_id=$PROD_ID" \
  -H "Authorization: Bearer $TOKEN" \
  | grep -q "AJUSTE_ENTRADA" && echo "PASS F4.ledger"
```

### Frontend — F4-FE

`EstoquePage.tsx`:
- Formulário de ajuste: produto (autocomplete), tipo, quantidade, motivo
- Tabela de ajustes — somente leitura (sem editar/excluir)
- Tabela de movimentações com `saldo_ant` → `saldo_atu`

```bash
cd frontend && pnpm tsc --noEmit && echo "PASS F4.ts"
```

---

## §8 — Fase 5: Módulo Compras

Objetivo: entrada de mercadoria com NF; confirmar compra atualiza estoque.

### Backend — F5

**Domínio:**
```go
type StatusCompra string
const (
    Rascunho   StatusCompra = "RASCUNHO"
    Confirmada StatusCompra = "CONFIRMADA"
)

type Compra struct {
    ID           uuid.UUID
    FornecedorID uuid.UUID
    NumeroNF     string
    Status       StatusCompra
    Itens        []DetalheCompra
    Total        float64
    CriadoEm    time.Time
    ConfirmadoEm *time.Time
}
type DetalheCompra struct {
    ID         uuid.UUID
    ProdutoID  uuid.UUID
    Quantidade int     // > 0
    Custo      float64
    Venda      float64 // custo < venda
}
```

**Caso de uso `ConfirmarCompra` (em transação única):**
1. Valida status `RASCUNHO`.
2. Para cada item: valida produto via `CatalogoReader.BuscarProduto`.
3. Persiste `compra_master` com status `CONFIRMADA`.
4. Para cada item: `EstoqueWriter.RegistrarMovimentacao(tipo=COMPRA, qtd, refID=compra.ID)`.
5. Opcionalmente atualiza preço de venda do produto via `CatalogoWriter`.
6. Atualiza `dt_ult_comp_for` do fornecedor no repo.

**Endpoints:**
- `GET /compras`, `GET /compras/{id}` — `compras:read`
- `POST /compras` — `compras:write`
- `POST /compras/{id}/confirmar` — `compras:write`

**Critério de aceitação F5-BE:**

```bash
cd backend && go build ./... && echo "PASS F5.build"

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@loja.local","senha":"admin123"}' | jq -r '.access_token')

PROD_ID="<uuid>" ; FOR_ID="<uuid>"

COMPRA=$(curl -s -X POST http://localhost:8080/api/v1/compras \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"fornecedor_id\":\"$FOR_ID\",\"numero_nf\":\"NF-001\",
       \"itens\":[{\"produto_id\":\"$PROD_ID\",\"quantidade\":20,
                   \"custo\":5.00,\"venda\":15.00}]}" | jq -r '.id')

curl -s -X POST "http://localhost:8080/api/v1/compras/$COMPRA/confirmar" \
  -H "Authorization: Bearer $TOKEN" \
  | grep -q "CONFIRMADA" && echo "PASS F5.confirmar"

curl -s "http://localhost:8080/api/v1/estoque/movimentacoes?produto_id=$PROD_ID" \
  -H "Authorization: Bearer $TOKEN" \
  | grep -q "COMPRA" && echo "PASS F5.estoque_atualizado"
```

### Frontend — F5-FE

`ComprasPage.tsx`: listagem de compras com status.
`NovaCompraPage.tsx` (ou modal largo):
- Selecionar fornecedor, número da NF
- Adicionar itens: produto (autocomplete), qtd, custo, venda
- Total calculado em tempo real
- Botão "Confirmar Compra"

```bash
cd frontend && pnpm tsc --noEmit && echo "PASS F5.ts"
```

---

## §9 — Fase 6: Módulo Vendas (PDV)

Objetivo: PDV com validação de saldo, baixa transacional e documento fiscal.
**Nunca permitir saldo negativo — esse é o invariante mais crítico do sistema.**

### Backend — F6

**Domínio:**
```go
type StatusVenda string
const (
    Aberta     StatusVenda = "ABERTA"
    Confirmada StatusVenda = "CONFIRMADA"
    Cancelada  StatusVenda = "CANCELADA"
)

type Venda struct {
    ID              uuid.UUID
    ClienteID       *uuid.UUID  // nulo = consumidor final
    ConsumidorFinal string      // XOR com ClienteID
    Status          StatusVenda
    Desconto        float64
    ValorTotal      float64
    DocFiscalURL    string
    Itens           []DetalheVenda
}
type DetalheVenda struct {
    ProdutoID  uuid.UUID
    Quantidade int
    Preco      float64
}
```

**Invariante XOR:** `(ClienteID != nil) XOR (ConsumidorFinal != "")`.
NF exige `ClienteID` preenchido — validar no caso de uso.

**Caso de uso `ConfirmarVenda` (em transação única):**
1. Valida status `ABERTA` e invariante XOR.
2. Para cada item: `CatalogoReader.ValidarSaldo(produtoID, qtd)` — `409` se falhar.
3. Grava `venda_master` + itens com status `CONFIRMADA`.
4. Para cada item: `EstoqueWriter.RegistrarMovimentacao(tipo=VENDA, qtd, refID=venda.ID)`.
   A baixa usa `UPDATE ... WHERE estoque_a_pro >= $qtd` — 0 linhas = rollback + `409`.
5. Chama `FiscalGateway.EmitirCupom(venda)` ou `EmitirNF(venda)`.
6. Salva URL do documento fiscal na venda.
7. Atualiza `dt_ult_comp_cli` do cliente (se houver).

**Adaptador `FiscalGateway`:**
```go
// adapters/outbound/fiscal/gateway.go
type Gateway struct {
    cupomURL string
    nfURL    string
    policy   resilience.Policy
}
func (g *Gateway) EmitirCupom(ctx context.Context, v *domain.Venda) (string, error)
func (g *Gateway) EmitirNF(ctx context.Context, v *domain.Venda) (string, error)
// Embrulhar chamadas HTTP na resilience.Policy (R8).
// Se a URL não estiver configurada (desenvolvimento), retornar URL fictícia.
```

**Endpoints:**
- `GET /vendas`, `GET /vendas/{id}` — `vendas:read`
- `POST /vendas` — `vendas:write`
- `POST /vendas/{id}/confirmar` — `vendas:write`
- `POST /vendas/{id}/cancelar` — `vendas:write`

**Critério de aceitação F6-BE (inclui teste de concorrência):**

```bash
cd backend && go build ./... && echo "PASS F6.build"

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@loja.local","senha":"admin123"}' | jq -r '.access_token')

# Pré-requisito: produto com saldo=1 (use ajuste de estoque se necessário)
PROD_ID="<uuid com saldo=1>" ; CLI_ID="<uuid>"

VENDA_A=$(curl -s -X POST http://localhost:8080/api/v1/vendas \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"cliente_id\":\"$CLI_ID\",
       \"itens\":[{\"produto_id\":\"$PROD_ID\",\"quantidade\":1,\"preco\":15.00}]}" \
  | jq -r '.id')
VENDA_B=$(curl -s -X POST http://localhost:8080/api/v1/vendas \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"cliente_id\":\"$CLI_ID\",
       \"itens\":[{\"produto_id\":\"$PROD_ID\",\"quantidade\":1,\"preco\":15.00}]}" \
  | jq -r '.id')

STATUS_A=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "http://localhost:8080/api/v1/vendas/$VENDA_A/confirmar" \
  -H "Authorization: Bearer $TOKEN") &
STATUS_B=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "http://localhost:8080/api/v1/vendas/$VENDA_B/confirmar" \
  -H "Authorization: Bearer $TOKEN") &
wait

echo "StatusA=$STATUS_A StatusB=$STATUS_B"
[[ ("$STATUS_A" == "200" && "$STATUS_B" == "409") ||
   ("$STATUS_A" == "409" && "$STATUS_B" == "200") ]] && echo "PASS F6.concorrencia"
```

### Frontend — F6-FE

`VendasPage.tsx`: listagem de vendas com status e link para documento fiscal.

`NovaVendaPage.tsx` (PDV):
- Buscar cliente por CPF → preencher ou botão "consumidor final"
- Buscar produto por código/nome → exibir preço e saldo disponível
- Validar `saldo > 0` ao adicionar item (aviso visual)
- Campo de desconto (aplicado no total)
- Botão "Confirmar Venda" → modal de confirmação → exibir URL do documento fiscal

```bash
cd frontend && pnpm tsc --noEmit && echo "PASS F6.ts"
```

---

## §10 — Fase 7: Relatórios

Objetivo: listagens operacionais que cobrem o critério de aceitação do cliente.

### Backend — F7

Módulo `relatorios` — sem domínio complexo; apenas queries agregadas.
RBAC `relatorios:read` em todos os endpoints.

**Endpoints:**

```
GET /relatorios/produtos/abaixo-do-minimo
    → [{id, descricao, estoque_a, estoque_min, diferenca}]

GET /relatorios/produtos/mais-vendidos?de=YYYY-MM-DD&ate=YYYY-MM-DD&limit=20
    → [{produto_id, descricao, total_vendido}]

GET /relatorios/produtos/menos-vendidos?de=&ate=&limit=20
    → [{produto_id, descricao, total_vendido}]

GET /relatorios/vendas?de=&ate=
    → {total_vendas, valor_total, ticket_medio, itens:[...]}

GET /relatorios/compras?de=&ate=
    → {total_compras, valor_total, itens:[...]}
```

**Critério de aceitação F7-BE:**

```bash
cd backend && go build ./... && echo "PASS F7.build"

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@loja.local","senha":"admin123"}' | jq -r '.access_token')

curl -s "http://localhost:8080/api/v1/relatorios/produtos/abaixo-do-minimo" \
  -H "Authorization: Bearer $TOKEN" | jq 'type' | grep -q "array" && echo "PASS F7.minimo"

curl -s "http://localhost:8080/api/v1/relatorios/vendas?de=2024-01-01&ate=2099-12-31" \
  -H "Authorization: Bearer $TOKEN" | grep -q "total_vendas" && echo "PASS F7.vendas"
```

### Frontend — F7-FE

`RelatoriosPage.tsx`:
- Aba "Abaixo do mínimo": tabela com destaque vermelho quando `diferenca < 0`
- Aba "Mais/menos vendidos": seletor de período + tabela ranqueada
- Aba "Vendas": seletor de período + resumo + tabela
- Aba "Compras": seletor de período + resumo + tabela

```bash
cd frontend && pnpm tsc --noEmit && echo "PASS F7.ts"
```

---

## §11 — Fase 8: Qualidade e Segurança

**Backend:**
```bash
cd backend && go vet ./... && echo "PASS F8.vet"
cd backend && go test ./... && echo "PASS F8.tests"
```

Testes mínimos por módulo (domínio puro, sem banco):
- `iam`: senha incorreta → `ErrCredenciaisInvalidas`; refresh revogado → erro
- `fornecedores`: CNPJ inválido → `ErrCNPJInvalido`; CNPJ duplicado → `ErrCNPJJaCadastrado`
- `catalogo`: `custo >= venda` → `ErrMargemInvalida`; `estoque_a < 0` → erro
- `estoque`: baixa maior que saldo → `ErrSaldoInsuficiente`
- `vendas`: venda sem cliente nem consumidor final → `ErrClienteOuConsumidorObrigatorio`

**Frontend:**
```bash
cd frontend && pnpm tsc --noEmit && echo "PASS F8.ts"
cd frontend && pnpm lint && echo "PASS F8.lint"
```

**Checklist de segurança pré-deploy:**
- [ ] `JWT_SECRET` ≥ 32 chars aleatórios (não derivado de senha humana)
- [ ] `ALLOWED_ORIGINS` lista apenas origins legítimas (sem `*`)
- [ ] Rate limiting em `POST /auth/login` (middleware chi ou nginx upstream)
- [ ] `admin@loja.local` / `admin123` trocados em produção
- [ ] Revisar `docs/reference/checklist.md` item a item

---

## §12 — Fase 9: Deploy (Railway + Supabase)

**Supabase:**
1. Criar projeto; copiar `DATABASE_URL` — host direto, porta `5432`, `sslmode=require`.
2. **Nunca usar porta `6543`** (transaction pooler — incompatível com migrations
   e criação de schema).

**Railway:**
1. Criar dois serviços no mesmo projeto Railway:
   - `erp-estoque-backend` via `backend/Dockerfile`
   - `erp-estoque-frontend` via `frontend/Dockerfile`
2. Variáveis do backend:
   ```
   DATABASE_URL=<supabase-direct-url>
   JWT_SECRET=<segredo-longo>
   JWT_ACCESS_TTL=15m
   JWT_REFRESH_TTL=720h
   ALLOWED_ORIGINS=https://<dominio-frontend>
   CEP_API_URL=https://viacep.com.br
   CUPOM_FISCAL_API_URL=<url>
   NOTA_FISCAL_API_URL=<url>
   ```
3. Pre-deploy command: `/app/migrate up`
4. Variável do frontend (definir **antes** do primeiro deploy):
   ```
   VITE_API_BASE_URL=https://<dominio-backend>
   ```

**Critério de aceitação F9:**

```bash
curl -s https://<backend-url>/health | grep -q "ok" && echo "PASS F9.health"

curl -s -X POST https://<backend-url>/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@loja.local","senha":"<nova-senha>"}' \
  | grep -q "access_token" && echo "PASS F9.login"
```

---

## §13 — Critério Final de Aceitação do Cliente

Execute o ciclo completo uma vez, sem planilha:

```
1.  Login como ESTOQUISTA
2.  Cadastrar fornecedor (CNPJ + CEP automático)
3.  Registrar compra com 2 produtos → Confirmar → verificar saldo atualizado
4.  Login como VENDEDOR
5.  Cadastrar cliente (CPF + CEP automático)
6.  Abrir nova venda → adicionar itens → aplicar desconto → Confirmar
    → verificar documento fiscal emitido e saldo baixado
7.  Login como ESTOQUISTA
8.  Lançar ajuste de estoque com motivo → verificar imutabilidade (sem botão editar)
9.  Consultar relatório "Abaixo do mínimo"
10. Login como ADMIN → consultar relatório de vendas do período

PASS FINAL: nenhum saldo negativo, nenhum erro fiscal, histórico completo.
```
