# Deploy no Railway

O repositório sobe no Railway como dois serviços dentro de um único projeto:

- `erp-estoque-backend` — servidor Go (chi), construído via `backend/Dockerfile`.
- `erp-estoque-frontend` — build estático Vite servido por Nginx, construído via `frontend/Dockerfile` (a criar).

O PostgreSQL fica hospedado no Supabase. Não adicione um Railway Postgres para este projeto.

## Antes de começar

Tenha em mãos:

- O repositório no GitHub (ou a Railway CLI com o projeto local).
- Um projeto Supabase configurado — veja [Supabase setup](supabase-setup.md).
- A connection string direta do Supabase (`sslmode=require`, porta 5432).

## Dockerfile do frontend (ainda não existe)

Antes de fazer o deploy do frontend, crie `frontend/Dockerfile`:

```dockerfile
# ---- build ----
FROM node:20-alpine AS build
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@latest --activate
COPY package.json pnpm-lock.yaml .npmrc ./
RUN pnpm install --frozen-lockfile
COPY . .
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
RUN pnpm build

# ---- runtime ----
FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

E crie `frontend/nginx.conf` para o SPA (redirecionar 404 para `index.html`):

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location /health {
        return 200 "ok";
        add_header Content-Type text/plain;
    }

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

## Migrations no Railway

O backend precisa das migrations aplicadas antes de subir. O `Dockerfile` do backend já copia a pasta `migrations/`. Para expor o runner de migrations como binário separado, ajuste o `backend/Dockerfile` para também compilar `cmd/migrate`:

```dockerfile
# adicionar no stage build, após compilar a api:
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate
# e no stage runtime:
COPY --from=build /out/migrate /app/migrate
```

Com isso, o comando de pre-deploy no Railway será:

```bash
/app/migrate up
```

Se preferir não alterar o Dockerfile agora, rode as migrations manualmente via CLI antes do primeiro deploy (veja opção B, passo 6).

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
DATABASE_URL=postgres://postgres:[senha]@db.[ref].supabase.co:5432/postgres?sslmode=require
JWT_SECRET=segredo-longo-e-aleatorio-para-producao
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h
ALLOWED_ORIGINS=http://localhost:5173
CEP_API_URL=https://viacep.com.br/ws
```

> `APP_PORT` não é necessário — o Railway injeta `PORT` automaticamente. O servidor Go deve ler `PORT` (ou fallback para `8080`).

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
  DATABASE_URL=postgres://postgres:[senha]@db.[ref].supabase.co:5432/postgres?sslmode=require \
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

- Não defina `PORT` manualmente. O Railway injeta automaticamente — o servidor Go deve usá-lo.
- As variáveis `VITE_*` são compiladas dentro do build. Defina-as **antes** de rodar `railway up` para o frontend; redeploy após qualquer mudança.
- Use a connection string de **sessão direta** (porta 5432) tanto para a API quanto para migrations — nunca o transaction pooler (porta 6543).
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
