-- =====================================================================
-- 000002_iam.up.sql  —  Contexto IAM (Login/Auth + Usuários + RBAC)
-- Autenticação JWT + autorização baseada em papéis (RBAC).
-- =====================================================================

-- Usuários ------------------------------------------------------------
CREATE TABLE iam.usuarios (
    id_usr           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome_usr         VARCHAR(120)  NOT NULL,
    email_usr        CITEXT        NOT NULL UNIQUE,
    senha_hash_usr   VARCHAR(255)  NOT NULL,          -- bcrypt/argon2
    atv_usr          BOOLEAN       NOT NULL DEFAULT TRUE,
    dt_ult_acesso_usr TIMESTAMPTZ,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT now()
);

-- Papéis (roles) ------------------------------------------------------
CREATE TABLE iam.papeis (
    id_pap     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome_pap   VARCHAR(60)  NOT NULL UNIQUE,    -- ex: ADMIN, VENDEDOR, ESTOQUISTA
    desc_pap   VARCHAR(160),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Permissões granulares ----------------------------------------------
CREATE TABLE iam.permissoes (
    id_per     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    codigo_per VARCHAR(80) NOT NULL UNIQUE,     -- ex: "clientes:create", "vendas:read"
    desc_per   VARCHAR(160)
);

-- N:N papel <-> permissão --------------------------------------------
CREATE TABLE iam.papel_permissoes (
    pap_id UUID NOT NULL REFERENCES iam.papeis(id_pap)     ON DELETE CASCADE,
    per_id UUID NOT NULL REFERENCES iam.permissoes(id_per) ON DELETE CASCADE,
    PRIMARY KEY (pap_id, per_id)
);

-- N:N usuário <-> papel ----------------------------------------------
CREATE TABLE iam.usuario_papeis (
    usr_id UUID NOT NULL REFERENCES iam.usuarios(id_usr) ON DELETE CASCADE,
    pap_id UUID NOT NULL REFERENCES iam.papeis(id_pap)   ON DELETE CASCADE,
    PRIMARY KEY (usr_id, pap_id)
);

-- Refresh tokens (rotação JWT) ---------------------------------------
CREATE TABLE iam.refresh_tokens (
    id_rt       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usr_id      UUID NOT NULL REFERENCES iam.usuarios(id_usr) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL,
    expira_em   TIMESTAMPTZ  NOT NULL,
    revogado    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_usr ON iam.refresh_tokens(usr_id);

CREATE TRIGGER trg_usuarios_updated BEFORE UPDATE ON iam.usuarios
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();
