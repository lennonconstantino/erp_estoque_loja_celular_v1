# Grafana Alloy que raspa o /metrics do backend e faz remote_write pro Grafana
# Cloud. Deploy no Railway: novo serviço com Root Directory = observability/alloy
# (usa este Dockerfile). As variáveis GRAFANA_CLOUD_*/ERP_METRICS_* vão nas
# Variables do serviço — ver docs/setup/observability-activation.md.
#
# Ajuste a tag conforme a release atual do Alloy (o repo pina versões de infra).
FROM grafana/alloy:v1.5.1

COPY config.alloy /etc/alloy/config.alloy

# Shell form para expandir $PORT (injetado pelo Railway; fallback 12345 local).
# --server.http.listen-addr expõe a UI/health (/-/ready) para o healthcheck.
ENTRYPOINT []
CMD /bin/alloy run \
	--server.http.listen-addr=0.0.0.0:${PORT:-12345} \
	--storage.path=/var/lib/alloy/data \
	/etc/alloy/config.alloy
