package domain

import (
	"time"
)

// Experiment represents an A/B test experiment
type Experiment struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "draft", "active", "paused", "completed"
	Variants    []Variant `json:"variants"`
	TrafficPct  int       `json:"traffic_pct"` // Percentage of traffic to include (0-100)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Variant represents a variant in an experiment
type Variant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TrafficPct  int    `json:"traffic_pct"` // Percentage in basis points (0-10000, sum should be 10000)
	ModelURI    string `json:"model_uri"`   // MLflow model URI
}

// Assignment represents a user's assignment to a variant
type Assignment struct {
	ExperimentID string    `json:"experiment_id"`
	EntityID     string    `json:"entity_id"` // User ID or Product ID
	VariantID    string    `json:"variant_id"`
	AssignedAt   time.Time `json:"assigned_at"`
}

// MetricEvent represents a tracked event for metrics
type MetricEvent struct {
	ID           string                 `json:"id"`
	ExperimentID string                 `json:"experiment_id"`
	VariantID    string                 `json:"variant_id"`
	EntityID     string                 `json:"entity_id"`
	MetricName   string                 `json:"metric_name"`
	MetricValue  float64                `json:"metric_value"`
	Timestamp    time.Time              `json:"timestamp"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}
