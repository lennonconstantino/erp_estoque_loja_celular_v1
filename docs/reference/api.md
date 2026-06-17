# 06 — API REST

- Prefixo: `/api/v1`
- Formato: JSON. Datas em ISO-8601 (UTC).
- Autenticação: `Authorization: Bearer <access_token>` (exceto `/auth/login` e
  `/auth/refresh`).
- Autorização: cada rota exige uma permissão `recurso:acao` (ver [Segurança](security.md)).
- Erros: `{ "error": { "code": "...", "message": "..." } }` com status HTTP adequado.

## Auth (iam)

| Método | Rota | Permissão | Descrição |
|--------|------|-----------|-----------|
| POST | `/auth/login` | público | Login (email+senha) → access+refresh token |
| POST | `/auth/refresh` | público | Renova access token via refresh token |
| POST | `/auth/logout` | autenticado | Revoga refresh token |
| GET | `/auth/me` | autenticado | Dados e permissões do usuário logado |

## Usuários (iam)

| Método | Rota | Permissão |
|--------|------|-----------|
| GET | `/usuarios` | `iam:admin` |
| POST | `/usuarios` | `iam:admin` |
| GET | `/usuarios/{id}` | `iam:admin` |
| PUT | `/usuarios/{id}` | `iam:admin` |
| DELETE | `/usuarios/{id}` | `iam:admin` |
| POST | `/usuarios/{id}/papeis` | `iam:admin` |

## Clientes

| Método | Rota | Permissão | Observação |
|--------|------|-----------|------------|
| GET | `/clientes` | `clientes:read` | lista/pesquisa (`?q=`) |
| GET | `/clientes/by-cpf/{cpf}` | `clientes:read` | consulta antes de cadastrar |
| POST | `/clientes` | `clientes:write` | valida CPF |
| GET | `/clientes/{id}` | `clientes:read` | |
| PUT | `/clientes/{id}` | `clientes:write` | |
| DELETE | `/clientes/{id}` | `clientes:write` | |
| GET | `/clientes/cep/{cep}` | `clientes:read` | proxy do CepGateway |

## Fornecedores

| Método | Rota | Permissão |
|--------|------|-----------|
| GET | `/fornecedores` | `fornecedores:read` |
| GET | `/fornecedores/by-cnpj/{cnpj}` | `fornecedores:read` |
| POST | `/fornecedores` | `fornecedores:write` |
| GET/PUT/DELETE | `/fornecedores/{id}` | `fornecedores:read` / `:write` |

## Catálogo

| Método | Rota | Permissão |
|--------|------|-----------|
| GET/POST | `/categorias` | `catalogo:read` / `:write` |
| GET/PUT/DELETE | `/categorias/{id}` | `catalogo:read` / `:write` |
| GET | `/produtos` | `catalogo:read` (`?abaixo_minimo=true`) |
| POST | `/produtos` | `catalogo:write` |
| GET/PUT/DELETE | `/produtos/{id}` | `catalogo:read` / `:write` |
| GET | `/produtos/{id}/margem` | `catalogo:read` (preview de margem) |

## Compras

| Método | Rota | Permissão | Observação |
|--------|------|-----------|------------|
| GET | `/compras` | `compras:read` | |
| POST | `/compras` | `compras:write` | cria rascunho (master+itens) |
| GET | `/compras/{id}` | `compras:read` | |
| POST | `/compras/{id}/confirmar` | `compras:write` | **entra no estoque** |
| POST | `/compras/{id}/cancelar` | `compras:write` | |

## Vendas

| Método | Rota | Permissão | Observação |
|--------|------|-----------|------------|
| GET | `/vendas` | `vendas:read` | |
| POST | `/vendas` | `vendas:write` | cria rascunho; valida saldo dos itens |
| GET | `/vendas/{id}` | `vendas:read` | |
| POST | `/vendas/{id}/confirmar` | `vendas:write` | **baixa estoque** + emite doc. fiscal |
| POST | `/vendas/{id}/cancelar` | `vendas:write` | |

## Estoque

| Método | Rota | Permissão | Observação |
|--------|------|-----------|------------|
| GET | `/estoque/movimentacoes` | `estoque:read` | razão (`?produto=&de=&ate=`) |
| GET | `/estoque/ajustes` | `estoque:read` | |
| POST | `/estoque/ajustes` | `estoque:write` | **somente inclusão** |
| GET | `/estoque/saldo/{produto}` | `estoque:read` | |

## Relatórios

| Método | Rota | Permissão |
|--------|------|-----------|
| GET | `/relatorios/produtos` | `relatorios:read` |
| GET | `/relatorios/produtos/abaixo-minimo` | `relatorios:read` |
| GET | `/relatorios/produtos/mais-vendidos` | `relatorios:read` |
| GET | `/relatorios/produtos/menos-vendidos` | `relatorios:read` |
| GET | `/relatorios/{clientes\|fornecedores\|vendas\|compras}` | `relatorios:read` |

## Exemplo — confirmar venda

```http
POST /api/v1/vendas/9b1.../confirmar
Authorization: Bearer eyJ...
```
Fluxo do caso de uso: valida saldo de cada item → grava `val_venda` →
emite movimentações `VENDA` (baixa saldo) → chama `FiscalGateway`
(Cupom/NF) → publica `VendaConfirmada`. Resposta `200` com a venda confirmada
e o link/dados do documento fiscal; `409` se algum item ficou sem saldo.
