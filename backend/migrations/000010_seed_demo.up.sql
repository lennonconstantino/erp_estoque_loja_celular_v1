-- =====================================================================
-- 000010_seed_demo.up.sql  —  Dados de demonstração (idempotente)
-- Popula fornecedores, clientes, produtos, compras e vendas de exemplo
-- para que as telas do frontend exibam informações ao entrar no sistema.
-- CPFs válidos (dígito verificador conferido); CNPJs usam formato livre.
-- UUIDs fixos com prefixo numérico distinto por entidade.
-- =====================================================================

-- ─── Fornecedores adicionais ─────────────────────────────────────────
INSERT INTO fornecedores.fornecedores
    (id_for, cnpj_for, razao_for, nome_fant_for, email_for, tel1_for, comercial_for,
     cit_for, uf_for, atv_for)
VALUES
    ('11000001-0000-0000-0000-000000000001',
     '12345678000195', 'TechPhone Distribuidora LTDA', 'TechPhone',
     'vendas@techphone.com.br', '11940001001', 'Pedro Alves',
     'São Paulo', 'SP', TRUE),
    ('11000001-0000-0000-0000-000000000002',
     '99887766000100', 'MobileCase Importadora LTDA', 'MobileCase',
     'contato@mobilecase.com.br', '21940002002', 'Fernanda Lima',
     'Rio de Janeiro', 'RJ', TRUE)
ON CONFLICT (id_for) DO NOTHING;

-- ─── Clientes ────────────────────────────────────────────────────────
-- CPFs: 529.982.247-25 / 111.444.777-35 / 123.456.789-09 / 987.654.321-00
INSERT INTO clientes.clientes
    (id_cli, cpf_cli, nome_cli, email_cli, tel_cli, cit_cli, uf_cli, atv_cli)
VALUES
    ('22000001-0000-0000-0000-000000000001',
     '52998224725', 'Maria Silva', 'maria.silva@email.com', '11991110001',
     'São Paulo', 'SP', TRUE),
    ('22000001-0000-0000-0000-000000000002',
     '11144477735', 'João Santos', 'joao.santos@email.com', '11991110002',
     'Campinas', 'SP', TRUE),
    ('22000001-0000-0000-0000-000000000003',
     '12345678909', 'Ana Rodrigues', 'ana.rodrigues@email.com', '21991110003',
     'Rio de Janeiro', 'RJ', TRUE),
    ('22000001-0000-0000-0000-000000000004',
     '98765432100', 'Carlos Ferreira', 'carlos.ferreira@email.com', '31991110004',
     'Belo Horizonte', 'MG', TRUE)
ON CONFLICT (id_cli) DO NOTHING;

-- ─── Produtos de demonstração ─────────────────────────────────────────
-- Referencia categorias pelo nome (já existem no seed 000009).
INSERT INTO catalogo.produtos
    (id_pro, cat_pro, desc_pro, p_custo_pro, p_venda_pro, estoque_m_pro,
     estoque_a_pro, garant_pro, mod_pro, disp_pro, atv_pro)
SELECT
    '33000001-0000-0000-0000-000000000001',
    id_cat, 'Película de Vidro iPhone 15 Pro', 8.00, 39.90,
    5, 18, 6, 'iPhone 15 Pro', TRUE, TRUE
FROM catalogo.categorias WHERE desc_cat = 'Película'
ON CONFLICT (id_pro) DO NOTHING;

INSERT INTO catalogo.produtos
    (id_pro, cat_pro, desc_pro, p_custo_pro, p_venda_pro, estoque_m_pro,
     estoque_a_pro, garant_pro, mod_pro, disp_pro, atv_pro)
SELECT
    '33000001-0000-0000-0000-000000000002',
    id_cat, 'Carregador USB-C 65W Turbo', 28.00, 99.90,
    3, 12, 12, 'Universal', TRUE, TRUE
FROM catalogo.categorias WHERE desc_cat = 'Carregador'
ON CONFLICT (id_pro) DO NOTHING;

INSERT INTO catalogo.produtos
    (id_pro, cat_pro, desc_pro, p_custo_pro, p_venda_pro, estoque_m_pro,
     estoque_a_pro, garant_pro, mod_pro, disp_pro, atv_pro)
SELECT
    '33000001-0000-0000-0000-000000000003',
    id_cat, 'Cabo USB-C 2m 3A Trançado', 9.00, 34.90,
    10, 30, 6, 'Universal', TRUE, TRUE
FROM catalogo.categorias WHERE desc_cat = 'Cabo tipo C'
ON CONFLICT (id_pro) DO NOTHING;

INSERT INTO catalogo.produtos
    (id_pro, cat_pro, desc_pro, p_custo_pro, p_venda_pro, estoque_m_pro,
     estoque_a_pro, garant_pro, mod_pro, disp_pro, atv_pro)
SELECT
    '33000001-0000-0000-0000-000000000004',
    id_cat, 'Capa Anti-Impacto Galaxy S24', 12.00, 49.90,
    4, 9, 12, 'Galaxy S24', TRUE, TRUE
FROM catalogo.categorias WHERE desc_cat = 'Capa'
ON CONFLICT (id_pro) DO NOTHING;

-- ─── Movimentações de estoque (ledger das entradas acima) ─────────────
INSERT INTO estoque.movimentacoes
    (id_mov, pro_mov, tipo_mov, qtd_mov, saldo_ant_mov, saldo_atu_mov,
     origem_tipo, dt_mov)
VALUES
    ('44000001-0000-0000-0000-000000000001',
     '33000001-0000-0000-0000-000000000001',
     'COMPRA', 18, 0, 18, 'COMPRA', now() - interval '30 days'),
    ('44000001-0000-0000-0000-000000000002',
     '33000001-0000-0000-0000-000000000002',
     'COMPRA', 12, 0, 12, 'COMPRA', now() - interval '30 days'),
    ('44000001-0000-0000-0000-000000000003',
     '33000001-0000-0000-0000-000000000003',
     'COMPRA', 30, 0, 30, 'COMPRA', now() - interval '30 days'),
    ('44000001-0000-0000-0000-000000000004',
     '33000001-0000-0000-0000-000000000004',
     'COMPRA', 9, 0, 9, 'COMPRA', now() - interval '30 days')
ON CONFLICT (id_mov) DO NOTHING;

-- ─── Compras de demonstração ──────────────────────────────────────────
INSERT INTO compras.compra_master
    (id_compra, for_compra, nf_compra, dt_compra, val_compra, status_compra)
VALUES
    ('55000001-0000-0000-0000-000000000001',
     '11000001-0000-0000-0000-000000000001',
     'NF-2025-001', '2025-12-01', 856.00, 'CONFIRMADA'),
    ('55000001-0000-0000-0000-000000000002',
     '11000001-0000-0000-0000-000000000002',
     'NF-2025-002', '2025-12-15', 270.00, 'CONFIRMADA')
ON CONFLICT (id_compra) DO NOTHING;

INSERT INTO compras.detalhe_compras
    (id_dt_compra, compra_id, pro_dt_compra, qtd_dt_compra,
     pre_compra_dt_compra, pre_venda_dt_compra, margem_dt_compra)
VALUES
    ('66000001-0000-0000-0000-000000000001',
     '55000001-0000-0000-0000-000000000001',
     '33000001-0000-0000-0000-000000000001', 18, 8.00, 39.90, 398.75),
    ('66000001-0000-0000-0000-000000000002',
     '55000001-0000-0000-0000-000000000001',
     '33000001-0000-0000-0000-000000000002', 12, 28.00, 99.90, 256.79),
    ('66000001-0000-0000-0000-000000000003',
     '55000001-0000-0000-0000-000000000002',
     '33000001-0000-0000-0000-000000000003', 30, 9.00, 34.90, 287.78)
ON CONFLICT (id_dt_compra) DO NOTHING;

-- ─── Vendas de demonstração ───────────────────────────────────────────
INSERT INTO vendas.venda_master
    (id_venda, dt_venda, val_venda, dsc_venda, forma_pgto_venda,
     cli_venda, c_final_venda, doc_fiscal_venda, status_venda)
VALUES
    ('77000001-0000-0000-0000-000000000001',
     now() - interval '10 days',
     149.70, 0.00, 'PIX',
     '22000001-0000-0000-0000-000000000001', FALSE, 'CUPOM', 'CONFIRMADA'),
    ('77000001-0000-0000-0000-000000000002',
     now() - interval '5 days',
     84.80, 5.00, 'CREDITO',
     '22000001-0000-0000-0000-000000000002', FALSE, 'CUPOM', 'CONFIRMADA'),
    ('77000001-0000-0000-0000-000000000003',
     now() - interval '2 days',
     99.90, 0.00, 'DINHEIRO',
     NULL, TRUE, 'CUPOM', 'CONFIRMADA')
ON CONFLICT (id_venda) DO NOTHING;

INSERT INTO vendas.detalhe_vendas
    (id_dt_venda, venda_id, pro_venda, qtd_venda, pre_venda_dt)
VALUES
    ('88000001-0000-0000-0000-000000000001',
     '77000001-0000-0000-0000-000000000001',
     '33000001-0000-0000-0000-000000000001', 2, 39.90),
    ('88000001-0000-0000-0000-000000000002',
     '77000001-0000-0000-0000-000000000001',
     '33000001-0000-0000-0000-000000000003', 2, 34.90),
    ('88000001-0000-0000-0000-000000000003',
     '77000001-0000-0000-0000-000000000002',
     '33000001-0000-0000-0000-000000000004', 1, 49.90),
    ('88000001-0000-0000-0000-000000000004',
     '77000001-0000-0000-0000-000000000002',
     '33000001-0000-0000-0000-000000000001', 1, 39.90),
    ('88000001-0000-0000-0000-000000000005',
     '77000001-0000-0000-0000-000000000003',
     '33000001-0000-0000-0000-000000000002', 1, 99.90)
ON CONFLICT (id_dt_venda) DO NOTHING;

-- ─── Atualizar dt_ult_comp_for dos fornecedores ───────────────────────
UPDATE fornecedores.fornecedores
SET dt_ult_comp_for = '2025-12-01'
WHERE id_for = '11000001-0000-0000-0000-000000000001'
  AND dt_ult_comp_for IS NULL;

UPDATE fornecedores.fornecedores
SET dt_ult_comp_for = '2025-12-15'
WHERE id_for = '11000001-0000-0000-0000-000000000002'
  AND dt_ult_comp_for IS NULL;

-- ─── Atualizar dt_ult_comp_cli dos clientes ───────────────────────────
UPDATE clientes.clientes
SET dt_ult_comp_cli = now() - interval '10 days'
WHERE id_cli = '22000001-0000-0000-0000-000000000001'
  AND dt_ult_comp_cli IS NULL;

UPDATE clientes.clientes
SET dt_ult_comp_cli = now() - interval '5 days'
WHERE id_cli = '22000001-0000-0000-0000-000000000002'
  AND dt_ult_comp_cli IS NULL;
