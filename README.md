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
helm repo add dcgm-metrics-api https://github.com/V01d42/dcgm-metrics-api
```

2. Install the chart:
```bash
helm install dcgm-metrics-api dcgm-metrics-api/dcgm-metrics-api \
  --set env.PROMETHEUS_URL="http://your-prometheus:9090" \
  --set env.METRIC_NAMES="DCGM_FI_DEV_GPU_UTIL"
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
