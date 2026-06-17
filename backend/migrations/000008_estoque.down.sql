-- 000008_estoque.down.sql
DROP TRIGGER IF EXISTS trg_movimentacoes_no_update ON estoque.movimentacoes;
DROP TRIGGER IF EXISTS trg_ajustes_no_update ON estoque.ajustes;
DROP FUNCTION IF EXISTS estoque.bloqueia_alteracao();
DROP TABLE IF EXISTS estoque.ajustes;
DROP TABLE IF EXISTS estoque.movimentacoes;
DROP TYPE  IF EXISTS estoque.tipo_mov;
