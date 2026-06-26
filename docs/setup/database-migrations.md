# 08 — Banco de Dados e Migrations

## Tecnologia

- **PostgreSQL 16**, acesso via `pgx/v5`.
- Migrations versionadas com **golang-migrate** (arquivos `.sql` em pares
  `*.up.sql` / `*.down.sql`).

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

Cada arquivo possui o par `.down.sql` para rollback.

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
make migrate-up             # aplica 000001..000009
make run                    # sobe a API
```

Comandos úteis (Makefile):

```bash
make migrate-up             # aplica todas
make migrate-down           # reverte a última
make migrate-create name=add_campo_x   # gera novo par .sql
make reset                  # drop total + up (recria do zero)
```

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
