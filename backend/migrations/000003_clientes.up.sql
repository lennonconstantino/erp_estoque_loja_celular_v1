-- =====================================================================
-- 000003_clientes.up.sql  —  Contexto Clientes
-- Regras: CPF validado e UNIQUE; obrigatórios CPF, Nome, Email.
-- Endereço preenchido via consulta de CEP (rua, bairro, cidade, UF).
-- =====================================================================

CREATE TABLE clientes.clientes (
    id_cli          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cpf_cli         VARCHAR(11)  NOT NULL UNIQUE,     -- somente dígitos
    nome_cli        VARCHAR(120) NOT NULL,
    email_cli       CITEXT       NOT NULL,
    tel_cli         VARCHAR(20),
    cep_cli         VARCHAR(8),
    rua_cli         VARCHAR(160),
    num_cli         VARCHAR(20),
    comp_cli        VARCHAR(80),
    bai_cli         VARCHAR(80),
    cit_cli         VARCHAR(80),
    uf_cli          CHAR(2),
    dt_ult_comp_cli TIMESTAMPTZ,                       -- última compra
    atv_cli         BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_cpf_cli_digits CHECK (cpf_cli ~ '^[0-9]{11}$')
);

CREATE INDEX idx_clientes_nome ON clientes.clientes(nome_cli);

CREATE TRIGGER trg_clientes_updated BEFORE UPDATE ON clientes.clientes
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();
