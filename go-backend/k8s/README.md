# Kubernetes Manifests

This directory contains Kubernetes templates for the backend service.

These manifests are starting points, not production-ready deployment proof. Review secrets, domains, resource limits, probes, storage, and network policy before using them outside local or staging experiments.

## Files

- `deployment.yaml` - backend deployment, service, config, and secret template
- `hpa.yaml` - horizontal pod autoscaler template
- `ingress.yaml` - ingress template
- `monitoring/` - Prometheus, Grafana, and Alertmanager templates

## Apply Order

```powershell
cd go-backend
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/ingress.yaml
```

## Required Review

- Replace placeholder secrets with environment-managed secrets.
- Set real image names and tags.
- Confirm `/health`, `/ready`, and `/liveness` probes match the runtime service.
- Configure real hostnames and TLS certificates.
- Review CPU and memory requests/limits against actual load.
- Confirm PostgreSQL, Redis, and object storage are managed and backed up.

## Related Docs

- Backend deployment notes: `../DEPLOYMENT.md`
- Monitoring templates: `monitoring/README.md`
