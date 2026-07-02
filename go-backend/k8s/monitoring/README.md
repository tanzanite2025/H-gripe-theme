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

## Backend Metrics

The backend exposes metrics at:

```text
/metrics
```
