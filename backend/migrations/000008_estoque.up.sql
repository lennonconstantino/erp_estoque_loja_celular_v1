-- =====================================================================
-- 000008_estoque.up.sql  —  Contexto Estoque (Ajustes + Razão/Ledger)
--
-- "movimentacoes" é o LIVRO-RAZÃO de estoque: fonte da verdade auditável
-- de toda entrada/saída de saldo. catalogo.produtos.estoque_a_pro é
-- apenas o saldo materializado (cache), atualizado pela aplicação a cada
-- movimentação. Compras, Vendas e Ajustes geram movimentações aqui.
--
-- Ajustes (tela "Ajustar Estoque"):
--   * estornar ENTRADA  => dá baixa  no saldo  (movimentação AJUSTE_SAIDA)
--   * estornar SAÍDA     => acrescenta no saldo  (movimentação AJUSTE_ENTRADA)
--   * SOMENTE consulta e inclusão: sem UPDATE, sem DELETE (regra de app).
-- =====================================================================

CREATE TYPE estoque.tipo_mov AS ENUM (
    'COMPRA',          -- entrada por compra
    'VENDA',           -- saída por venda
    'AJUSTE_ENTRADA',  -- estorno de saída (acrescenta)
    'AJUSTE_SAIDA'     -- estorno de entrada (dá baixa)
);

-- Razão de movimentações (append-only) --------------------------------
CREATE TABLE estoque.movimentacoes (
    id_mov        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pro_mov       UUID NOT NULL,                       -- catalogo.produtos.id_pro (sem FK)
    tipo_mov      estoque.tipo_mov NOT NULL,
    qtd_mov       INTEGER NOT NULL,                    -- sempre positivo; sinal vem do tipo
    saldo_ant_mov INTEGER NOT NULL,
    saldo_atu_mov INTEGER NOT NULL,
    origem_tipo   VARCHAR(20),                         -- 'COMPRA' | 'VENDA' | 'AJUSTE'
    origem_id     UUID,                                -- id do documento de origem
    resp_mov      UUID,                                -- iam.usuarios.id_usr (sem FK)
    dt_mov        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_qtd_mov_pos CHECK (qtd_mov > 0)
);
CREATE INDEX idx_movimentacoes_pro  ON estoque.movimentacoes(pro_mov);
CREATE INDEX idx_movimentacoes_dt   ON estoque.movimentacoes(dt_mov);
CREATE INDEX idx_movimentacoes_orig ON estoque.movimentacoes(origem_tipo, origem_id);

-- Ajustes de estoque (documento da tela) ------------------------------
CREATE TABLE estoque.ajustes (
    id_ajs          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pro_ajs         UUID NOT NULL,                     -- catalogo.produtos.id_pro (sem FK)
    qtd_entrada_ajs INTEGER NOT NULL DEFAULT 0,        -- estorno de saída
    qtd_saida_ajs   INTEGER NOT NULL DEFAULT 0,        -- estorno de entrada
    mot_ajs         VARCHAR(255) NOT NULL,             -- motivo (obrigatório)
    resp_ajs        UUID NOT NULL,                     -- iam.usuarios.id_usr (sem FK)
    dt_ajs          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_ajs_nao_neg CHECK (qtd_entrada_ajs >= 0 AND qtd_saida_ajs >= 0),
    CONSTRAINT chk_ajs_algum   CHECK (qtd_entrada_ajs > 0 OR qtd_saida_ajs > 0)
);
CREATE INDEX idx_ajustes_pro ON estoque.ajustes(pro_ajs);

-- Proteção física da regra "somente consulta e inclusão" --------------
CREATE OR REPLACE FUNCTION estoque.bloqueia_alteracao()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Ajustes/movimentações são imutáveis: somente inclusão e consulta.';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_ajustes_no_update BEFORE UPDATE OR DELETE ON estoque.ajustes
    FOR EACH ROW EXECUTE FUNCTION estoque.bloqueia_alteracao();
CREATE TRIGGER trg_movimentacoes_no_update BEFORE UPDATE OR DELETE ON estoque.movimentacoes
    FOR EACH ROW EXECUTE FUNCTION estoque.bloqueia_alteracao();
