#!/bin/bash

# Example script to create an experiment and test the API

BASE_URL="http://localhost:8080"

echo "Creating experiment..."
EXPERIMENT=$(curl -s -X POST "$BASE_URL/api/experiments" \
  -H "Content-Type: application/json" \
  -d @examples/experiment.json)

EXPERIMENT_ID=$(echo $EXPERIMENT | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

echo "Experiment created with ID: $EXPERIMENT_ID"
echo ""

echo "Assigning user-001 to a variant..."
ASSIGNMENT=$(curl -s -X POST "$BASE_URL/api/assign" \
  -H "Content-Type: application/json" \
  -d "{\"experiment_id\":\"$EXPERIMENT_ID\",\"entity_id\":\"user-001\"}")

VARIANT_ID=$(echo $ASSIGNMENT | grep -o '"variant_id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "User assigned to variant: $VARIANT_ID"
echo ""

echo "Tracking conversion metric..."
curl -s -X POST "$BASE_URL/api/metrics" \
  -H "Content-Type: application/json" \
  -d "{\"experiment_id\":\"$EXPERIMENT_ID\",\"variant_id\":\"$VARIANT_ID\",\"entity_id\":\"user-001\",\"metric_name\":\"conversion\",\"metric_value\":1.0}"

echo ""
echo "Tracking match quality metric..."
curl -s -X POST "$BASE_URL/api/metrics" \
  -H "Content-Type: application/json" \
  -d "{\"experiment_id\":\"$EXPERIMENT_ID\",\"variant_id\":\"$VARIANT_ID\",\"entity_id\":\"user-001\",\"metric_name\":\"match_quality\",\"metric_value\":0.85}"

echo ""
echo ""
echo "Listing all experiments..."
curl -s "$BASE_URL/api/experiments" | json_pp || cat

echo ""
echo "Getting metric aggregation..."
curl -s "$BASE_URL/api/metrics/$EXPERIMENT_ID/$VARIANT_ID/conversion" | json_pp || cat
