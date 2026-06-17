-- =====================================================================
-- 000006_compras.up.sql  —  Contexto Compras (entrada de estoque)
-- Regras: quantidade > 0; preço de custo < preço de venda no item.
-- Ao CONFIRMAR a compra: para cada item, soma a quantidade ao saldo do
-- produto (efetivado via contexto estoque — movimentação tipo COMPRA).
--
-- Referências a fornecedores (for_compra) e produtos (pro_dt_compra) são
-- guardadas por ID, SEM foreign key cross-schema (isolamento de contexto).
-- =====================================================================

CREATE TYPE compras.status_compra AS ENUM ('RASCUNHO', 'CONFIRMADA', 'CANCELADA');

-- Cabeçalho da compra -------------------------------------------------
CREATE TABLE compras.compra_master (
    id_compra     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    for_compra    UUID NOT NULL,                       -- fornecedores.id_for (sem FK)
    nf_compra     VARCHAR(60),
    dt_compra     DATE NOT NULL,
    val_compra    NUMERIC(14,2) NOT NULL DEFAULT 0,    -- total calculado pelos itens
    status_compra compras.status_compra NOT NULL DEFAULT 'RASCUNHO',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_compra_master_for ON compras.compra_master(for_compra);

-- Itens da compra -----------------------------------------------------
CREATE TABLE compras.detalhe_compras (
    id_dt_compra         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    compra_id            UUID NOT NULL REFERENCES compras.compra_master(id_compra) ON DELETE CASCADE,
    pro_dt_compra        UUID NOT NULL,                 -- catalogo.produtos.id_pro (sem FK)
    qtd_dt_compra        INTEGER NOT NULL,
    pre_compra_dt_compra NUMERIC(12,2) NOT NULL,        -- preço de custo
    pre_venda_dt_compra  NUMERIC(12,2) NOT NULL,        -- novo preço de venda sugerido
    margem_dt_compra     NUMERIC(7,2),                  -- % margem (derivado)
    CONSTRAINT chk_qtd_compra_pos   CHECK (qtd_dt_compra > 0),
    CONSTRAINT chk_compra_margem    CHECK (pre_compra_dt_compra < pre_venda_dt_compra)
);
CREATE INDEX idx_detalhe_compras_compra ON compras.detalhe_compras(compra_id);
CREATE INDEX idx_detalhe_compras_pro    ON compras.detalhe_compras(pro_dt_compra);

CREATE TRIGGER trg_compra_master_updated BEFORE UPDATE ON compras.compra_master
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();
