# Client Brief — Loja de Acessórios para Celular

## O cliente

Uma **loja de acessórios para celular** de pequeno a médio porte com operação presencial.
O negócio comercializa produtos como capas, películas, carregadores, cabos e acessórios em geral para consumidores finais.
A equipe é enxuta: um administrador, um ou dois vendedores e um estoquista.

## Como a loja ganha dinheiro

- Compra produtos de fornecedores atacadistas e revende com margem.
- A margem é variável por categoria: produtos de giro rápido (cabos, películas) têm margens menores; acessórios de nicho têm margens maiores.
- Parte relevante do faturamento vem de clientes recorrentes — fidelização depende de atendimento ágil e estoque bem gerido.

## Como agrega valor ao cliente

- Vendedores precisam saber, na hora da venda, se o produto tem saldo disponível.
- O estoquista precisa saber quais produtos estão abaixo do mínimo para acionar o fornecedor.
- O administrador precisa enxergar compras, vendas e margem sem abrir planilha.
- A emissão de documento fiscal (cupom ou NF) precisa acontecer dentro do mesmo fluxo de venda, sem trocar de sistema.

## O problema

A operação hoje é controlada por **planilhas manuais e anotações em papel**. Isso gera:

- Estoque desatualizado — vendas são baixadas com atraso ou esquecidas.
- Ruptura frequente — nenhum alerta quando o saldo cai abaixo do mínimo.
- Retrabalho na entrada de mercadoria — preços de custo e venda recalculados manualmente a cada compra.
- Sem histórico confiável — impossível saber quais produtos mais vendem, quais ficam parados, ou qual fornecedor é mais acionado.
- Erros em documentos fiscais — cupons e NFs gerados fora do fluxo de venda, com risco de divergência de valor.

## O que querem

Um **ERP web** acessível pelo navegador, com login por usuário e senha e controle de acesso por papel (`ADMIN`, `VENDEDOR`, `ESTOQUISTA`), que cubra os seguintes módulos:

| Módulo | O que resolve |
|--------|--------------|
| **Clientes** | Cadastro com CPF, endereço preenchido via CEP, histórico de compras |
| **Fornecedores** | Cadastro com CNPJ, contatos comercial e financeiro, histórico de compras |
| **Categorias** | Agrupamento de produtos (Capa, Película, Carregador, Cabo…) |
| **Produtos** | Catálogo com custo, venda, margem calculada e saldo atual |
| **Compras** | Entrada de mercadoria com NF, atualiza saldo automaticamente |
| **Vendas** | PDV com validação de saldo, desconto, cliente ou consumidor final, emissão de cupom/NF |
| **Ajuste de Estoque** | Correção de saldo com motivo e responsável, append-only |
| **Relatórios** | Listagens, produtos abaixo do mínimo, mais/menos vendidos |

## Exemplos de operações típicas

1. Vendedor informa o CPF do cliente → sistema identifica se já existe e preenche os dados; caso contrário, abre o cadastro.
2. Ao lançar um item na venda, o sistema valida em tempo real se há saldo suficiente.
3. Ao confirmar a venda, o saldo é baixado e o documento fiscal (cupom ou NF) é emitido pela API integrada.
4. Estoquista registra entrada de mercadoria: informa fornecedor, número da NF e itens (produto, quantidade, custo, venda); ao confirmar, o saldo é acrescido.
5. Administrador consulta o relatório de produtos abaixo do mínimo antes de acionar fornecedores.
6. Estoquista registra ajuste de estoque por quebra ou devolução, com motivo; o registro não pode ser alterado nem excluído.
7. Administrador emite relatório de vendas do mês para comparar com compras e calcular giro do estoque.

## O que "confiabilidade" significa aqui

O estoque é a principal garantia operacional da loja. O sistema deve:

- **Nunca permitir saldo negativo.** Vendas simultâneas não podem ultrapassar o saldo disponível.
- **Manter rastreabilidade completa.** Cada movimentação de estoque registra saldo anterior e posterior.
- **Garantir consistência fiscal.** O valor da venda confirmado pelo sistema deve ser idêntico ao do documento fiscal emitido.
- **Preservar o histórico de ajustes.** Ajustes são append-only — sem edição e sem exclusão.

Um saldo errado ou um documento fiscal divergente gera problema com o fisco e perda de confiança da equipe no sistema.

## Restrições

- Usuários: equipe da loja (~5 pessoas) com papéis `ADMIN`, `VENDEDOR`, `ESTOQUISTA`.
- Acesso: navegador (desktop), login com e-mail e senha — sem SSO externo.
- Hospedagem: Railway (backend Go + frontend React) + Supabase como banco PostgreSQL gerenciado.
- Não há equipe de infraestrutura — o sistema deve se auto-gerenciar (migrations automáticas no deploy).
- Documento fiscal: integração com API externa de cupom e NF (adaptadores de saída — `FiscalGateway`).
- CEP: preenchimento automático via API pública (adaptador de saída — `CepGateway`).

## Fora do escopo (explicitamente)

- App mobile.
- E-commerce ou venda online.
- Integração com marketplaces (Mercado Livre, Shopee, etc.).
- Multi-loja ou multi-filial.
- Módulo financeiro / contas a pagar e receber.
- Emissão própria de NF-e (usa API de terceiro).
- Relatórios avançados de BI ou dashboards com gráficos.

## Definição de pronto

O sistema está pronto quando a equipe da loja conseguir executar o ciclo completo sem papel ou planilha:

1. Receber mercadoria → registrar compra → saldo atualizado.
2. Atender cliente → registrar venda → saldo baixado → documento fiscal emitido.
3. Identificar ruptura → consultar relatório de mínimos → acionar fornecedor.
4. Corrigir divergência → lançar ajuste de estoque com motivo.
5. Revisar performance → emitir relatório de vendas e produtos.

Critério de aceitação: operação de um dia completo (compras, vendas, ajustes) executada sem recorrer a planilhas, sem saldo negativo e sem erro em documento fiscal.
