# 08 — Banco de Dados e Migrations

## Tecnologia

- **PostgreSQL 16**, acesso via `pgx/v5`.
- Migrations versionadas com **golang-migrate** (arquivos `.sql` em pares
  `*.up.sql` / `*.down.sql`).
- Dois caminhos de execução, **os mesmos arquivos**: o `migrate/migrate` CLI
  (alvos `make migrate-*` e serviço `migrate` do docker-compose) em dev, e um
  **runner Go embarcado** ([`backend/cmd/migrate`](../../backend/cmd/migrate),
  golang-migrate como biblioteca) para produção — é o binário `/app/migrate` da
  imagem do backend, rodado no pre-deploy do Railway.

## Arquivos de migration

Ordem de aplicação (em [`backend/migrations/`](../backend/migrations)):

| Versão | Arquivo | Conteúdo |
|--------|---------|----------|
| 000001 | `init` | extensões (`uuid-ossp`, `citext`), schemas, função `set_updated_at` |
| 000002 | `iam` | usuários, papéis, permissões, refresh tokens |
| 000003 | `clientes` | tabela de clientes (CPF único) |
| 000004 | `fornecedores` | tabela de fornecedores (CNPJ único) |
| 000005 | `catalogo` | categorias + produtos (CHECK custo<venda) |
| 000006 | `compras` | compra_master + detalhe_compras |
| 000007 | `vendas` | venda_master + detalhe_vendas (XOR cliente/consumidor) |
| 000008 | `estoque` | movimentações (append-only) + ajustes (imutáveis) |
| 000009 | `seed` | categorias exemplo, permissões, papel ADMIN, usuário admin |
| 000010 | `seed_demo` | dados de demonstração **idempotentes** (UUIDs fixos + `ON CONFLICT DO NOTHING`): fornecedores, clientes, produtos com saldo, compras e vendas CONFIRMADA — povoam as telas do frontend no primeiro acesso |

Cada arquivo possui o par `.down.sql` para rollback. O `seed_demo` é seguro
para reaplicar (não duplica) e seu `.down.sql` remove apenas os registros de
demonstração (filtrando pelos prefixos de UUID fixos). Como o ledger
`estoque.movimentacoes` é imutável por trigger, o `.down.sql` desabilita o
trigger pelo owner (`ALTER TABLE … DISABLE/ENABLE TRIGGER`) apenas durante o
`DELETE` dos registros de demonstração.

## Como inicializar

### Via Docker (recomendado)

```bash
cp backend/.env.example backend/.env
# edite backend/.env: preencha DB_PASSWORD e DATABASE_URL com URL literal
make up                     # db -> migrate (aplica tudo) -> api -> frontend
```

O `docker-compose.yml` fica na raiz do projeto. O serviço `migrate` roda
`migrate ... up` automaticamente após o Postgres ficar saudável; a `api` só
sobe depois das migrations.

> **Não use `docker compose up -d` diretamente** — o Makefile precisa passar
> `--env-file backend/.env` para que as variáveis do arquivo sejam lidas
> corretamente pelo Docker Compose.

### Local (golang-migrate CLI)

```bash
cd backend
cp .env.example .env
make migrate-up             # aplica 000001..000010
make run                    # sobe a API
```

Comandos úteis (Makefile):

```bash
make migrate-up             # aplica todas
make migrate-down           # reverte a última
make migrate-create name=add_campo_x   # gera novo par .sql
make reset                  # drop total + up (recria do zero)
```

### Banco remoto (Supabase) com o runner embarcado

Para provisionar um Postgres gerenciado do zero (schemas + seed + dados de
demonstração), use o runner Go via script — ele resolve a URL de
`backend/.env.production`, confirma o alvo e verifica o resultado:

```bash
make supabase-setup                 # usa backend/.env.production
make supabase-setup ARGS="-y"       # sem confirmação
scripts/supabase-setup.sh "<DATABASE_URL>"   # mira uma URL explícita
```

Precedência da URL: **argumento explícito → `backend/.env.production` → `$DATABASE_URL`**
(o `.env.production` vem antes do ambiente de propósito, já que o `make` exporta o
`DATABASE_URL` local). O comando equivalente embarcado é `/app/migrate up`
(também aceita `down`, `version`, `force`). Detalhes do deploy em
[Deploy no Railway](railway-deployment.md) e [Configuração do Supabase](supabase-setup.md).

## Estratégia de isolamento no banco

- **1 schema por domínio.** Facilita permissões por schema e, no futuro, separar
  cada um em seu próprio banco/instância.
- **Sem FK entre schemas.** Referências cross-context (`cli_venda`,
  `pro_dt_compra`, `for_compra`, etc.) são UUID sem constraint física. A
  integridade é responsabilidade da aplicação.
- **Migrations por domínio.** Quando um módulo virar serviço, leva consigo
  apenas suas migrations (`000003_clientes.*`, etc.) — elas já são autocontidas.

## Integridade e regras no banco (defesa em profundidade)

Além das validações no domínio Go, o banco reforça invariantes críticas:

- `catalogo.produtos`: `CHECK (p_custo_pro < p_venda_pro)`, `estoque_a_pro >= 0`.
- `vendas.venda_master`: `CHECK` cliente **xor** consumidor final.
- `compras.detalhe_compras` / `vendas.detalhe_vendas`: quantidade > 0.
- `estoque.ajustes` e `estoque.movimentacoes`: trigger bloqueia `UPDATE`/`DELETE`
  (imutabilidade / "somente consulta e inclusão").

## Consistência de estoque

A baixa/entrada de saldo ocorre **dentro de uma transação** junto com a gravação
do documento (venda/compra/ajuste) e da movimentação no razão. Para vendas
concorrentes, usar trava pessimista:

```sql
-- baixa segura: falha (0 linhas) se não houver saldo suficiente
UPDATE catalogo.produtos
   SET estoque_a_pro = estoque_a_pro - $qtd,
       disp_pro      = (estoque_a_pro - $qtd) > 0
 WHERE id_pro = $id AND estoque_a_pro >= $qtd;
```
