package domain

import (
	"time"
)

// Experiment represents an A/B test experiment
type Experiment struct {
	ID          string
	Name        string
	Description string
	Status      string // "draft", "active", "paused", "completed"
	Variants    []Variant
	TrafficPct  int       // Percentage of traffic to include (0-100)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Variant represents a variant in an experiment
type Variant struct {
	ID          string
	Name        string
	Description string
	TrafficPct  int    // Percentage of experiment traffic (sum should be 100)
	ModelURI    string // MLflow model URI
}

// Assignment represents a user's assignment to a variant
type Assignment struct {
	ExperimentID string
	EntityID     string // User ID or Product ID
	VariantID    string
	AssignedAt   time.Time
}

// MetricEvent represents a tracked event for metrics
type MetricEvent struct {
	ID           string
	ExperimentID string
	VariantID    string
	EntityID     string
	MetricName   string
	MetricValue  float64
	Timestamp    time.Time
	Properties   map[string]interface{}
}
