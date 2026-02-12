# Testing Guide

This document provides step-by-step instructions for testing the Odak A/B testing service.

## Prerequisites

- Docker and Docker Compose installed
- curl or similar HTTP client
- (Optional) PostgreSQL client for direct database queries

## Starting the Services

1. Start all services using Docker Compose:
```bash
docker compose up -d
```

2. Verify all services are running:
```bash
docker compose ps
```

You should see 4 services:
- `odak-postgres-1` (PostgreSQL database)
- `odak-mlflow-1` (MLflow model tracking)
- `odak-grafana-1` (Grafana dashboards)
- `odak-odak-1` (Odak API server)

3. Check service health:
```bash
# Odak API
curl http://localhost:8080/health

# MLflow
curl http://localhost:5000/version

# PostgreSQL (from within container)
docker compose exec postgres psql -U postgres -d odak -c "SELECT version();"
```

## Testing the API

### 1. Create an Experiment

Create a new A/B test experiment with two variants:

```bash
curl -X POST http://localhost:8080/api/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Product Matching Embedding Test",
    "description": "Testing new transformer-based model",
    "status": "active",
    "traffic_pct": 100,
    "variants": [
      {
        "name": "Control - TF-IDF",
        "description": "Current TF-IDF model",
        "traffic_pct": 5000,
        "model_uri": "models:/product-matching/1"
      },
      {
        "name": "Treatment - Transformer",
        "description": "New transformer model",
        "traffic_pct": 5000,
        "model_uri": "models:/product-matching/2"
      }
    ]
  }'
```

Save the returned `id` field for subsequent tests.

### 2. List All Experiments

```bash
curl http://localhost:8080/api/experiments | python3 -m json.tool
```

### 3. Get a Specific Experiment

Replace `EXPERIMENT_ID` with the ID from step 1:

```bash
curl http://localhost:8080/api/experiments/EXPERIMENT_ID | python3 -m json.tool
```

### 4. Assign Users to Variants

Assign multiple users to get variant distribution:

```bash
# User 1
curl -X POST http://localhost:8080/api/assign \
  -H "Content-Type: application/json" \
  -d '{"experiment_id":"EXPERIMENT_ID","entity_id":"user-001"}'

# User 2
curl -X POST http://localhost:8080/api/assign \
  -H "Content-Type: application/json" \
  -d '{"experiment_id":"EXPERIMENT_ID","entity_id":"user-002"}'

# User 3
curl -X POST http://localhost:8080/api/assign \
  -H "Content-Type: application/json" \
  -d '{"experiment_id":"EXPERIMENT_ID","entity_id":"user-003"}'
```

Note: The bucketing algorithm ensures the same user gets the same variant consistently.

### 5. Track Metrics

Track conversion events for assigned users:

```bash
curl -X POST http://localhost:8080/api/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "experiment_id": "EXPERIMENT_ID",
    "variant_id": "VARIANT_ID",
    "entity_id": "user-001",
    "metric_name": "conversion",
    "metric_value": 1.0
  }'
```

Track match quality scores:

```bash
curl -X POST http://localhost:8080/api/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "experiment_id": "EXPERIMENT_ID",
    "variant_id": "VARIANT_ID",
    "entity_id": "user-001",
    "metric_name": "match_quality",
    "metric_value": 0.85
  }'
```

### 6. Get Metric Aggregations

Get average metric values for a variant:

```bash
curl http://localhost:8080/api/metrics/EXPERIMENT_ID/VARIANT_ID/conversion
```

## Testing Database Queries

Connect to PostgreSQL directly:

```bash
docker compose exec postgres psql -U postgres -d odak
```

Sample queries:

```sql
-- View all experiments
SELECT id, name, status, created_at FROM experiments;

-- View variant assignments
SELECT experiment_id, entity_id, variant_id, assigned_at 
FROM assignments 
LIMIT 10;

-- View metric events
SELECT experiment_id, variant_id, metric_name, 
       AVG(metric_value) as avg_value, 
       COUNT(*) as count
FROM metric_events
GROUP BY experiment_id, variant_id, metric_name;

-- Get assignment distribution
SELECT variant_id, COUNT(*) as user_count
FROM assignments
WHERE experiment_id = 'EXPERIMENT_ID'
GROUP BY variant_id;
```

## Testing Grafana Dashboards

1. Open Grafana in your browser:
```
http://localhost:3000
```

2. Login with default credentials:
- Username: `admin`
- Password: `admin`

3. Navigate to the "Odak A/B Testing Dashboard":
- Go to Dashboards → Browse
- Select "Odak A/B Testing Dashboard"

4. Verify the following panels display data:
- Experiments Overview (table)
- Active Experiments (gauge)
- Variant Assignments (bar chart)
- Conversion Rate by Variant (time series)
- Metrics Summary (table)

## Testing MLflow Integration

1. Open MLflow UI in your browser:
```
http://localhost:5000
```

2. You can register models via the API:

```bash
# This is a placeholder - in production you would register actual ML models
curl -X POST http://localhost:5000/api/2.0/mlflow/registered-models/create \
  -H "Content-Type: application/json" \
  -d '{"name": "product-matching"}'
```

3. Access model information via Odak API:

```bash
curl http://localhost:8080/api/mlflow/models/product-matching/1
```

## Automated Test Script

Use the provided test script:

```bash
chmod +x examples/test-api.sh
./examples/test-api.sh
```

This script will:
- Create a test experiment
- Assign a user to a variant
- Track multiple metrics
- Display aggregated results

## Verifying Deterministic Bucketing

Test that the same user always gets the same variant:

```bash
# Assign user-test multiple times
for i in {1..5}; do
  echo "Attempt $i:"
  curl -s -X POST http://localhost:8080/api/assign \
    -H "Content-Type: application/json" \
    -d '{"experiment_id":"EXPERIMENT_ID","entity_id":"user-test"}' \
    | python3 -c "import sys, json; print('Variant:', json.load(sys.stdin)['variant_id'])"
done
```

All attempts should return the same variant ID.

## Load Testing (Optional)

Simple load test to verify performance:

```bash
# Create 100 users and track their assignments
for i in {1..100}; do
  curl -s -X POST http://localhost:8080/api/assign \
    -H "Content-Type: application/json" \
    -d "{\"experiment_id\":\"EXPERIMENT_ID\",\"entity_id\":\"user-$i\"}" > /dev/null
  echo "Assigned user-$i"
done

# Check distribution in database
docker compose exec postgres psql -U postgres -d odak -c \
  "SELECT variant_id, COUNT(*) as count FROM assignments WHERE experiment_id = 'EXPERIMENT_ID' GROUP BY variant_id;"
```

The distribution should be approximately 50/50 for equal traffic percentages.

## Cleanup

Stop all services:

```bash
docker compose down
```

Stop and remove all data:

```bash
docker compose down -v
```

## Troubleshooting

### Service not responding

Check logs:
```bash
docker compose logs odak
docker compose logs postgres
docker compose logs mlflow
docker compose logs grafana
```

### Database connection issues

Verify PostgreSQL is accessible:
```bash
docker compose exec postgres psql -U postgres -c "SELECT 1;"
```

### Port conflicts

If ports 3000, 5000, 5432, or 8080 are in use, modify `docker-compose.yml`:
```yaml
services:
  odak:
    ports:
      - "8081:8080"  # Change 8081 to any free port
```

## Expected Results

After running the tests, you should see:
- ✅ Experiments created and listed successfully
- ✅ Users deterministically assigned to variants
- ✅ Metrics tracked and aggregated correctly
- ✅ Data visible in PostgreSQL
- ✅ Grafana dashboards showing visualizations
- ✅ MLflow API accessible
- ✅ 50/50 traffic distribution for equal variant weights
