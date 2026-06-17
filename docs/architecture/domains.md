# 04 — Domínios (Bounded Contexts)

Cada domínio = 1 pacote Go isolado + 1 schema no banco. A tabela abaixo mapeia
módulo → schema → responsabilidade.

| Domínio | Schema | Responsabilidade | Depende de (porta) |
|---------|--------|------------------|--------------------|
| `iam` | `iam` | Usuários, autenticação JWT, papéis e permissões | — |
| `clientes` | `clientes` | Cadastro de clientes, validação de CPF, CEP | `CepGateway` |
| `fornecedores` | `fornecedores` | Cadastro de fornecedores, validação de CNPJ, CEP | `CepGateway` |
| `catalogo` | `catalogo` | Categorias e produtos, margem, disponibilidade | `EstoqueReader` |
| `compras` | `compras` | Compras e itens; entrada de estoque | `EstoqueWriter`, `CatalogoReader` |
| `vendas` | `vendas` | Vendas e itens; saída de estoque; doc. fiscal | `EstoqueWriter`, `CatalogoReader`, `FiscalGateway` |
| `estoque` | `estoque` | Razão de movimentações + ajustes + saldo | `CatalogoWriter` (atualiza saldo) |

> `CepGateway`, `FiscalGateway` são adaptadores de saída para APIs externas
> (consulta de CEP; emissão de cupom/NF).

---

## iam
Núcleo de identidade e acesso. Emite access/refresh tokens, valida senha
(bcrypt), resolve papéis→permissões. Expõe middleware de autorização para os
demais módulos via `platform/auth`. **Regra:** toda rota protegida exige uma
permissão (`recurso:acao`).

## clientes
- **Invariantes:** CPF com 11 dígitos válidos e único; Nome e E-mail obrigatórios.
- **Fluxo de cadastro:** informar CPF → consultar existência → atualizar ou criar.
- **CEP:** ao informar CEP, `CepGateway` preenche rua/bairro/cidade/UF.
- Mantém `dt_ult_comp_cli` (atualizada por evento `VendaConfirmada`).

## fornecedores
- **Invariantes:** CNPJ com 14 dígitos válidos e único; obrigatórios CNPJ,
  Razão Social, Nome Fantasia, E-mail, Telefone 1, Contato Comercial.
- Mantém `dt_ult_comp_for` (atualizada por evento `CompraConfirmada`).

## catalogo
- **Categorias:** descrição única (CRUD).
- **Produtos — invariantes:**
  - `p_custo_pro < p_venda_pro` (garantido no domínio e por CHECK no banco);
  - margem % = `(venda - custo) / custo * 100` (derivada, exibida no cadastro);
  - `estoque_a_pro` **não** é editável pelo cadastro;
  - `disp_pro = (estoque_a_pro > 0)`.
- Expõe `CatalogoReader` (consultar produto/preço/saldo) e `CatalogoWriter`
  (atualizar saldo materializado) — usados por compras/vendas/estoque.

## compras
- Cabeçalho (`compra_master`) + itens (`detalhe_compras`).
- **Invariantes:** quantidade > 0; custo < venda no item.
- **Confirmar compra (caso de uso):** transação que (1) marca `CONFIRMADA`,
  (2) para cada item emite movimentação `COMPRA` no `estoque`,
  (3) opcionalmente atualiza preço de venda do produto,
  (4) publica `CompraConfirmada` (atualiza `dt_ult_comp_for`).

## vendas
- Cabeçalho (`venda_master`) + itens (`detalhe_vendas`).
- **Invariantes:** quantidade > 0; desconto ≥ 0; cliente **xor** consumidor final.
- **Pré-validação de item:** via `CatalogoReader`, saldo > 0 e ≥ quantidade.
- **Confirmar venda (caso de uso):** transação que (1) calcula `val_venda`,
  (2) para cada item emite movimentação `VENDA` no `estoque`,
  (3) chama `FiscalGateway` (Cupom ou NF),
  (4) publica `VendaConfirmada` (atualiza `dt_ult_comp_cli`).
- NF exige cliente cadastrado (validação do caso de uso).

## estoque
- **`movimentacoes`** (append-only): fonte da verdade. Toda entrada/saída passa
  por aqui com `saldo_ant`/`saldo_atu`.
- **`ajustes`:** documento da tela "Ajustar Estoque".
  - estornar entrada → movimentação `AJUSTE_SAIDA` (baixa o saldo);
  - estornar saída → movimentação `AJUSTE_ENTRADA` (acresce o saldo).
  - **Somente consulta e inclusão** (UPDATE/DELETE bloqueados por trigger).
- Após cada movimentação, atualiza `catalogo.produtos.estoque_a_pro` via
  `CatalogoWriter` e recalcula `disp_pro`.
- **Concorrência:** baixa/entrada de saldo usa `SELECT ... FOR UPDATE` (ou
  `UPDATE ... WHERE estoque_a_pro >= qtd`) para evitar saldo negativo em vendas
  simultâneas.
