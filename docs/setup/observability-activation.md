# Ativar a observabilidade em produção (Railway + Grafana Cloud)

Passo a passo para ligar o **push gerenciado** de métricas: proteger o `/metrics`,
subir o **Grafana Alloy** no Railway (scrape → `remote_write`) e carregar os
**alertas** no Grafana Cloud. Modelo e trade-offs em
[docs/architecture/observability.md](../architecture/observability.md).

> **Estes passos exigem suas credenciais** (Railway + Grafana Cloud) e alteram
> produção — execute-os você. Os artefatos já estão no repo:
> [`observability/alloy/`](../../observability/alloy) (Dockerfile + config + railway.json)
> e [`observability/alerts/erp-alerts.yml`](../../observability/alerts/erp-alerts.yml).
>
> **Pré-requisitos:** backend no ar no Railway; conta no **Grafana Cloud** (o
> _free tier_ atende — ~10k séries, 14 dias de retenção).

---

## Modelo mental (leia antes)

Dois esclarecimentos que evitam a maior parte dos tropeços:

- **Repo ≠ variáveis.** O repo tem a **receita** (Dockerfile, `config.alloy`,
  `railway.json`, alertas). Os **valores** das variáveis vivem nas **Variables do
  serviço no Railway** (dashboard), setados **à mão uma vez** — merge de código
  **não** carrega nem sobrescreve variáveis. Um `.env` no repo **não é lido** pelo
  Railway; os `*.env.production.example` são só a checklist versionada.
- **Dois tokens, duas pernas.** Não confunda:
  - `METRICS_TOKEN` (backend) **=** `ERP_METRICS_TOKEN` (Alloy) → autentica o
    **scrape** do `/metrics`. É um segredo **nosso**.
  - `GRAFANA_CLOUD_TOKEN` (Alloy) → autentica o **remote_write** pro Grafana Cloud.
    É da **Grafana Cloud** (Access Policy).
  - As pernas são independentes: **Alloy→backend** (scrape, define o _valor_ de
    `up`) e **Alloy→Grafana Cloud** (remote_write, _entrega_ os dados). Consertar
    uma costuma **revelar** o problema da outra.

---

## 1. Proteger o `/metrics` com `METRICS_TOKEN`

1. Gere um token forte (rode **você**, não compartilhe):

   ```bash
   openssl rand -hex 32
   ```

2. Railway → serviço **erp-estoque-backend** → **Variables** → adicione
   `METRICS_TOKEN=<valor>` → **Deploy**.

3. Verifique (sem token = 404; com token = 200):

   ```bash
   BASE=https://erp-estoque-backend-production.up.railway.app
   curl -s -o /dev/null -w "sem token: %{http_code}\n" $BASE/metrics
   curl -s -o /dev/null -w "com token: %{http_code}\n" -H "Authorization: Bearer <valor>" $BASE/metrics
   ```

> Enquanto `METRICS_TOKEN` estiver vazio em produção, o `/metrics` responde **404**
> (fail-safe de [`ProtectMetrics`](../../backend/internal/platform/httpserver/metrics.go)).

## 2. Criar o stack e o token no Grafana Cloud

1. grafana.com → **Grafana Cloud** → crie/abra um stack.
2. **Connections → Prometheus (Hosted)** e anote:
   - **Remote Write Endpoint** (URL, termina em `/api/prom/push`) → `GRAFANA_CLOUD_URL`
   - **Username / Instance ID** (numérico) → `GRAFANA_CLOUD_USER`
3. **Access Policies** → crie uma policy com os escopos abaixo (realm apontando
   para o stack) → gere um **token** (`glc_...`) → `GRAFANA_CLOUD_TOKEN`:
   - `metrics:write` — usado pelo **Alloy** (remote_write, passo 3).
   - `rules:read` + `rules:write` — usados pelo **mimirtool** (ruler, passo 4).

   > O ruler exige `rules:*`; um token só com `metrics:*` faz o mimirtool falhar
   > com `401 "authentication error: invalid scope requested"`. Dá para usar uma
   > policy única com os três escopos, ou separar Alloy (metrics) do mimirtool (rules).

## 3. Subir o serviço Alloy no Railway

1. **New Service → Deploy from Repo** → **Root Directory** = `observability/alloy`
   (usa o [Dockerfile](../../observability/alloy/Dockerfile) e o
   [railway.json](../../observability/alloy/railway.json) do diretório).
2. **Variables** do serviço Alloy:

   | Variável | Valor |
   |----------|-------|
   | `ERP_METRICS_ADDR` | `${{erp-estoque-backend.RAILWAY_PRIVATE_DOMAIN}}:${{erp-estoque-backend.APP_PORT}}` |
   | `ERP_METRICS_TOKEN` | o **mesmo** `METRICS_TOKEN` do passo 1 |
   | `GRAFANA_CLOUD_URL` | endpoint remote_write |
   | `GRAFANA_CLOUD_USER` | instance ID |
   | `GRAFANA_CLOUD_TOKEN` | token da access policy (com `metrics:write`) |

   > Checklist versionada dessas variáveis: [`observability/alloy/.env.production.example`](../../observability/alloy/.env.production.example).

3. **Deploy**. Nos logs do Alloy, confirme scrape do alvo e `remote_write` **sem 401/403**.

> **A porta do `ERP_METRICS_ADDR`:** use `${{...APP_PORT}}`, **não** `${{...PORT}}`.
> O backend lê `PORT → APP_PORT → 8080`; como não há `PORT`, referenciar `PORT`
> resolve **vazio** → `host:` (porta em branco) → scrape falha → `up=0` → dispara
> `ERPBackendDown`. Confira a porta real no deploy log do backend: `API ouvindo em :XXXX`.
> Rede privada (`*.railway.internal`) é interna → `scheme = http`.

## 4. Carregar os alertas no Grafana Cloud

Os alertas vivem versionados em `observability/alerts/erp-alerts.yml`. Suba-os para
o ruler (Mimir) com o [`mimirtool`](https://grafana.com/docs/mimir/latest/manage/tools/mimirtool/).

> ⚠️ **`--address` é a BASE do stack, NÃO o endpoint de remote_write.** O mimirtool
> **anexa** `/prometheus/config/v1/rules/...` ao que você passar. Se você usar o
> endpoint de push (`.../api/prom/push`), o caminho final fica
> `.../api/prom/push/prometheus/config/v1/rules` → **404 "requested resource not found"**.
> Use o **host base** (só `https://prometheus-prod-NN-....grafana.net`).

```bash
# 1) valide endereço + credenciais SEM escrever (read-only):
mimirtool rules list \
  --address="https://prometheus-prod-40-prod-sa-east-1.grafana.net" \
  --id="$GRAFANA_CLOUD_USER" --key="$GRAFANA_CLOUD_TOKEN"

# 2) se o list respondeu, carregue as regras:
mimirtool rules load observability/alerts/erp-alerts.yml \
  --address="https://prometheus-prod-40-prod-sa-east-1.grafana.net" \
  --id="$GRAFANA_CLOUD_USER" \
  --key="$GRAFANA_CLOUD_TOKEN"
```

Ajuste o host para o do seu stack (Connections → Prometheus → detalhes). Notas:
- `$GRAFANA_CLOUD_USER`/`$GRAFANA_CLOUD_TOKEN` precisam estar exportados **no shell
  local** que roda o mimirtool (não bastam estar nas Variables do Railway).
- Se o `list` ainda der **404** com o host base, seu stack expõe o ruler sob
  `/api/prom` → use `--address=".../api/prom"`. **401/403** = credencial/escopo do token.

Depois, em **Grafana Alerting → Contact points**, crie um destino (e-mail/Slack) e
uma **notification policy** para as severidades `critical`/`warning`.

## 5. Verificar ponta a ponta

- **Grafana Cloud → Explore** → `up{job="erp-api"}` retorna **1**.
- `http_server_request_duration_seconds_count` tem dados (gere tráfego no ERP).
- **Alerting → Rules**: as 4 regras `ERP*` aparecem em estado **Normal**.

## Troubleshooting

Estes são os erros que enfrentamos na primeira ativação — sintoma → causa → correção.
Vale reler o **Modelo mental** acima (dois tokens, duas pernas) antes de debugar.

| Sintoma | Onde | Causa | Correção |
|---|---|---|---|
| `404 requested resource not found` | mimirtool | `--address` era o endpoint de **push** (`.../api/prom/push`); o mimirtool anexa `/prometheus/config/v1/rules` | use o **host base** do stack (sem `/api/prom/push`) |
| `401 authentication error: invalid scope requested` | mimirtool | token **sem** `rules:read`/`rules:write` (só tinha `metrics:*`) | adicione os escopos `rules:*` na Access Policy → gere token novo |
| `401 authentication error: invalid token` no `remote_write` | logs do Alloy | `GRAFANA_CLOUD_TOKEN` errado/revogado ou sem `metrics:write`; **ou** `GRAFANA_CLOUD_USER` ≠ instance ID | atualize o token (com `metrics:write`) nas Variables do Alloy → redeploy |
| `up=0` / `ERPBackendDown` firing | Grafana Cloud | scrape falhando: (a) `ERP_METRICS_ADDR` com **porta vazia** (`${{...PORT}}` em vez de `APP_PORT`); (b) `ERP_METRICS_TOKEN` ≠ `METRICS_TOKEN`; (c) nome do serviço backend errado no endereço | corrija endereço/token; confira a porta no deploy log (`API ouvindo em :XXXX`) |
| `ERPBackendUnreachable` firing (`absent`) | Grafana Cloud | **nada** chega: Alloy não deployado, ou remote_write quebrado | suba/conserte o Alloy (ver `invalid token` acima) |
| `/metrics` → **404** em produção | curl no backend | `METRICS_TOKEN` não setado (fail-safe fecha o endpoint) | setar `METRICS_TOKEN` no backend → redeploy |
| `/metrics` → **401** com header correto | curl no backend | valor do header ≠ `METRICS_TOKEN` do backend | use o valor exato; alinhe `ERP_METRICS_TOKEN` do Alloy |

**Lendo o alerta `ERPBackendUnreachable` (é `absent()` — gráfico invertido):** linha
em **1** = série **ausente** (ruim, firing); a linha **sumir** = série presente
(bom). Não é "morreu" — sumir é saúde. Confirme sempre **no positivo**:
`up{job="erp-api"}` = 1 no Explore.

## Manual vs. reproduzível (aceite isto)

Nem tudo sobe sozinho do repo — e tudo bem, desde que esteja claro o que é o quê:

- **Reproduzível pelo repo** (merge → redeploy): `Dockerfile`, `config.alloy`,
  `railway.json`, `erp-alerts.yml`, migrations. A **lista** de variáveis está nos
  `*.env.production.example`.
- **Manual, sempre** (não vem do repo): os **valores** dos segredos, digitados nas
  **Variables do serviço no Railway**; e o setup no **Grafana Cloud** (stack, Access
  Policy, token, contact point). Isso é inerente — segredo não mora no git.
- **Antes de recriar serviços do zero**, faça **backup das Variables** (`railway
  variables` na CLI, ou o Raw Editor no dashboard) num cofre de segredos. Sem isso,
  você regenera cada segredo e re-liga o Grafana Cloud na mão.

## Desligar / rollback

- **Parar a coleta:** pare/derrube o serviço **Alloy** (o `/metrics` continua
  protegido). Não remova o `METRICS_TOKEN` do backend — em produção isso apenas
  fecha o `/metrics`, não é "abrir".
- **Remover alertas:** `mimirtool rules delete <namespace> <group>` (ou apague o
  namespace `erp-api`).
