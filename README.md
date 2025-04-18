# dcgm-metrics-api

A Helm chart for deploying the DCGM Metrics API service, which provides GPU metrics from NVIDIA DCGM (Data Center GPU Manager) through a RESTful API.

## Features

- Exposes GPU utilization metrics via HTTP endpoints
- Configurable Prometheus integration
- Health and readiness probes
- Resource limits and requests configuration
- Flexible deployment options (ClusterIP, LoadBalancer)

## Installation

1. Add the Helm repository:
```bash
helm repo add dcgm-metrics-api https://v01d42.github.io/dcgm-metrics-api/
helm repo update
```

2. Install the chart:
```bash
helm install dcgm-metrics-api -n <namespace> dcgm-metrics-api/dcgm-metrics-api -f values.yaml
```

## Configuration

Key configuration options in `values.yaml`:
- `replicaCount`: Number of replicas
- `image.repository`: Container image repository
- `env.PROMETHEUS_URL`: Prometheus server URL
- `env.METRIC_NAMES`: List of DCGM metrics to collect
- `service.type`: Service type (ClusterIP, LoadBalancer)
- `resources`: CPU and memory limits/requests

For detailed configuration options, see `values.yaml`.
