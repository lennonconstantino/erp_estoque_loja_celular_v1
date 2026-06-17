-- =====================================================================
-- 000001_init.up.sql
-- Inicialização do banco: extensões e schemas (1 schema por bounded context).
--
-- Estratégia de isolamento: cada domínio vive em seu próprio schema.
-- NÃO existem foreign keys entre schemas diferentes. A integridade
-- referencial entre contextos é responsabilidade da aplicação (e, no
-- futuro de microsserviços, de eventos/sagas). Isso permite extrair
-- cada schema para um banco/serviço independente sem reescrever DDL.
-- =====================================================================

-- Extensões -----------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";   -- uuid_generate_v4()
CREATE EXTENSION IF NOT EXISTS "citext";       -- e-mails case-insensitive

-- Schemas (bounded contexts) ------------------------------------------
CREATE SCHEMA IF NOT EXISTS iam;            -- usuários, autenticação, RBAC
CREATE SCHEMA IF NOT EXISTS clientes;       -- cadastro de clientes
CREATE SCHEMA IF NOT EXISTS fornecedores;   -- cadastro de fornecedores
CREATE SCHEMA IF NOT EXISTS catalogo;       -- categorias + produtos
CREATE SCHEMA IF NOT EXISTS compras;        -- compras (entrada de estoque)
CREATE SCHEMA IF NOT EXISTS vendas;         -- vendas (saída de estoque)
CREATE SCHEMA IF NOT EXISTS estoque;        -- ajustes + razão (ledger) de movimentações

-- Função utilitária de auditoria (updated_at) -------------------------
CREATE OR REPLACE FUNCTION public.set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
