-- =====================================================================
-- 000009_seed.up.sql  —  Dados iniciais (idempotente)
-- Categorias de exemplo, permissões base, papel ADMIN e usuário admin.
-- A senha abaixo é um hash bcrypt de "admin123" — TROQUE em produção.
-- =====================================================================

-- Categorias (exemplo dos diagramas) ----------------------------------
INSERT INTO catalogo.categorias (desc_cat) VALUES
    ('Capa'), ('Película'), ('Carregador'),
    ('Cabo tipo C'), ('Cabo USB Mini B'), ('Cabo USB Micro B')
ON CONFLICT (desc_cat) DO NOTHING;

-- Permissões base (codigo "<recurso>:<acao>") -------------------------
INSERT INTO iam.permissoes (codigo_per, desc_per) VALUES
    ('clientes:read','Consultar clientes'),       ('clientes:write','Manter clientes'),
    ('fornecedores:read','Consultar fornecedores'),('fornecedores:write','Manter fornecedores'),
    ('catalogo:read','Consultar catálogo'),        ('catalogo:write','Manter catálogo'),
    ('compras:read','Consultar compras'),          ('compras:write','Registrar compras'),
    ('vendas:read','Consultar vendas'),            ('vendas:write','Registrar vendas'),
    ('estoque:read','Consultar estoque'),          ('estoque:write','Lançar ajustes'),
    ('relatorios:read','Consultar relatórios'),    ('iam:admin','Administrar usuários e papéis')
ON CONFLICT (codigo_per) DO NOTHING;

-- Papel ADMIN ---------------------------------------------------------
INSERT INTO iam.papeis (nome_pap, desc_pap) VALUES
    ('ADMIN','Acesso total ao sistema')
ON CONFLICT (nome_pap) DO NOTHING;

-- ADMIN recebe todas as permissões ------------------------------------
INSERT INTO iam.papel_permissoes (pap_id, per_id)
SELECT p.id_pap, perm.id_per
FROM iam.papeis p CROSS JOIN iam.permissoes perm
WHERE p.nome_pap = 'ADMIN'
ON CONFLICT DO NOTHING;

-- Usuário admin -------------------------------------------------------
INSERT INTO iam.usuarios (nome_usr, email_usr, senha_hash_usr)
VALUES ('Administrador','admin@loja.local',
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy')  -- "admin123"
ON CONFLICT (email_usr) DO NOTHING;

-- Vincula admin ao papel ADMIN ----------------------------------------
INSERT INTO iam.usuario_papeis (usr_id, pap_id)
SELECT u.id_usr, p.id_pap
FROM iam.usuarios u, iam.papeis p
WHERE u.email_usr = 'admin@loja.local' AND p.nome_pap = 'ADMIN'
ON CONFLICT DO NOTHING;
