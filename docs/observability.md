# Observability

## Commands

```bash
make obs-up       # Start Prometheus + Loki + Promtail + Grafana
make obs-down     # Stop observability stack
make obs-status   # Check container status
```

## Links

| Service    | URL                            | Notes                    |
|------------|--------------------------------|--------------------------|
| Grafana    | http://localhost:3001          | admin/admin, anonymous OK |
| Prometheus | http://localhost:9091          | Targets, query UI        |
| Loki       | http://localhost:3100          | Log aggregation backend  |

## Grafana

### Dashboard

Pre-provisioned **"Widia API"** dashboard with 8 panels:

- Request Rate (req/s by method)
- Error Rate (4xx + 5xx %)
- Latency p50 / p95 / p99
- Request Duration Heatmap
- Active Subscriptions by Tier
- Asynq Queue Depth
- Asynq Job Failures
- Entitlement Limits Hit

### Log Exploration

Grafana Explore > Loki datasource:

```logql
{job="widia-api"}                          # all logs
{job="widia-api"} |= "error"              # filter by text
{job="widia-api"} | json | level="error"  # structured filter
```

## Prometheus Targets

| Job           | Endpoint                        |
|---------------|---------------------------------|
| widia-api     | http://localhost:8080/metrics   |
| widia-worker  | http://localhost:9090/metrics   |

Check targets: http://localhost:9091/targets

## Useful PromQL

```promql
# Request rate
sum(rate(http_requests_total[5m])) by (method)

# Error rate
sum(rate(http_requests_total{status=~"4..|5.."}[5m])) / sum(rate(http_requests_total[5m]))

# p95 latency
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# Queue depth
asynq_queue_depth

# Job failures
sum(rate(asynq_job_failures_total[5m])) by (task_type)
```

## Logs

App writes JSON to `logs/app.log` (gitignored). Promtail tails this file and pushes to Loki.

Dev mode also outputs colored human-readable logs to stdout.
