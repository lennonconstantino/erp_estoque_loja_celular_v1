# Configuração do Supabase

O Supabase é usado exclusivamente como **host do PostgreSQL**. Autenticação é JWT gerenciado pelo próprio backend Go — o Supabase Auth **não é utilizado**. Você precisa de um projeto Supabase antes de rodar o backend ou as migrations.

## 1. Criar conta

1. Acesse [supabase.com](https://supabase.com) e crie uma conta (GitHub ou e-mail).
2. Confirme o e-mail se solicitado.
3. Você cai no [dashboard](https://supabase.com/dashboard). O plano gratuito é suficiente para desenvolvimento.

## 2. Criar projeto

1. Clique em [New project](https://supabase.com/dashboard/new).
2. Escolha sua organização (criada automaticamente no primeiro acesso).
3. Defina um **nome de projeto** (ex.: `erp-estoque`).
4. Crie uma **senha de banco de dados** — guarde em local seguro; ela compõe a connection string.
5. Escolha uma **região** próxima (ex.: South America — São Paulo).
6. Clique em **Create new project** e aguarde o status ficar *healthy* (~1–2 min).

## 3. Coletar credenciais

O backend Go precisa apenas da connection string. O frontend não conecta no Supabase — ele fala apenas com a API Go.

| Valor | Onde encontrar | Usado por |
|-------|---------------|-----------|
| **Connection string (Session pooler)** | Dashboard → **Connect** → aba *Connection string* → **Session pooler** | Backend + migrations |
| **Database password** | A senha definida na criação do projeto | Parte da connection string |
| **Project ref** | URL do dashboard: `supabase.com/dashboard/project/<ref>` | Referência nos comandos CLI |

> **Atenção:** garanta `sslmode=require` na connection string do Supabase. O PostgreSQL local usa `disable`; o Supabase exige SSL.

### Qual connection string usar — e por quê (lição do deploy no Railway)

O Supabase oferece três formas de conexão. A escolha certa depende de **onde o
backend roda**:

| Forma | Host / porta | IP | Serve p/ migrations? | Quando usar |
|-------|--------------|----|----|-------------|
| **Direct** | `db.<ref>.supabase.co:5432` | **IPv6-only** | Sim | Só de máquinas com IPv6 (ex.: seu laptop) |
| **Session pooler** | `aws-1-<região>.pooler.supabase.com:5432` (user `postgres.<ref>`) | **IPv4** | Sim | **Produção no Railway** (egress IPv4) e qualquer host sem IPv6 |
| **Transaction pooler** | `...pooler.supabase.com:6543` | IPv4 | **Não** (quebra prepared statements do pgx/migrate) | Serverless de altíssima concorrência — não neste projeto |

> **Railway (e a maioria dos PaaS) só tem egress IPv4.** A conexão **direta** do
> Supabase publica apenas registro AAAA (IPv6), então o pre-deploy `migrate up`
> falha com `dial tcp [2600:…]:5432: network is unreachable`. Em produção use a
> **Session pooler** (IPv4, porta 5432 — suporta migrations). Detalhes e o runbook
> em [../licoes-aprendidas.md](../licoes-aprendidas.md).

Exemplos de `DATABASE_URL`:

```
# Session pooler (produção no Railway — IPv4, porta 5432)
postgres://postgres.<ref>:[senha]@aws-1-<região>.pooler.supabase.com:5432/postgres?sslmode=require

# Direct (dev, só de máquina com IPv6)
postgres://postgres:[senha]@db.<ref>.supabase.co:5432/postgres?sslmode=require
```

## 4. Configurar variáveis de ambiente

Copie `backend/.env.example` para `backend/.env` e preencha:

```dotenv
# Banco de dados — Supabase (produção / staging)
# Em produção (Railway), use a Session pooler (IPv4). Ver §3 acima.
DATABASE_URL=postgres://postgres.[ref]:[senha]@aws-1-[região].pooler.supabase.com:5432/postgres?sslmode=require

# JWT — troque em produção
JWT_SECRET=troque-este-segredo-em-producao
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h

# CEP e fiscal (opcional por agora)
CEP_API_URL=https://viacep.com.br/ws
```

O frontend usa apenas `VITE_API_BASE_URL` (URL do backend Go) — nenhuma credencial do Supabase vai para o cliente.

## 5. Executar as migrations

As migrations são gerenciadas pelo [golang-migrate](https://github.com/golang-migrate/migrate) e ficam em `backend/migrations/`. Para o detalhamento completo de cada arquivo e dos comandos disponíveis, veja [Banco de Dados e Migrations](database-migrations.md).

Resumo dos comandos a partir de `backend/`:

```bash
make migrate-up                        # aplica todas as migrations
make migrate-down                      # reverte a última
make reset                             # drop total + up (recria do zero)
```

### Criar e popular o Supabase com um comando

Para provisionar o banco do Supabase do zero (criar schemas/tabelas **e** popular
admin + dados de demonstração), use o script guardado — ele resolve a URL de
`backend/.env.production`, pede confirmação e verifica o resultado:

```bash
make supabase-setup                    # usa backend/.env.production
make supabase-setup ARGS="-y"          # sem confirmação interativa
# ou direto, com uma URL explícita:
scripts/supabase-setup.sh "postgres://postgres:SENHA@db.<ref>.supabase.co:5432/postgres?sslmode=require"
```

O script é idempotente (rodar de novo num banco já migrado é no-op) e usa o mesmo
runner `cmd/migrate` do pre-deploy do Railway.

> **Schemas:** `iam`, `clientes`, `fornecedores`, `catalogo`, `compras`, `vendas`, `estoque`. Não há foreign keys físicas entre schemas — referências externas usam UUID solto. Veja o modelo completo em [Modelo de Dados](../reference/data-model.md).

## 6. Autenticação — o que o Supabase Auth NÃO faz aqui

O Supabase Auth **não é configurado nem utilizado**. Toda a autenticação é gerenciada pelo backend Go:

- Senhas com hash **bcrypt** em `iam.usuarios.senha_hash_usr`.
- Login retorna um **access token JWT** (15 min) + **refresh token** opaco (30 dias).
- O frontend envia `Authorization: Bearer <access_token>` no header de cada requisição.
- Nenhuma chave `anon` ou `service_role` do Supabase é usada no frontend.

Para entender o fluxo completo de auth, veja [Segurança](../reference/security.md).

## Próximos passos

- [Backend setup](backend-setup.md) — variáveis de ambiente, build e execução do servidor Go
- [Frontend setup](frontend-setup.md) — SPA React e integração com a API
