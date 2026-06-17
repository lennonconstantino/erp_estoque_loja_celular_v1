# 05 — Modelo de Dados

PostgreSQL, **1 schema por bounded context**. Sem foreign keys entre schemas
(isolamento). IDs externos são UUID "soltos". DDL completo em
[`backend/migrations/`](../backend/migrations).

## Convenções

- PK: `id_<prefixo>` UUID (`uuid_generate_v4()`).
- Colunas sufixadas pelo prefixo do domínio (ex.: `_cli`, `_for`, `_pro`).
- Auditoria: `created_at` / `updated_at` (trigger `set_updated_at`).
- Valores monetários: `NUMERIC(12,2)` / totais `NUMERIC(14,2)`.

## Diagrama de relacionamentos (lógico)

```
iam.usuarios ─┬─< iam.usuario_papeis >─ iam.papeis ─< iam.papel_permissoes >─ iam.permissoes
              └─< iam.refresh_tokens

clientes.clientes            fornecedores.fornecedores
        ▲ (id solto em vendas)        ▲ (id solto em compras)
        │                             │
catalogo.categorias ─< catalogo.produtos
                              ▲ (id solto)
        ┌─────────────────────┼─────────────────────┐
compras.compra_master      vendas.venda_master    estoque.ajustes
   └─< compras.detalhe_compras  └─< vendas.detalhe_vendas
                              │
                  estoque.movimentacoes  (razão: COMPRA|VENDA|AJUSTE_*)
                       └─ atualiza catalogo.produtos.estoque_a_pro
```
Linhas tracejadas (id solto) = referência por UUID sem FK física.

## Tabelas

### iam
| Tabela | Colunas-chave |
|--------|---------------|
| `usuarios` | `id_usr`, `email_usr` (UNIQUE), `senha_hash_usr`, `atv_usr` |
| `papeis` | `id_pap`, `nome_pap` (UNIQUE) |
| `permissoes` | `id_per`, `codigo_per` (UNIQUE, ex. `vendas:write`) |
| `papel_permissoes` | (`pap_id`,`per_id`) PK |
| `usuario_papeis` | (`usr_id`,`pap_id`) PK |
| `refresh_tokens` | `id_rt`, `usr_id`, `token_hash`, `expira_em`, `revogado` |

### clientes.clientes
`id_cli`, **`cpf_cli` (UNIQUE, 11 díg.)**, `nome_cli`, `email_cli`, `tel_cli`,
`cep_cli`, `rua_cli`, `num_cli`, `comp_cli`, `bai_cli`, `cit_cli`, `uf_cli`,
`dt_ult_comp_cli`, `atv_cli`.

### fornecedores.fornecedores
`id_for`, **`cnpj_for` (UNIQUE, 14 díg.)**, `razao_for`, `nome_fant_for`,
`email_for`, `tel1_for`, `tel2_for`, endereço (`cep/rua/num/comp/bai/cit/uf`),
`comercial_for`, `financeiro_for`, `dt_ult_comp_for`, `atv_for`.

### catalogo
| Tabela | Colunas-chave |
|--------|---------------|
| `categorias` | `id_cat`, `desc_cat` (UNIQUE) |
| `produtos` | `id_pro`, `cat_pro`→categorias, `desc_pro`, `p_custo_pro`, `p_venda_pro`, `estoque_m_pro`, `estoque_a_pro`, `garant_pro`, `mod_pro`, `disp_pro` |

**Constraints de produto:** `p_custo_pro < p_venda_pro`; `estoque_a_pro >= 0`.

### compras
| Tabela | Colunas-chave |
|--------|---------------|
| `compra_master` | `id_compra`, `for_compra` (id solto), `nf_compra`, `dt_compra`, `val_compra`, `status_compra` |
| `detalhe_compras` | `id_dt_compra`, `compra_id`→master, `pro_dt_compra` (id solto), `qtd_dt_compra` (>0), `pre_compra_dt_compra`, `pre_venda_dt_compra`, `margem_dt_compra` |

### vendas
| Tabela | Colunas-chave |
|--------|---------------|
| `venda_master` | `id_venda`, `dt_venda`, `val_venda`, `dsc_venda`, `forma_pgto_venda`, `cli_venda` (id solto, NULL se consumidor final), `c_final_venda`, `doc_fiscal_venda`, `status_venda` |
| `detalhe_vendas` | `id_dt_venda`, `venda_id`→master, `pro_venda` (id solto), `qtd_venda` (>0), `pre_venda_dt` |

**Constraint:** `c_final_venda XOR cli_venda` (consumidor final não tem cliente).

### estoque
| Tabela | Colunas-chave |
|--------|---------------|
| `movimentacoes` | `id_mov`, `pro_mov`, `tipo_mov` (COMPRA/VENDA/AJUSTE_*), `qtd_mov`, `saldo_ant_mov`, `saldo_atu_mov`, `origem_tipo`, `origem_id`, `resp_mov`, `dt_mov` — **append-only** |
| `ajustes` | `id_ajs`, `pro_ajs`, `qtd_entrada_ajs`, `qtd_saida_ajs`, `mot_ajs`, `resp_ajs`, `dt_ajs` — **sem UPDATE/DELETE** |

## Mapeamento diagrama → tabela

| Diagrama (print) | Tabela |
|------------------|--------|
| Cadastro de Clientes | `clientes.clientes` |
| Cadastro de Fornecedores | `fornecedores.fornecedores` |
| Cadastro de Categorias | `catalogo.categorias` |
| Cadastro de Produtos | `catalogo.produtos` |
| Módulo de Compras (master/detalhe) | `compras.compra_master` / `compras.detalhe_compras` |
| Módulo Vendas (master/detalhe) | `vendas.venda_master` / `vendas.detalhe_vendas` |
| Ajustar Estoque | `estoque.ajustes` (+ razão `estoque.movimentacoes`) |
| Login (JWT + RBAC) / Usuários | `iam.*` |

> **Nota sobre os nomes do diagrama:** mantidos os prefixos abreviados
> (`id_cli`, `cpf_cli`, ...). Ajustes feitos: `dt_uti_comp_clli` → `dt_ult_comp_cli`;
> incluído `desc_pro` (descrição do produto, faltante no print) e `uf_*` no
> endereço (necessário ao retorno do CEP).
