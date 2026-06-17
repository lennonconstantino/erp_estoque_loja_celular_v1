-- =====================================================================
-- 000005_catalogo.up.sql  —  Contexto Catálogo (Categorias + Produtos)
-- Regras de produto:
--   * preço de custo SEMPRE menor que preço de venda;
--   * estoque atual NÃO é ajustado pelo cadastro de produto
--     (somente por Compras/Vendas/Ajustes — ver contexto estoque);
--   * estoque = 0  => produto indisponível (disp_pro = false).
-- =====================================================================

-- Categorias ----------------------------------------------------------
CREATE TABLE catalogo.categorias (
    id_cat     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    desc_cat   VARCHAR(80) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Produtos ------------------------------------------------------------
CREATE TABLE catalogo.produtos (
    id_pro       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cat_pro      UUID NOT NULL REFERENCES catalogo.categorias(id_cat),
    desc_pro     VARCHAR(160) NOT NULL,
    p_custo_pro  NUMERIC(12,2) NOT NULL,
    p_venda_pro  NUMERIC(12,2) NOT NULL,
    estoque_m_pro INTEGER NOT NULL DEFAULT 0,          -- estoque mínimo
    estoque_a_pro INTEGER NOT NULL DEFAULT 0,          -- estoque atual (saldo cacheado)
    garant_pro   INTEGER NOT NULL DEFAULT 0,           -- garantia em meses
    mod_pro      VARCHAR(120),                         -- modelo de celular
    disp_pro     BOOLEAN NOT NULL DEFAULT FALSE,       -- disponível p/ venda
    atv_pro      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_preco_margem   CHECK (p_custo_pro < p_venda_pro),
    CONSTRAINT chk_estoque_nao_neg CHECK (estoque_a_pro >= 0),
    CONSTRAINT chk_estoque_min_nao_neg CHECK (estoque_m_pro >= 0)
);

CREATE INDEX idx_produtos_cat  ON catalogo.produtos(cat_pro);
CREATE INDEX idx_produtos_desc ON catalogo.produtos(desc_pro);

CREATE TRIGGER trg_produtos_updated BEFORE UPDATE ON catalogo.produtos
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();
