-- 000010_seed_demo.down.sql — Reverte os dados de demonstração

DELETE FROM vendas.detalhe_vendas   WHERE id_dt_venda::text LIKE '88000001%';
DELETE FROM vendas.venda_master     WHERE id_venda::text    LIKE '77000001%';
DELETE FROM compras.detalhe_compras WHERE id_dt_compra::text LIKE '66000001%';
DELETE FROM compras.compra_master   WHERE id_compra::text   LIKE '55000001%';
DELETE FROM estoque.movimentacoes   WHERE id_mov::text      LIKE '44000001%';
DELETE FROM catalogo.produtos       WHERE id_pro::text      LIKE '33000001%';
DELETE FROM clientes.clientes       WHERE id_cli::text      LIKE '22000001%';
DELETE FROM fornecedores.fornecedores WHERE id_for::text    LIKE '11000001%';
