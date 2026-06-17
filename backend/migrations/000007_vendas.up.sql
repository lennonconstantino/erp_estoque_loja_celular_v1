-- =====================================================================
-- 000007_vendas.up.sql  —  Contexto Vendas (saída de estoque)
-- Regras:
--   * valor da venda é gerado pelos itens (detalhe); desconto é gravado;
--   * antes de aceitar o item: estoque > 0 e saldo >= qtd da venda;
--   * venda para cliente cadastrado OU consumidor final (c_final_venda);
--   * documento fiscal: CUPOM (API cupom fiscal) ou NF (API nota fiscal);
--   * ao CONFIRMAR a venda: baixa o saldo do produto
--     (saldo = saldo - qtd_venda) via contexto estoque (tipo VENDA).
--
-- cli_venda e pro_venda são guardados por ID, SEM FK cross-schema.
-- =====================================================================

CREATE TYPE vendas.status_venda  AS ENUM ('RASCUNHO', 'CONFIRMADA', 'CANCELADA');
CREATE TYPE vendas.doc_fiscal    AS ENUM ('CUPOM', 'NF');
CREATE TYPE vendas.forma_pgto    AS ENUM ('DINHEIRO', 'PIX', 'DEBITO', 'CREDITO', 'OUTRO');

-- Cabeçalho da venda --------------------------------------------------
CREATE TABLE vendas.venda_master (
    id_venda        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dt_venda        TIMESTAMPTZ NOT NULL DEFAULT now(),
    val_venda       NUMERIC(14,2) NOT NULL DEFAULT 0,  -- total dos itens - desconto
    dsc_venda       NUMERIC(12,2) NOT NULL DEFAULT 0,
    forma_pgto_venda vendas.forma_pgto NOT NULL,
    cli_venda       UUID,                              -- clientes.id_cli (NULL se consumidor final)
    c_final_venda   BOOLEAN NOT NULL DEFAULT FALSE,    -- venda p/ consumidor final
    doc_fiscal_venda vendas.doc_fiscal NOT NULL DEFAULT 'CUPOM',
    status_venda    vendas.status_venda NOT NULL DEFAULT 'RASCUNHO',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_desconto_nao_neg CHECK (dsc_venda >= 0),
    -- consumidor final => sem cliente; cliente informado => não consumidor final
    CONSTRAINT chk_cliente_ou_final CHECK (
        (c_final_venda = TRUE  AND cli_venda IS NULL) OR
        (c_final_venda = FALSE AND cli_venda IS NOT NULL)
    )
);
CREATE INDEX idx_venda_master_cli ON vendas.venda_master(cli_venda);
CREATE INDEX idx_venda_master_dt  ON vendas.venda_master(dt_venda);

-- Itens da venda ------------------------------------------------------
CREATE TABLE vendas.detalhe_vendas (
    id_dt_venda     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    venda_id        UUID NOT NULL REFERENCES vendas.venda_master(id_venda) ON DELETE CASCADE,
    pro_venda       UUID NOT NULL,                     -- catalogo.produtos.id_pro (sem FK)
    qtd_venda       INTEGER NOT NULL,
    pre_venda_dt    NUMERIC(12,2) NOT NULL,            -- preço unitário praticado
    CONSTRAINT chk_qtd_venda_pos CHECK (qtd_venda > 0)
);
CREATE INDEX idx_detalhe_vendas_venda ON vendas.detalhe_vendas(venda_id);
CREATE INDEX idx_detalhe_vendas_pro   ON vendas.detalhe_vendas(pro_venda);

CREATE TRIGGER trg_venda_master_updated BEFORE UPDATE ON vendas.venda_master
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();
