#!/usr/bin/env bash
# =============================================================================
# supabase-setup.sh — cria e popula o banco no Supabase (ou qualquer Postgres).
#
# Aplica TODAS as migrations (`migrate up`): cria os schemas/tabelas e popula os
# seeds — usuário admin (000009) e dados de demonstração (000010). É idempotente:
# rodar de novo num banco já migrado não faz nada (no-op).
#
# Uso:
#   scripts/supabase-setup.sh                 # usa backend/.env.production
#   scripts/supabase-setup.sh "<DATABASE_URL>"# URL explícita
#   scripts/supabase-setup.sh -y              # sem confirmação interativa
#   AUTO_YES=1 scripts/supabase-setup.sh      # idem (para CI)
#
# Resolução da URL: 1º argumento  >  backend/.env.production  >  $DATABASE_URL
# =============================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND="$ROOT/backend"

# --- separa flags de argumentos posicionais ---
AUTO="${AUTO_YES:-}"
URL_ARG=""
for arg in "$@"; do
  case "$arg" in
    -y|--yes) AUTO=1 ;;
    -*) echo "flag desconhecida: $arg" >&2; exit 2 ;;
    *) URL_ARG="$arg" ;;
  esac
done

# --- resolve DATABASE_URL ---
# Precedência: argumento explícito > backend/.env.production > $DATABASE_URL.
# O .env.production vem ANTES do ambiente de propósito: rodando via `make`, o
# Makefile exporta o DATABASE_URL local (backend/.env) e ele não deve sequestrar
# um script cujo alvo é produção. Para mirar outro banco, passe a URL como argumento.
DB_URL="$URL_ARG"
if [[ -z "$DB_URL" && -f "$BACKEND/.env.production" ]]; then
  DB_URL="$(grep -E '^DATABASE_URL=' "$BACKEND/.env.production" | head -1 | cut -d= -f2-)"
fi
DB_URL="${DB_URL:-${DATABASE_URL:-}}"
if [[ -z "$DB_URL" ]]; then
  echo "erro: DATABASE_URL não encontrada." >&2
  echo "      passe como argumento, exporte a variável, ou preencha backend/.env.production." >&2
  exit 1
fi

# host:porta/db sem expor a senha
HOST_DISPLAY="$(printf '%s' "$DB_URL" | sed -E 's#^[a-zA-Z]+://[^@]*@##; s#\?.*$##')"

echo "──────────────────────────────────────────────"
echo " Alvo : $HOST_DISPLAY"
echo " Ação : criar schemas + popular seed/demo (migrate up)"
echo "──────────────────────────────────────────────"
case "$DB_URL" in
  *localhost*|*127.0.0.1*) echo "  (aviso: a URL aponta para um banco LOCAL)";;
esac

# --- confirmação ---
if [[ "$AUTO" != "1" ]]; then
  read -r -p "Confirmar? [y/N] " ans
  [[ "$ans" == "y" || "$ans" == "Y" ]] || { echo "abortado."; exit 0; }
fi

# --- compila o runner ---
echo "==> compilando o runner de migrations..."
( cd "$BACKEND" && go build -o "$BACKEND/bin/migrate" ./cmd/migrate )
MIGRATE="$BACKEND/bin/migrate"
export DATABASE_URL="$DB_URL"

# --- aplica ---
echo "==> versão atual:"
"$MIGRATE" -path "$BACKEND/migrations" version || true
echo "==> aplicando migrations (up)..."
"$MIGRATE" -path "$BACKEND/migrations" up
echo "==> versão final:"
"$MIGRATE" -path "$BACKEND/migrations" version

# --- verificação (best-effort: só se houver psql) ---
if command -v psql >/dev/null 2>&1; then
  echo "==> verificação:"
  psql "$DB_URL" -X -q -v ON_ERROR_STOP=0 <<'SQL' || true
\echo  · schemas de negócio:
select string_agg(schema_name, ', ' order by schema_name)
  from information_schema.schemata
 where schema_name in ('iam','clientes','fornecedores','catalogo','estoque','compras','vendas');
\echo  · usuário admin:
select email_usr from iam.usuarios where email_usr = 'admin@loja.local';
\echo  · contagens (dados de demonstração):
select
  (select count(*) from catalogo.produtos)       as produtos,
  (select count(*) from clientes.clientes)        as clientes,
  (select count(*) from fornecedores.fornecedores) as fornecedores,
  (select count(*) from vendas.venda_master)      as vendas;
SQL
else
  echo "(psql não encontrado — verificação detalhada pulada; a 'versão final' acima confirma o sucesso)"
fi

echo ""
echo "✅ Banco criado e populado em $HOST_DISPLAY"
echo "   Login inicial: admin@loja.local / admin123 (troque em produção)."
