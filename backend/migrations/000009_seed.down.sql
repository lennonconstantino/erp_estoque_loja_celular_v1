-- 000009_seed.down.sql
DELETE FROM iam.usuario_papeis WHERE usr_id IN (SELECT id_usr FROM iam.usuarios WHERE email_usr = 'admin@loja.local');
DELETE FROM iam.usuarios WHERE email_usr = 'admin@loja.local';
DELETE FROM iam.papel_permissoes WHERE pap_id IN (SELECT id_pap FROM iam.papeis WHERE nome_pap = 'ADMIN');
DELETE FROM iam.papeis WHERE nome_pap = 'ADMIN';
DELETE FROM iam.permissoes;
DELETE FROM catalogo.categorias WHERE desc_cat IN ('Capa','Película','Carregador','Cabo tipo C','Cabo USB Mini B','Cabo USB Micro B');
