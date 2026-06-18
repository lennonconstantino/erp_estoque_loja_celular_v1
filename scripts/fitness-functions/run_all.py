#!/usr/bin/env python3
"""
Fitness Functions — ERP Estoque (monólito modular, arquitetura hexagonal / Go)

FF1: Boundary Isolation — análise estática das camadas hexagonais (domain/ports/
     application/adapters) e do isolamento entre bounded contexts (módulos).
FF2: Contract Tests     — endpoints reais da API (/health + /api/v1/clientes) e
     enforcement de autenticação/RBAC (401 sem token, 403 sem permissão).
FF3: Latência p99       — carga leve em /health e /api/v1/clientes.
FF4: Chaos / Degradação — derruba o Postgres e valida que o processo sobrevive
     (/health continua 200, rota dependente degrada com 5xx) e se recupera.
     Destrutivo: desabilitado por padrão (ative com FF4_ENABLE=1).

Variáveis de ambiente:
  BASE_URL     (default http://localhost:8080)
  JWT_SECRET   (obrigatória; se ausente, tenta ler de backend/.env)
  P99_LIMIT_MS (default 300)
  FF4_ENABLE   (=1 para rodar o teste de caos que para o container do banco)
"""
import sys, os, base64, hashlib, hmac, json, time, asyncio, subprocess, statistics, pathlib
import httpx

# scripts/fitness-functions/run_all.py → raiz do repositório
ROOT        = pathlib.Path(__file__).resolve().parents[2]
BACKEND     = ROOT / "backend"
MODULE_BASE = "github.com/lennonconstantino/erp_estoque_loja_celular/backend"
PROJECT     = ROOT.name                       # nome do projeto docker compose

BASE_URL     = os.getenv("BASE_URL", "http://localhost:8080")
P99_LIMIT_MS = int(os.getenv("P99_LIMIT_MS", "300"))

def _maybe_load_backend_env() -> None:
    env_path = BACKEND / ".env"
    if not env_path.exists():
        return
    for raw in env_path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        key = key.strip()
        value = value.strip().strip('"').strip("'")
        if key and key not in os.environ and value:
            os.environ[key] = value

if not os.getenv("JWT_SECRET"):
    _maybe_load_backend_env()

JWT_SECRET = os.getenv("JWT_SECRET")
if not JWT_SECRET:
    print("JWT_SECRET ausente. Defina JWT_SECRET no ambiente (ou em backend/.env).")
    sys.exit(2)


# ══════════════════════════════════════════════════════════════════
# Helper — emite um JWT HS256 compatível com platform/auth (sem PyJWT)
# ══════════════════════════════════════════════════════════════════

def _b64(raw: bytes) -> bytes:
    return base64.urlsafe_b64encode(raw).rstrip(b"=")


def make_token(perms: list[str], roles: list[str] | None = None,
               sub: str = "ff-suite", ttl: int = 900) -> str:
    """Gera um access token HS256 com as claims roles/perms/sub/iat/exp."""
    now = int(time.time())
    header  = {"alg": "HS256", "typ": "JWT"}
    payload = {"roles": roles or [], "perms": perms,
               "sub": sub, "iat": now, "exp": now + ttl}
    seg = (_b64(json.dumps(header,  separators=(",", ":")).encode()) + b"."
           + _b64(json.dumps(payload, separators=(",", ":")).encode()))
    sig = hmac.new(JWT_SECRET.encode(), seg, hashlib.sha256).digest()
    return (seg + b"." + _b64(sig)).decode()


def _bearer(perms: list[str]) -> dict:
    return {"Authorization": f"Bearer {make_token(perms)}"}


# ══════════════════════════════════════════════════════════════════
# FF1 — Boundary Isolation (camadas hexagonais + isolamento de módulos)
# ══════════════════════════════════════════════════════════════════
#
# Regras (dependências apontam só para dentro):
#   domain/      → núcleo puro: sem HTTP, SQL, adapters ou platform de infra.
#   ports/       → só interfaces + domain: sem HTTP, SQL ou adapters.
#   application/ → orquestra via portas: sem HTTP nem SQL diretos.
#   qualquer/    → não importa outro módulo (comunicação só por portas).

INFRA_HTTP = ('"net/http"', "go-chi/chi")
INFRA_SQL  = ("jackc/pgx", '"database/sql"', "platform/database")

LAYER_FORBIDDEN = {
    "domain":      INFRA_HTTP + INFRA_SQL + ("/adapters/", "platform/httpserver"),
    "ports":       INFRA_HTTP + INFRA_SQL + ("/adapters/",),
    "application": INFRA_HTTP + INFRA_SQL,
}


def _imports(content: str) -> list[str]:
    """Extrai os caminhos de import de um arquivo Go (bloco ou linha única)."""
    out, in_block = [], False
    for line in content.splitlines():
        s = line.strip()
        if s.startswith("import ("):
            in_block = True
            continue
        if in_block:
            if s == ")":
                in_block = False
                continue
            if '"' in s:
                out.append(s.split('"')[1])
        elif s.startswith("import ") and '"' in s:
            out.append(s.split('"')[1])
    return out


def _layer_of(rel: pathlib.PurePath) -> str | None:
    for part in rel.parts:
        if part in ("domain", "ports", "application", "adapters"):
            return part
    return None


def run_ff1() -> bool:
    print("─" * 60)
    print("FF1 — Boundary Isolation (camadas hexagonais + módulos)")
    print("─" * 60)

    modules_dir = BACKEND / "internal" / "modules"
    if not modules_dir.exists():
        print("  SKIP — backend/internal/modules não encontrado")
        return True

    modules = sorted(p.name for p in modules_dir.iterdir() if p.is_dir())
    all_ok = True

    for mod in modules:
        mod_dir = modules_dir / mod
        violations: list[str] = []

        for go in mod_dir.rglob("*.go"):
            rel    = go.relative_to(BACKEND)
            layer  = _layer_of(go.relative_to(mod_dir))
            imps   = _imports(go.read_text(errors="ignore"))

            for imp in imps:
                # isolamento entre bounded contexts: nunca importar outro módulo
                marker = f"{MODULE_BASE}/internal/modules/"
                if imp.startswith(marker):
                    other = imp[len(marker):].split("/")[0]
                    if other != mod:
                        violations.append(f"    {rel}: importa módulo '{other}' (use portas)")
                        continue

                # regras por camada (comparadas sobre o import com aspas)
                quoted = f'"{imp}"'
                for bad in LAYER_FORBIDDEN.get(layer, ()):
                    if bad in quoted or bad in imp:
                        violations.append(f"    [{layer}] {rel}: import proibido «{imp}»")
                        break

        if violations:
            print(f"  FAIL {mod}")
            for v in violations:
                print(v)
            all_ok = False
        else:
            print(f"  PASS {mod}")

    return all_ok


# ══════════════════════════════════════════════════════════════════
# FF2 — Contract Tests (endpoints reais + RBAC)
# ══════════════════════════════════════════════════════════════════

CONTRACTS = [
    {
        "name": "GET /health retorna status ok",
        "method": "GET", "path": "/health", "headers": {},
        "expected_status": 200, "required_fields": ["status"],
    },
    {
        "name": "GET /api/v1/clientes sem token → 401 (auth exigida)",
        "method": "GET", "path": "/api/v1/clientes/", "headers": {},
        "expected_status": 401,
    },
    {
        "name": "GET /api/v1/clientes com clientes:read → 200 + envelope items",
        "method": "GET", "path": "/api/v1/clientes/",
        "headers": _bearer(["clientes:read"]),
        "expected_status": 200, "required_fields": ["items"],
    },
    {
        "name": "GET /api/v1/clientes sem a permissão → 403 (RBAC)",
        "method": "GET", "path": "/api/v1/clientes/",
        "headers": _bearer(["outro:read"]),
        "expected_status": 403,
    },
]


async def run_ff2() -> bool:
    print("\n" + "─" * 60)
    print(f"FF2 — Contract Tests ({BASE_URL})")
    print("─" * 60)
    all_ok = True
    async with httpx.AsyncClient(base_url=BASE_URL, timeout=8.0) as client:
        for c in CONTRACTS:
            try:
                r = await client.request(c["method"], c["path"],
                                         headers=c.get("headers", {}),
                                         json=c.get("body"))
                passed = r.status_code == c["expected_status"]
                if passed and c.get("required_fields"):
                    body = r.json()
                    passed = all(f in body for f in c["required_fields"])

                print(f"  {'PASS' if passed else 'FAIL'} {c['name']}")
                if not passed:
                    print(f"       status={r.status_code} body={r.text[:120]}")
                    all_ok = False
            except Exception as e:
                print(f"  FAIL {c['name']} — {e}")
                all_ok = False
    return all_ok


# ══════════════════════════════════════════════════════════════════
# FF3 — Latência p99
# ══════════════════════════════════════════════════════════════════

N_REQUESTS  = 50
CONCURRENCY = 5


async def run_ff3() -> bool:
    print("\n" + "─" * 60)
    print(f"FF3 — Latência p99 < {P99_LIMIT_MS}ms ({N_REQUESTS} req, concurrency={CONCURRENCY})")
    print("─" * 60)

    endpoints = [
        ("GET /health",            "/health",            {}),
        ("GET /api/v1/clientes/",  "/api/v1/clientes/",  _bearer(["clientes:read"])),
    ]
    all_ok = True

    async with httpx.AsyncClient(base_url=BASE_URL, timeout=8.0) as client:
        sem = asyncio.Semaphore(CONCURRENCY)

        async def timed(path, headers):
            async with sem:
                start = time.perf_counter()
                try:
                    r = await client.get(path, headers=headers)
                    return (time.perf_counter() - start) * 1000 if r.status_code < 500 else 9999.0
                except Exception:
                    return 9999.0

        for name, path, headers in endpoints:
            latencies = await asyncio.gather(*[timed(path, headers) for _ in range(N_REQUESTS)])
            valid = sorted(l for l in latencies if l < 9000)
            if not valid:
                print(f"  FAIL {name} — todos os requests falharam")
                all_ok = False
                continue
            p99 = valid[min(int(len(valid) * 0.99), len(valid) - 1)]
            avg = statistics.mean(valid)
            ok  = p99 <= P99_LIMIT_MS
            print(f"  {'PASS' if ok else 'FAIL'} {name}  p99={p99:.1f}ms  avg={avg:.1f}ms  ({len(valid)}/{N_REQUESTS} ok)")
            if not ok:
                all_ok = False

    return all_ok


# ══════════════════════════════════════════════════════════════════
# FF4 — Chaos / Degradação graciosa (queda do Postgres)
# ══════════════════════════════════════════════════════════════════

DB_CONTAINER = f"{PROJECT}-db-1"


def _docker(*args) -> subprocess.CompletedProcess:
    return subprocess.run(["docker", *args], capture_output=True, text=True)


async def run_ff4() -> bool:
    print("\n" + "─" * 60)
    print("FF4 — Chaos: degradação graciosa na queda do Postgres")
    print("─" * 60)

    if os.getenv("FF4_ENABLE") != "1":
        print("  SKIP — destrutivo (para o banco). Ative com FF4_ENABLE=1")
        return True

    if _docker("container", "inspect", DB_CONTAINER).returncode != 0:
        print(f"  SKIP — container {DB_CONTAINER!r} não encontrado (suba com `make up`)")
        return True

    read_hdr = _bearer(["clientes:read"])
    async with httpx.AsyncClient(base_url=BASE_URL, timeout=8.0) as client:
        # Baseline
        try:
            base_health = (await client.get("/health")).status_code
            base_data   = (await client.get("/api/v1/clientes/", headers=read_hdr)).status_code
        except Exception as e:
            print(f"  SKIP — API não acessível em {BASE_URL}: {e}")
            return True
        if base_health != 200 or base_data != 200:
            print(f"  SKIP — baseline não saudável (health={base_health} clientes={base_data})")
            return True
        print(f"  Baseline: /health={base_health}  /api/v1/clientes={base_data}")

        # Caos: para o banco
        print(f"  Parando {DB_CONTAINER}...")
        if _docker("stop", DB_CONTAINER).returncode != 0:
            print("  SKIP — falha ao parar o container do banco")
            return True
        await asyncio.sleep(2)

        # Processo deve sobreviver: health up, rota dependente degrada com 5xx
        try:
            health_down = (await client.get("/health")).status_code
        except Exception:
            health_down = 0
        try:
            data_down = (await client.get("/api/v1/clientes/", headers=read_hdr)).status_code
        except Exception:
            data_down = 0

        process_alive = health_down == 200
        degrades_5xx  = data_down >= 500
        print(f"  /health (processo vivo): {'UP ✓' if process_alive else f'{health_down} ✗'} (esperado 200)")
        print(f"  /api/v1/clientes (rota dependente): {data_down} {'✓' if degrades_5xx else '✗'} (esperado 5xx)")

        # Recuperação
        print(f"  Restaurando {DB_CONTAINER}...")
        _docker("start", DB_CONTAINER)
        recovered = False
        for _ in range(20):                       # ~20s de janela de recuperação
            await asyncio.sleep(1)
            try:
                if (await client.get("/api/v1/clientes/", headers=read_hdr)).status_code == 200:
                    recovered = True
                    break
            except Exception:
                pass
        print(f"  Recuperação: {'OK ✓' if recovered else 'FALHOU ✗'}")

    ok = process_alive and degrades_5xx and recovered
    print(f"  {'PASS' if ok else 'FAIL'} — degradação graciosa {'validada' if ok else 'falhou'}")
    return ok


# ══════════════════════════════════════════════════════════════════
# Suite runner
# ══════════════════════════════════════════════════════════════════

async def main():
    print("╔══════════════════════════════════════════════════════╗")
    print("║   SUITE DE FITNESS FUNCTIONS — ERP Estoque (Go)      ║")
    print("╚══════════════════════════════════════════════════════╝")

    results = [
        ("FF1 Boundary Isolation",   run_ff1()),
        ("FF2 Contract Tests",       await run_ff2()),
        ("FF3 Latência p99",         await run_ff3()),
        ("FF4 Chaos / Degradação",   await run_ff4()),
    ]

    print("\n╔══════════════════════════════════════════════════════╗")
    print("║                   RESULTADO FINAL                   ║")
    print("╚══════════════════════════════════════════════════════╝")
    for name, ok in results:
        print(f"  {'✓ PASS' if ok else '✗ FAIL'}  {name}")

    failed = [n for n, ok in results if not ok]
    print()
    if failed:
        print(f"SUITE REPROVADA — {len(failed)} fitness function(s) falharam")
        sys.exit(1)
    print("SUITE APROVADA — Arquitetura em conformidade")


if __name__ == "__main__":
    asyncio.run(main())
