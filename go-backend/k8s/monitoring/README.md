# Monitoring Manifests

This directory contains monitoring templates for Kubernetes deployments.

The files are examples and must be reviewed before use in a real environment.

## Files

- `prometheus-config.yaml` - Prometheus scrape/config template
- `grafana-config.yaml` - Grafana dashboard/config template
- `alertmanager-config.yaml` - Alertmanager routing template

## Apply

```powershell
cd go-backend
kubectl apply -f k8s/monitoring/prometheus-config.yaml
kubectl apply -f k8s/monitoring/grafana-config.yaml
kubectl apply -f k8s/monitoring/alertmanager-config.yaml
```

## Required Review

- Replace default credentials and webhook URLs.
- Confirm scrape targets match deployed service labels and ports.
- Keep alert thresholds environment-specific.
- Restrict dashboard and metrics access.
- Avoid logging or exporting sensitive request data.
- Honeypot timing panels are observation-only; calibrate the real-user distribution before adding a rejection alert or enabling timing enforcement.

## Backend Metrics

The backend exposes metrics at:

```text
/metrics
```

Honeypot timing observation is exposed as `commerce_platform_honeypot_timing_evaluations_total`, `commerce_platform_honeypot_timing_seconds`, and `commerce_platform_honeypot_timing_replays_total`. These signals are for calibration only; do not turn them into blocking alerts without a measured false-positive baseline.
