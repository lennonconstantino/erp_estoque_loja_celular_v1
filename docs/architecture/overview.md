# 01 — Visão Geral

## Objetivo

Sistema ERP de estoque para uma loja de acessórios de celular. Permite cadastrar
clientes, fornecedores, categorias e produtos; registrar compras (entrada de
estoque) e vendas (saída de estoque); lançar ajustes de estoque; e emitir
relatórios. O acesso é controlado por autenticação JWT e papéis (RBAC).

## Módulos e regras de negócio

### Login / Auth
- Autenticação via **JWT** (access token + refresh token).
- Autorização por papéis e permissões (**RBAC**).

### Usuários
- Cadastro de usuários, vínculo com papéis. CRUD restrito ao papel `ADMIN`.

### Clientes
- Campos: CPF, Nome, E-mail, Telefone, CEP, Rua, Número, Complemento, Bairro, Cidade.
- **Regras:** CPF validado e único; ao informar CPF, consultar se já existe
  (se sim, abrir para atualização; se não, pedir os demais campos).
  Obrigatórios: **CPF, Nome, E-mail**. CEP preenche Rua/Bairro/Cidade/UF.
- Operações: **CRUD**.

### Fornecedores
- Campos: CNPJ, Razão Social, Nome Fantasia, E-mail, Telefone 1/2, CEP, Rua,
  Número, Complemento, Bairro, Cidade, Contato Comercial, Contato Financeiro.
- **Regras:** CNPJ validado e único; consulta antes de cadastrar.
  Obrigatórios: **CNPJ, Razão Social, Nome Fantasia, E-mail, Telefone 1,
  Contato Comercial**. CEP preenche endereço.
- Operações: **CRUD**.

### Categorias
- Campo: Descrição (única). Operações: **CRUD**.
- Exemplos: Capa, Película, Carregador, Cabo tipo C, Cabo USB Mini B, Cabo USB Micro B.

### Produtos
- Campos: Categoria, Descrição, Preço de Custo, Preço de Venda, Estoque Mínimo,
  Estoque Atual, Garantia, Modelo de Celular.
- **Regras:**
  - Preço de custo **sempre menor** que preço de venda.
  - Estoque igual a zero ⇒ produto **indisponível**.
  - O saldo **não** é ajustado pelo cadastro de produto (somente por
    Compras / Vendas / Ajustes).
  - Ao informar custo e venda, exibir o **% de margem**.
- Operações: **CRUD**.

### Compras (entrada de estoque)
- Cabeçalho: Fornecedor, Número da NF, Data da Compra.
- Itens: Produto, Quantidade, Preço de Custo, Margem %, Preço de Venda.
- **Regras:** preço de custo < preço de venda; quantidade > 0.
  Ao **confirmar a compra**, somar a quantidade ao saldo de cada produto.

### Vendas (saída de estoque)
- Cabeçalho: Data, Valor, Forma de Pagamento, Desconto, Cliente.
- Itens: Produto, Quantidade.
- **Regras:**
  - Valor da venda é calculado pelos itens; desconto é gravado.
  - Antes de aceitar um item: estoque > 0 e saldo ≥ quantidade pedida.
  - Venda para **cliente cadastrado** ou **consumidor final**.
  - Documento fiscal: **Cupom** (API de cupom fiscal) ou **NF** (API de nota fiscal).
  - Para NF, o cliente precisa estar cadastrado — oferecer link para cadastro na hora.
  - Ao **confirmar a venda**, baixar o saldo: `saldo = saldo - qtd_venda`.

### Ajuste de Estoque
- Campos: Produto, Estornar Entrada, Estornar Saída, Motivo, Data, Responsável.
- **Regras:**
  - Estornar entrada ⇒ **baixa** no saldo do produto.
  - Estornar saída ⇒ **acréscimo** no saldo do produto.
  - **Somente consulta e inclusão** — sem alteração e sem exclusão.

### Relatórios
- **Produtos:** Listagem, Abaixo do estoque mínimo, Mais vendidos, Menos vendidos.
- **Fornecedores / Clientes / Vendas / Compras:** Listagem.

## Atores

| Ator | Papel sugerido | Acesso |
|------|----------------|--------|
| Administrador | `ADMIN` | Tudo, incl. usuários |
| Vendedor | `VENDEDOR` | Vendas, clientes, consulta de produtos |
| Estoquista | `ESTOQUISTA` | Compras, ajustes, produtos, fornecedores |
