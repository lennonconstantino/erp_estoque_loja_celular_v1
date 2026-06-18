#!/usr/bin/env bash
# teardown.sh — ERP Estoque — Para e remove os recursos Docker do projeto
#
# Uso (a partir de qualquer diretório):
#   ./scripts/docker/teardown.sh           # remove containers, volume (pgdata) e rede
#   ./scripts/docker/teardown.sh --images  # idem + apaga as imagens buildadas (api, frontend)
set -uo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'
RED='\033[0;31m'; BOLD='\033[1m'; DIM='\033[2m'; NC='\033[0m'

step() { echo -e "${YELLOW}▶ $1${NC}"; }
ok()   { echo -e "  ${GREEN}✓${NC} $1"; }
err()  { echo -e "  ${RED}✗${NC} $1"; }
info() { echo -e "  ${DIM}$1${NC}"; }

# ── Localiza a raiz do projeto (compose vive na raiz) ──────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
PROJECT="$(basename "$ROOT_DIR")"   # nome do projeto compose (default = nome do diretório)

REMOVE_IMAGES=false
for arg in "$@"; do
  [ "$arg" = "--images" ] && REMOVE_IMAGES=true
done

echo
echo -e "${CYAN}${BOLD}"
echo "  ╔══════════════════════════════════════════════════════╗"
echo "  ║   ERP Estoque — Teardown                              ║"
echo "  ╚══════════════════════════════════════════════════════╝"
echo -e "${NC}"

if ! docker info >/dev/null 2>&1; then
  err "Docker não está rodando — nada a remover."
  exit 0
fi

if [ ! -f "$COMPOSE_FILE" ]; then
  err "docker-compose.yml não encontrado em $ROOT_DIR"
  exit 1
fi

COMPOSE=(docker compose -f "$COMPOSE_FILE")

# ── 1. Containers + volume nomeado (pgdata) + rede ─────────────────────────────
step "Parando e removendo containers, volume e rede ($PROJECT)..."
info "docker compose down -v --remove-orphans"
if "${COMPOSE[@]}" down -v --remove-orphans 2>&1 | grep -iE '(stopping|removing|removed|stopped|error)' | sed 's/^/  /'; then
  ok "Containers (db, migrate, api, frontend), volume pgdata e rede removidos"
else
  ok "Nenhum container ativo encontrado"
fi
echo

# ── 2. Imagens buildadas (opcional) ───────────────────────────────────────────
if [ "$REMOVE_IMAGES" = true ]; then
  step "Removendo imagens buildadas do projeto..."
  IMAGES=(
    "${PROJECT}-api"
    "${PROJECT}-frontend"
  )
  removed=0
  for img in "${IMAGES[@]}"; do
    if docker image inspect "$img" >/dev/null 2>&1; then
      docker rmi "$img" >/dev/null 2>&1 && ok "$img removida" && ((removed++)) || err "falha ao remover $img"
    else
      info "$img não encontrada (já removida ou nunca buildada)"
    fi
  done
  [ "$removed" -gt 0 ] && ok "$removed imagem(ns) removida(s)" || info "Nenhuma imagem removida"
  info "Imagens-base (postgres, migrate/migrate) são preservadas; remova-as manualmente se desejar."
  echo
fi

# ── 3. Verificação final ───────────────────────────────────────────────────────
step "Verificando se restam recursos do projeto $PROJECT..."
remaining=$(docker ps -a --filter "label=com.docker.compose.project=$PROJECT" --format "{{.Names}}" 2>/dev/null)
if [ -z "$remaining" ]; then
  ok "Nenhum container remanescente"
else
  err "Containers ainda presentes:"
  echo "$remaining" | sed 's/^/    /'
fi

volumes_left=$(docker volume ls --filter "label=com.docker.compose.project=$PROJECT" --format "{{.Name}}" 2>/dev/null)
if [ -z "$volumes_left" ]; then
  ok "Nenhum volume remanescente (pgdata removido)"
else
  err "Volumes ainda presentes: $volumes_left"
fi

network_left=$(docker network ls --filter "name=${PROJECT}_default" --format "{{.Name}}" 2>/dev/null)
if [ -z "$network_left" ]; then
  ok "Rede ${PROJECT}_default removida"
else
  err "Rede ainda presente: $network_left"
fi

echo
echo -e "  ${BOLD}Recursos liberados:${NC}"
info "  · 4 containers (db, migrate, api, frontend)"
info "  · 1 banco PostgreSQL (volume pgdata — dados apagados)"
info "  · Rede Docker ${PROJECT}_default"
if [ "$REMOVE_IMAGES" = true ]; then
  info "  · Imagens buildadas: ${PROJECT}-api, ${PROJECT}-frontend"
fi
echo
info "Para subir tudo novamente: make up"
if [ "$REMOVE_IMAGES" = false ]; then
  info "Para remover também as imagens buildadas: ./scripts/docker/teardown.sh --images"
fi
echo
