#!/usr/bin/env bash
# teardown.sh — ERP Estoque — Para e remove os recursos Docker do projeto
#
# Uso (a partir de qualquer diretório):
#   ./scripts/docker/teardown.sh                                    # containers, volume e rede
#   ./scripts/docker/teardown.sh --images                           # idem + imagens buildadas
#   ./scripts/docker/teardown.sh --recreate-volume                  # idem + recria pgdata vazio
#   ./scripts/docker/teardown.sh --obs                              # idem + stack de observabilidade
#   ./scripts/docker/teardown.sh --obs --images --recreate-volume   # tudo
set -uo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'
RED='\033[0;31m'; BOLD='\033[1m'; DIM='\033[2m'; NC='\033[0m'

step()  { echo -e "${YELLOW}▶ $1${NC}"; }
ok()    { echo -e "  ${GREEN}✓${NC} $1"; }
err()   { echo -e "  ${RED}✗${NC} $1"; }
info()  { echo -e "  ${DIM}$1${NC}"; }
warn()  { echo -e "  ${YELLOW}!${NC} $1"; }

# ── Raiz do projeto ────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
COMPOSE_OBS_FILE="$ROOT_DIR/docker-compose.observability.yml"
ENV_FILE="$ROOT_DIR/backend/.env"
PROJECT="$(basename "$ROOT_DIR")"
VOLUME_NAME="${PROJECT}_pgdata"

# ── Flags ─────────────────────────────────────────────────────────────────────
REMOVE_IMAGES=false
RECREATE_VOLUME=false
TEARDOWN_OBS=false
for arg in "$@"; do
  case "$arg" in
    --images)          REMOVE_IMAGES=true ;;
    --recreate-volume) RECREATE_VOLUME=true ;;
    --obs)             TEARDOWN_OBS=true ;;
    --help|-h)
      echo "Uso: $0 [--images] [--recreate-volume] [--obs]"
      echo "  --images           Remove as imagens buildadas (api, frontend)"
      echo "  --recreate-volume  Recria o volume pgdata vazio após o teardown"
      echo "  --obs              Para e remove também a stack de observabilidade"
      echo "                     (Prometheus + Grafana + seus volumes de dados)"
      exit 0
      ;;
    *) echo "Flag desconhecida: $arg (use --help)"; exit 1 ;;
  esac
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

# ── Monta o comando base do compose ───────────────────────────────────────────
# docker compose exige as variáveis obrigatórias (DB_PASSWORD etc.) mesmo no down.
# Carrega backend/.env se existir; caso contrário injeta dummies para evitar o erro
# de interpolação — os valores não importam no teardown.
COMPOSE=(docker compose -f "$COMPOSE_FILE")
if [ -f "$ENV_FILE" ]; then
  COMPOSE+=(--env-file "$ENV_FILE")
  info "Usando $ENV_FILE"
else
  warn "backend/.env não encontrado — injetando variáveis dummy para o compose"
  export DB_PASSWORD=dummy DB_USER=dummy DB_NAME=dummy JWT_SECRET=dummy
fi

# ── 1. Containers + volume nomeado (pgdata) + rede ────────────────────────────
step "Parando e removendo containers, volume e rede ($PROJECT)..."
info "docker compose down -v --remove-orphans"
compose_output=$("${COMPOSE[@]}" down -v --remove-orphans 2>&1)
compose_exit=$?

if [ $compose_exit -eq 0 ]; then
  echo "$compose_output" | grep -iE '(stopping|removing|removed|stopped)' | sed 's/^/  /' || true
  ok "Containers, volume pgdata e rede removidos via compose"
else
  # Compose falhou — fallback com comandos docker diretos
  warn "Compose retornou erro (código $compose_exit); executando limpeza manual..."
  echo "$compose_output" | sed 's/^/  /' || true

  containers=$(docker ps -a --filter "label=com.docker.compose.project=$PROJECT" \
    --filter "label=com.docker.compose.service" --format "{{.ID}}" 2>/dev/null \
    | grep -v "$(docker ps -a --filter "label=com.docker.compose.config-hash" \
       --filter "label=com.docker.compose.project=$PROJECT" \
       --format "{{.ID}}" 2>/dev/null | head -0)" || \
    docker ps -a --filter "label=com.docker.compose.project=$PROJECT" --format "{{.ID}}" 2>/dev/null)

  if [ -n "$containers" ]; then
    # shellcheck disable=SC2086
    docker stop $containers >/dev/null 2>&1 || true
    # shellcheck disable=SC2086
    docker rm   $containers >/dev/null 2>&1 && ok "Containers removidos manualmente" \
      || err "Falha ao remover containers"
  else
    ok "Nenhum container ativo encontrado"
  fi

  if docker volume inspect "$VOLUME_NAME" >/dev/null 2>&1; then
    docker volume rm "$VOLUME_NAME" >/dev/null 2>&1 \
      && ok "Volume $VOLUME_NAME removido manualmente" \
      || err "Falha ao remover $VOLUME_NAME (pode estar em uso)"
  else
    ok "Volume $VOLUME_NAME já inexistente"
  fi

  network="${PROJECT}_default"
  if docker network inspect "$network" >/dev/null 2>&1; then
    docker network rm "$network" >/dev/null 2>&1 \
      && ok "Rede $network removida manualmente" \
      || err "Falha ao remover rede $network"
  else
    ok "Rede $network já inexistente"
  fi
fi
echo

# ── 2. Stack de observabilidade (opcional) ────────────────────────────────────
if [ "$TEARDOWN_OBS" = true ]; then
  step "Parando e removendo stack de observabilidade (Prometheus + Grafana)..."

  if [ ! -f "$COMPOSE_OBS_FILE" ]; then
    warn "docker-compose.observability.yml não encontrado em $ROOT_DIR — pulando"
  else
    # A stack de obs não tem variáveis obrigatórias; usa defaults do compose.
    COMPOSE_OBS=(docker compose -f "$COMPOSE_OBS_FILE")
    info "docker compose -f docker-compose.observability.yml down -v"
    obs_output=$("${COMPOSE_OBS[@]}" down -v --remove-orphans 2>&1)
    obs_exit=$?

    if [ $obs_exit -eq 0 ]; then
      echo "$obs_output" | grep -iE '(stopping|removing|removed|stopped)' | sed 's/^/  /' || true
      ok "Prometheus, Grafana e volumes de dados removidos"
    else
      warn "Compose obs retornou erro (código $obs_exit); removendo manualmente..."
      echo "$obs_output" | sed 's/^/  /' || true

      for vol in "${PROJECT}_prometheus-data" "${PROJECT}_grafana-data"; do
        if docker volume inspect "$vol" >/dev/null 2>&1; then
          docker volume rm "$vol" >/dev/null 2>&1 \
            && ok "Volume $vol removido" \
            || err "Falha ao remover $vol"
        else
          info "Volume $vol já inexistente"
        fi
      done

      obs_network="${PROJECT}_default"
      # A rede pode já ter sido removida pelo step 1 (mesmo nome de projeto)
      docker network rm "$obs_network" >/dev/null 2>&1 || true
    fi
  fi
  echo
fi

# ── 3. Imagens buildadas (opcional) ───────────────────────────────────────────
if [ "$REMOVE_IMAGES" = true ]; then
  step "Removendo imagens buildadas do projeto..."
  IMAGES=("${PROJECT}-api" "${PROJECT}-frontend")
  removed=0
  for img in "${IMAGES[@]}"; do
    if docker image inspect "$img" >/dev/null 2>&1; then
      if docker rmi "$img" >/dev/null 2>&1; then
        ok "$img removida"
        ((removed++)) || true
      else
        err "Falha ao remover $img"
      fi
    else
      info "$img não encontrada (já removida ou nunca buildada)"
    fi
  done
  [ "$removed" -gt 0 ] && ok "$removed imagem(ns) removida(s)" || info "Nenhuma imagem removida"
  info "Imagens-base (postgres, prometheus, grafana) preservadas — remova manualmente se desejar."
  echo
fi

# ── 4. Recriação do volume pgdata (opcional) ──────────────────────────────────
if [ "$RECREATE_VOLUME" = true ]; then
  step "Recriando volume $VOLUME_NAME vazio..."

  if docker volume inspect "$VOLUME_NAME" >/dev/null 2>&1; then
    docker volume rm "$VOLUME_NAME" >/dev/null 2>&1 \
      && info "Volume anterior removido" \
      || { err "Não foi possível remover $VOLUME_NAME — está em uso?"; exit 1; }
  fi

  if docker volume create "$VOLUME_NAME" >/dev/null 2>&1; then
    ok "Volume $VOLUME_NAME recriado vazio"
    info "Próximo 'make up' rodará as migrations do zero"
  else
    err "Falha ao recriar volume $VOLUME_NAME"
  fi
  echo
fi

# ── 5. Verificação final ──────────────────────────────────────────────────────
step "Verificando recursos remanescentes do projeto $PROJECT..."
has_problem=false

remaining=$(docker ps -a --filter "label=com.docker.compose.project=$PROJECT" --format "{{.Names}}" 2>/dev/null)
if [ -z "$remaining" ]; then
  ok "Nenhum container remanescente"
else
  err "Containers ainda presentes:"; echo "$remaining" | sed 's/^/    /'
  has_problem=true
fi

volumes_left=$(docker volume ls --filter "label=com.docker.compose.project=$PROJECT" --format "{{.Name}}" 2>/dev/null)
if [ "$RECREATE_VOLUME" = true ]; then
  expected_vol="$VOLUME_NAME"
  unexpected=$(echo "$volumes_left" | grep -v "^$expected_vol$" || true)
  if [ -z "$unexpected" ]; then
    ok "Volume $VOLUME_NAME recriado e pronto"
  else
    err "Volumes inesperados: $unexpected"; has_problem=true
  fi
elif [ "$TEARDOWN_OBS" = true ]; then
  # Sem --recreate-volume: nenhum volume deve restar
  if [ -z "$volumes_left" ]; then
    ok "Nenhum volume remanescente (pgdata + obs removidos)"
  else
    err "Volumes ainda presentes: $volumes_left"; has_problem=true
  fi
elif [ -z "$volumes_left" ]; then
  ok "Nenhum volume remanescente (pgdata removido)"
else
  err "Volumes ainda presentes: $volumes_left"; has_problem=true
fi

network_left=$(docker network ls --filter "name=${PROJECT}_default" --format "{{.Name}}" 2>/dev/null)
if [ -z "$network_left" ]; then
  ok "Rede ${PROJECT}_default removida"
else
  err "Rede ainda presente: $network_left"; has_problem=true
fi

# ── Sumário ────────────────────────────────────────────────────────────────────
echo
echo -e "  ${BOLD}Recursos liberados:${NC}"
info "  · Containers: db, migrate, api, frontend"
[ "$TEARDOWN_OBS" = true ] && info "  · Containers obs: prometheus, grafana"
if [ "$RECREATE_VOLUME" = true ]; then
  info "  · Volume $VOLUME_NAME recriado vazio (banco zerado)"
else
  info "  · Volume $VOLUME_NAME removido (dados do PostgreSQL apagados)"
fi
[ "$TEARDOWN_OBS" = true ] && info "  · Volumes obs: prometheus-data, grafana-data"
info "  · Rede ${PROJECT}_default"
[ "$REMOVE_IMAGES" = true ] && info "  · Imagens: ${PROJECT}-api, ${PROJECT}-frontend"
echo

if [ "$has_problem" = true ]; then
  warn "Alguns recursos não puderam ser removidos (veja acima)."
else
  info "Para subir a aplicação: make up"
  [ "$TEARDOWN_OBS" = true ] && \
    info "Para subir a observabilidade: docker compose -f docker-compose.observability.yml up -d"
fi

[ "$REMOVE_IMAGES" = false ]    && info "Para remover também as imagens: $0 --images"
[ "$RECREATE_VOLUME" = false ]  && info "Para recriar o volume pgdata vazio: $0 --recreate-volume"
[ "$TEARDOWN_OBS" = false ]     && info "Para incluir Prometheus/Grafana: $0 --obs"
echo
