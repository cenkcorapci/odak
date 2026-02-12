# Odak - Product Matching A/B Testing Framework

A simplified A/B testing tool for ML models focused on product matching, using PostgreSQL, Grafana, and MLflow.

## Overview

Odak is an open-source A/B testing framework specifically designed for evaluating machine learning models in production environments. It simplifies the original design by using only three core technologies:

- **PostgreSQL**: Data storage for experiments, assignments, and metrics
- **Grafana**: Real-time dashboards and visualization
- **MLflow**: ML model tracking and management

## Architecture

The system follows hexagonal architecture with clear separation of concerns:

```
├── cmd/server          # HTTP server entry point
├── internal/
│   ├── domain          # Core business logic (bucketing, experiments)
│   ├── application     # Use cases and services
│   └── adapters        # External integrations (PostgreSQL, MLflow)
└── configs/            # Configuration files for Grafana
```

## Key Features

1. **Deterministic Bucketing**: Uses MurmurHash3 for consistent user assignment
2. **Experiment Management**: Create, update, and monitor A/B tests
3. **Metric Tracking**: Real-time metric collection and aggregation
4. **MLflow Integration**: Track and deploy ML models
5. **Grafana Dashboards**: Pre-configured dashboards for experiment monitoring

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Running with Docker Compose

1. Clone the repository:
```bash
git clone https://github.com/cenkcorapci/odak.git
cd odak
```

2. Start all services:
```bash
docker compose up -d
```

This will start:
- PostgreSQL (port 5432)
- MLflow (port 5000)
- Grafana (port 3000)
- Odak API (port 8080)

3. Access the services:
- Odak API: http://localhost:8080
- MLflow UI: http://localhost:5000
- Grafana: http://localhost:3000 (admin/admin)

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Start PostgreSQL and MLflow:
```bash
docker compose up -d postgres mlflow
```

3. Run the server:
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/odak?sslmode=disable"
export MLFLOW_URL="http://localhost:5000"
export BUCKETING_SALT="your-secret-salt"
go run cmd/server/main.go
```

## API Usage

### Create an Experiment

```bash
curl -X POST http://localhost:8080/api/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Product Matching Test",
    "description": "Testing new embedding model",
    "status": "active",
    "traffic_pct": 100,
    "variants": [
      {
        "name": "Control",
        "description": "Current model",
        "traffic_pct": 5000,
        "model_uri": "models:/product-matching/1"
      },
      {
        "name": "Variant A",
        "description": "New embedding model",
        "traffic_pct": 5000,
        "model_uri": "models:/product-matching/2"
      }
    ]
  }'
```

### Assign a Variant

```bash
curl -X POST http://localhost:8080/api/assign \
  -H "Content-Type: application/json" \
  -d '{
    "experiment_id": "experiment-id-here",
    "entity_id": "user-123"
  }'
```

### Track a Metric

```bash
curl -X POST http://localhost:8080/api/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "experiment_id": "experiment-id-here",
    "variant_id": "variant-id-here",
    "entity_id": "user-123",
    "metric_name": "conversion_rate",
    "metric_value": 1.0
  }'
```

### List Experiments

```bash
curl http://localhost:8080/api/experiments
```

### Get Metric Aggregation

```bash
curl http://localhost:8080/api/metrics/{experimentId}/{variantId}/{metricName}
```

## Database Schema

### Experiments Table
- `id`: Unique identifier
- `name`: Experiment name
- `description`: Experiment description
- `status`: Current status (draft, active, paused, completed)
- `variants`: JSON array of variants
- `traffic_pct`: Percentage of traffic to include
- `created_at`, `updated_at`: Timestamps

### Assignments Table
- `experiment_id`: Reference to experiment
- `entity_id`: User or product ID
- `variant_id`: Assigned variant
- `assigned_at`: Assignment timestamp

### Metric Events Table
- `id`: Unique identifier
- `experiment_id`: Reference to experiment
- `variant_id`: Reference to variant
- `entity_id`: User or product ID
- `metric_name`: Metric name
- `metric_value`: Metric value
- `timestamp`: Event timestamp
- `properties`: JSON metadata

## Grafana Dashboards

The system includes a pre-configured Grafana dashboard with:

1. **Experiments Overview**: List of all experiments and their status
2. **Active Experiments Gauge**: Count of currently active experiments
3. **Variant Assignments**: Bar chart showing distribution across variants
4. **Conversion Rate Timeline**: Time-series view of conversion rates
5. **Metrics Summary**: Table with aggregated metrics per variant

Access the dashboard at: http://localhost:3000/d/odak-main

## MLflow Integration

MLflow is used for:
- Model versioning and registry
- Experiment tracking
- Model deployment URIs

Each variant can reference a specific MLflow model version using the `model_uri` field (e.g., `models:/product-matching/2`).

## Configuration

Environment variables:

- `DATABASE_URL`: PostgreSQL connection string (default: postgres://postgres:postgres@localhost:5432/odak?sslmode=disable)
- `MLFLOW_URL`: MLflow server URL (default: http://localhost:5000)
- `BUCKETING_SALT`: Secret salt for bucketing algorithm (default: odak-secret-salt)
- `PORT`: HTTP server port (default: 8080)

## Traffic Distribution

Variants use a 0-9999 bucketing system where traffic percentages are specified in basis points:
- 5000 = 50% traffic
- 2500 = 25% traffic
- 7500 = 75% traffic

The sum of all variant traffic percentages must equal 10000.

## Health Checks

Check service health:
```bash
curl http://localhost:8080/health
```

## Stopping Services

```bash
docker compose down
```

To also remove volumes:
```bash
docker compose down -v
```

## License

See LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
