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

O backend Go precisa apenas da connection string direta. O frontend não conecta no Supabase — ele fala apenas com a API Go.

| Valor | Onde encontrar | Usado por |
|-------|---------------|-----------|
| **Connection string (Session mode)** | Dashboard → **Project Settings** → **Database** → *Connection string* → aba *URI* | Backend + migrations |
| **Database password** | A senha definida na criação do projeto | Parte da connection string |
| **Project ref** | URL do dashboard: `supabase.com/dashboard/project/<ref>` | Referência nos comandos CLI |

> **Atenção:** troque `sslmode=disable` por `sslmode=require` na connection string do Supabase. O PostgreSQL local usa `disable`; o Supabase exige SSL.

Exemplo de `DATABASE_URL` para Supabase:

```
postgres://postgres:[senha]@db.[ref].supabase.co:5432/postgres?sslmode=require
```

Nunca use o **Transaction pooler** (porta 6543) para migrations — use a conexão direta (porta 5432).

## 4. Configurar variáveis de ambiente

Copie `backend/.env.example` para `backend/.env` e preencha:

```dotenv
# Banco de dados — Supabase (produção / staging)
DATABASE_URL=postgres://postgres:[senha]@db.[ref].supabase.co:5432/postgres?sslmode=require

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
