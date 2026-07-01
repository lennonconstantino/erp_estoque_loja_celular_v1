# Deploy no Railway

O repositório sobe no Railway como dois serviços dentro de um único projeto:

- `erp-estoque-backend` — servidor Go (chi), construído via `backend/Dockerfile`.
- `erp-estoque-frontend` — build estático Vite servido por Nginx, construído via `frontend/Dockerfile`.

O PostgreSQL fica hospedado no Supabase. Não adicione um Railway Postgres para este projeto.

## Antes de começar

Tenha em mãos:

- O repositório no GitHub (ou a Railway CLI com o projeto local).
- Um projeto Supabase configurado — veja [Supabase setup](supabase-setup.md).
- A connection string direta do Supabase (`sslmode=require`, porta 5432).

## O que já vem pronto no repositório

A Fase 9 deixou a infraestrutura de deploy commitada — não há nada de Dockerfile/config a
criar na mão:

- **`backend/Dockerfile`** compila dois binários no build stage e os copia para o runtime:
  `/app/api` (servidor) e `/app/migrate` (runner de migrations). A pasta `migrations/` também
  é copiada.
- **`frontend/Dockerfile` + `frontend/nginx.conf`** fazem o build estático do Vite e servem por
  nginx com SPA fallback para `/index.html` e um endpoint `GET /health` (texto `ok`) para o
  healthcheck.
- **`backend/railway.json`** e **`frontend/railway.json`** declaram builder Dockerfile,
  `healthcheckPath: /health` e — no backend — o `preDeployCommand: /app/migrate up`. Com os
  Root Directories corretos (`backend`/`frontend`), o Railway lê esses arquivos e aplica
  pre-deploy/healthcheck automaticamente, sem configuração manual no painel.
- **Templates de variáveis**: `backend/.env.production.example` e `frontend/.env.production.example`
  listam exatamente o que preencher nas Variables de cada serviço.

> O backend lê a porta de `PORT` (injetada pelo Railway), com fallback para `APP_PORT` e
> depois `8080`. Não defina `PORT`/`APP_PORT` manualmente no Railway.

## Migrations no Railway

O binário `/app/migrate` (golang-migrate embarcado, mesmos arquivos `migrations/*.sql` do
docker-compose) já está na imagem do backend e é executado pelo `preDeployCommand` declarado em
`backend/railway.json`:

```bash
/app/migrate up
```

Comandos disponíveis: `up`, `down [n]`, `version`, `force <v>`. Em produção use sempre `up`.

## Opção A: Railway UI

Use este caminho se preferir configurar pelo painel.

1. No Railway, clique em **New Project** → **Deploy from GitHub repo** e selecione este repositório.
2. Nomeie o serviço de backend como `erp-estoque-backend`.
3. Abra **Settings** do backend:
   - **Root Directory:** `backend`
   - **Healthcheck Path:** `/health`
   - Deixe build e start commands em branco — Railway usa o `Dockerfile`.
4. Adicione as **variáveis** do backend:

```text
APP_ENV=production
DATABASE_URL=postgres://postgres.[ref]:[senha]@aws-1-[região].pooler.supabase.com:5432/postgres?sslmode=require
JWT_SECRET=segredo-longo-e-aleatorio-para-producao
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h
ALLOWED_ORIGINS=http://localhost:5173
CEP_API_URL=https://viacep.com.br/ws
```

> **`DATABASE_URL` = Session pooler (IPv4).** Use a connection string da **Session pooler**
> do Supabase (`aws-1-<região>.pooler.supabase.com:5432`, user `postgres.<ref>`), **não** a
> conexão direta `db.<ref>.supabase.co` — esta é IPv6-only e o Railway é IPv4, o que faz o
> `migrate` falhar com `network is unreachable`. Ver [supabase-setup.md](supabase-setup.md#qual-connection-string-usar--e-por-quê-lição-do-deploy-no-railway).

> `APP_PORT` não é necessário — o Railway injeta `PORT` automaticamente. O servidor Go deve ler `PORT` (ou fallback para `8080`).

> **Fail-closed:** com `APP_ENV=production` o backend **aborta no boot** se `JWT_SECRET`
> ou `DATABASE_URL` estiverem vazios ou no default inseguro de dev. Se o serviço entrar
> em crash-loop logo ao subir, confira o log `configuração inválida: ...` — quase sempre
> é uma dessas variáveis faltando. Gere o `JWT_SECRET` com `openssl rand -base64 64 | tr -d '\n'`.

5. Em **Settings** → **Deploy**, defina o **pre-deploy command**:

```bash
/app/migrate up
```

6. Faça o deploy do backend e gere um domínio público em **Networking**.
7. Confirme: `https://seu-backend.up.railway.app/health` deve retornar `{"status":"ok"}`.
8. No mesmo projeto, clique em **New** → **GitHub Repo** e selecione o repositório novamente.
9. Nomeie o serviço de frontend como `erp-estoque-frontend`.
10. Abra **Settings** do frontend:
    - **Root Directory:** `frontend`
    - **Healthcheck Path:** `/health`
    - Build command: deixar em branco (Dockerfile cuida de tudo).
11. Adicione as **variáveis** do frontend:

```text
VITE_API_BASE_URL=https://seu-backend.up.railway.app
```

12. Faça o deploy do frontend e gere um domínio público.
13. Atualize a variável do backend e faça redeploy:

```text
ALLOWED_ORIGINS=https://seu-frontend.up.railway.app
```

## Opção B: CLI + MCP

Use este caminho para deixar um agente (ex.: Cursor) conduzir o deploy.

1. Instale e autentique a CLI:

```bash
railway login
railway setup agent -y
```

2. Crie um projeto privado via MCP:

```text
Criar um projeto Railway privado chamado erp-estoque.
```

3. Vincule o repositório local ao projeto:

```bash
railway link --project <project-id> --environment production
```

4. Crie os dois serviços vazios:

```bash
railway add --service erp-estoque-backend --json
railway add --service erp-estoque-frontend --json
```

5. Configure as variáveis do backend (use `--stdin` para segredos):

```bash
railway variable set \
  APP_ENV=production \
  DATABASE_URL=postgres://postgres.[ref]:[senha]@aws-1-[região].pooler.supabase.com:5432/postgres?sslmode=require \
  JWT_ACCESS_TTL=15m \
  JWT_REFRESH_TTL=720h \
  CEP_API_URL=https://viacep.com.br/ws \
  ALLOWED_ORIGINS=http://localhost:5173 \
  --service erp-estoque-backend \
  --skip-deploys

printf "%s" "$JWT_SECRET" | railway variable set JWT_SECRET \
  --stdin \
  --service erp-estoque-backend \
  --skip-deploys
```

6. Faça o deploy do backend e gere o domínio:

```bash
railway up ./backend --path-as-root --service erp-estoque-backend --detach
railway domain --service erp-estoque-backend --json
```

7. Rode as migrations contra o banco de produção:

```bash
railway run --service erp-estoque-backend -- /app/migrate up
```

8. Configure as variáveis do frontend e faça o deploy:

```bash
railway variable set \
  VITE_API_BASE_URL=https://seu-backend.up.railway.app \
  --service erp-estoque-frontend \
  --skip-deploys

railway up ./frontend --path-as-root --service erp-estoque-frontend --detach
railway domain --service erp-estoque-frontend --json
```

9. Atualize o CORS do backend e faça redeploy:

```bash
railway variable set ALLOWED_ORIGINS=https://seu-frontend.up.railway.app \
  --service erp-estoque-backend

railway redeploy --service erp-estoque-backend --yes
```

10. Use `get-status` / `get-logs` via MCP para verificar se ambos os serviços estão `SUCCESS`.

## Lições práticas

- **`DATABASE_URL` = Session Pooler do Supabase (IPv4), não a conexão direta.** A
  direta `db.<ref>.supabase.co:5432` só tem registro **AAAA (IPv6)** e o egress do
  Railway é IPv4 → o pre-deploy `migrate up` falha com `dial tcp [..]:5432: network
  is unreachable`. Use o **Session pooler** (Supabase → Connect → *Session pooler*):
  `postgres://postgres.<ref>:<senha>@aws-1-<região>.pooler.supabase.com:5432/postgres?sslmode=require`
  (usuário vira `postgres.<ref>`, host `...pooler.supabase.com`). É IPv4, porta 5432
  e suporta migrations. **Não** use o *Transaction pooler* (porta 6543): quebra os
  prepared statements do pgx e o golang-migrate.
- **nginx do frontend deve escutar em `$PORT`.** O Railway injeta `PORT` e faz o
  healthcheck nessa porta; o `nginx.conf` declara `listen 80;`, então o
  `frontend/Dockerfile` reescreve para `$PORT` no boot (`sed … listen ${PORT:-80}`).
  Sem isso o healthcheck nunca alcança o nginx e o deploy fica `Failed` (o container
  sobe e o Railway o mata em seguida).
- Não defina `PORT` manualmente. O Railway injeta automaticamente — o servidor Go deve usá-lo.
- As variáveis `VITE_*` são compiladas dentro do build. Defina-as **antes** de rodar `railway up` para o frontend; redeploy após qualquer mudança.
- Deploy do frontend via CLI: como o serviço tem **Root Directory = `frontend`**, rode
  `railway up --service erp-estoque-frontend` a partir da **raiz do repo** (não
  `railway up ./frontend --path-as-root`, que faz o builder procurar `frontend/frontend`).
- O `JWT_SECRET` em produção deve ser uma string longa e aleatória. Nunca use o valor padrão do `.env.example`.

## Verificação final

1. Abra a URL do frontend no Railway.
2. Faça login com e-mail e senha (usuário `admin` criado pelo seed).
3. Teste um fluxo básico: cadastrar produto, registrar uma venda.
4. Verifique os endpoints de health:

```text
https://seu-backend.up.railway.app/health
https://seu-frontend.up.railway.app/health
```

O backend deve retornar `{"status":"ok"}` e o frontend `ok`.
