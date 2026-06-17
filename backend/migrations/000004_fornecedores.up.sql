-- =====================================================================
-- 000004_fornecedores.up.sql  —  Contexto Fornecedores
-- Regras: CNPJ validado e UNIQUE; obrigatórios CNPJ, Razão Social,
-- Nome Fantasia, Email, Telefone 1, Contato Comercial.
-- Endereço preenchido via consulta de CEP.
-- =====================================================================

CREATE TABLE fornecedores.fornecedores (
    id_for          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cnpj_for        VARCHAR(14)  NOT NULL UNIQUE,     -- somente dígitos
    razao_for       VARCHAR(160) NOT NULL,
    nome_fant_for   VARCHAR(160) NOT NULL,
    email_for       CITEXT       NOT NULL,
    tel1_for        VARCHAR(20)  NOT NULL,
    tel2_for        VARCHAR(20),
    cep_for         VARCHAR(8),
    rua_for         VARCHAR(160),
    num_for         VARCHAR(20),
    comp_for        VARCHAR(80),
    bai_for         VARCHAR(80),
    cit_for         VARCHAR(80),
    uf_for          CHAR(2),
    comercial_for   VARCHAR(120) NOT NULL,            -- contato comercial
    financeiro_for  VARCHAR(120),                     -- contato financeiro
    dt_ult_comp_for TIMESTAMPTZ,
    atv_for         BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_cnpj_for_digits CHECK (cnpj_for ~ '^[0-9]{14}$')
);

CREATE INDEX idx_fornecedores_razao ON fornecedores.fornecedores(razao_for);

CREATE TRIGGER trg_fornecedores_updated BEFORE UPDATE ON fornecedores.fornecedores
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();
