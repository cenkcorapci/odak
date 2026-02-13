package postgres

import (
	"database/sql"
	"fmt"

	"github.com/cenkcorapci/odak/internal/domain"
)

// MetricRepository implements domain.MetricRepository using PostgreSQL
type MetricRepository struct {
	db *DB
}

// NewMetricRepository creates a new PostgreSQL metric repository
func NewMetricRepository(db *DB) *MetricRepository {
	return &MetricRepository{db: db}
}

// Create creates a new metric event
func (r *MetricRepository) Create(event *domain.MetricEvent) error {
	propsJSON, err := jsonToString(event.Properties)
	if err != nil {
		return fmt.Errorf("failed to marshal properties: %w", err)
	}

	query := `
		INSERT INTO metric_events (id, experiment_id, variant_id, entity_id, metric_name, metric_value, timestamp, properties)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.conn.Exec(query,
		event.ID, event.ExperimentID, event.VariantID, event.EntityID,
		event.MetricName, event.MetricValue, event.Timestamp, propsJSON)
	if err != nil {
		return fmt.Errorf("failed to create metric event: %w", err)
	}

	return nil
}

// GetByExperiment retrieves all metric events for an experiment
func (r *MetricRepository) GetByExperiment(experimentID string) ([]*domain.MetricEvent, error) {
	query := `
		SELECT id, experiment_id, variant_id, entity_id, metric_name, metric_value, timestamp, properties
		FROM metric_events WHERE experiment_id = $1 ORDER BY timestamp DESC
	`

	rows, err := r.db.conn.Query(query, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric events: %w", err)
	}
	defer rows.Close()

	var events []*domain.MetricEvent
	for rows.Next() {
		event := &domain.MetricEvent{}
		var propsJSON string

		err := rows.Scan(&event.ID, &event.ExperimentID, &event.VariantID, &event.EntityID,
			&event.MetricName, &event.MetricValue, &event.Timestamp, &propsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric event: %w", err)
		}

		if err := stringToJSON(propsJSON, &event.Properties); err != nil {
			return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
		}

		events = append(events, event)
	}

	return events, nil
}

// GetAggregated retrieves aggregated metrics for a variant
func (r *MetricRepository) GetAggregated(experimentID, variantID, metricName string) (float64, int, error) {
	query := `
		SELECT AVG(metric_value) as avg_value, COUNT(*) as count
		FROM metric_events
		WHERE experiment_id = $1 AND variant_id = $2 AND metric_name = $3
	`

	var avgValue sql.NullFloat64
	var count int

	err := r.db.conn.QueryRow(query, experimentID, variantID, metricName).Scan(&avgValue, &count)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get aggregated metrics: %w", err)
	}

	if !avgValue.Valid {
		return 0, 0, nil
	}

	return avgValue.Float64, count, nil
}
