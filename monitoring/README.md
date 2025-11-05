# Monitoring Stack

📍 **Location**: `monitoring/`
🔗 **Parent**: [Main README](../README.md)

## Overview
Complete monitoring solution with Prometheus, Grafana, and Alertmanager for tracking blockchain network health and performance.

## Components
- **Prometheus**: Metrics collection (Port 9090)
- **Grafana**: Visualization dashboards (Port 3000)
- **Alertmanager**: Alert routing (Port 9093)
- **Node Exporter**: System metrics (Port 9100)

## Quick Start
```bash
# Start monitoring stack
make monitoring-up

# Access Grafana
open http://localhost:3000
# Login: admin / admin

# Access Prometheus
open http://localhost:9090
```

## Key Metrics Monitored
- Ledger block height
- Transaction throughput
- Peer health status
- Chaincode execution time
- Resource usage (CPU, memory, disk)

## Directory Structure
```
monitoring/
├── prometheus/
│   ├── prometheus.yml    ← Scrape configuration
│   └── alerts.yml        ← Alert rules
├── grafana/
│   └── dashboards/       ← Pre-built dashboards
└── alertmanager/
    └── config.yml        ← Alert routing
```

## Alerts Configured
- Peer/Orderer down
- High error rate (>5%)
- Ledger height divergence
- High resource usage
- Slow chaincode execution

📍 **Navigation**: [Main](../README.md) | [Prometheus →](prometheus/)
