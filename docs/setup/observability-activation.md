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
3. **Access Policies** → crie uma policy com escopo `metrics:write` (e `metrics:read`
   para o mimirtool do passo 4) → gere um **token** → `GRAFANA_CLOUD_TOKEN`.

## 3. Subir o serviço Alloy no Railway

1. **New Service → Deploy from Repo** → **Root Directory** = `observability/alloy`
   (usa o [Dockerfile](../../observability/alloy/Dockerfile) e o
   [railway.json](../../observability/alloy/railway.json) do diretório).
2. **Variables** do serviço Alloy:

   | Variável | Valor |
   |----------|-------|
   | `ERP_METRICS_ADDR` | `${{erp-estoque-backend.RAILWAY_PRIVATE_DOMAIN}}:${{erp-estoque-backend.PORT}}` |
   | `ERP_METRICS_TOKEN` | o **mesmo** `METRICS_TOKEN` do passo 1 |
   | `GRAFANA_CLOUD_URL` | endpoint remote_write |
   | `GRAFANA_CLOUD_USER` | instance ID |
   | `GRAFANA_CLOUD_TOKEN` | token da access policy |

3. **Deploy**. Nos logs do Alloy, confirme scrape do alvo e `remote_write` **sem 401/403**.

> `ERP_METRICS_ADDR` usa **variáveis de referência** do Railway porque o backend
> escuta no `$PORT` injetado (não é fixo). A rede privada (`*.railway.internal`)
> é interna → `scheme = http`.

## 4. Carregar os alertas no Grafana Cloud

Os alertas vivem versionados em `observability/alerts/erp-alerts.yml`. Suba-os para
o ruler (Mimir) com o [`mimirtool`](https://grafana.com/docs/mimir/latest/manage/tools/mimirtool/):

```bash
mimirtool rules load observability/alerts/erp-alerts.yml \
  --address="<GRAFANA_CLOUD_PROM_BASE_URL>" \  # base do stack, sem /api/prom/push
  --id="$GRAFANA_CLOUD_USER" \
  --key="$GRAFANA_CLOUD_TOKEN"

# conferir
mimirtool rules list --address="<GRAFANA_CLOUD_PROM_BASE_URL>" --id="$GRAFANA_CLOUD_USER" --key="$GRAFANA_CLOUD_TOKEN"
```

Depois, em **Grafana Alerting → Contact points**, crie um destino (e-mail/Slack) e
uma **notification policy** para as severidades `critical`/`warning`.

## 5. Verificar ponta a ponta

- **Grafana Cloud → Explore** → `up{job="erp-api"}` retorna **1**.
- `http_server_request_duration_seconds_count` tem dados (gere tráfego no ERP).
- **Alerting → Rules**: as 4 regras `ERP*` aparecem em estado **Normal**.

## Desligar / rollback

- **Parar a coleta:** pare/derrube o serviço **Alloy** (o `/metrics` continua
  protegido). Não remova o `METRICS_TOKEN` do backend — em produção isso apenas
  fecha o `/metrics`, não é "abrir".
- **Remover alertas:** `mimirtool rules delete <namespace> <group>` (ou apague o
  namespace `erp-api`).
