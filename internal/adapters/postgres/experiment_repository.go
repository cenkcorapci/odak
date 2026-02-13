package postgres

import (
	"database/sql"
	"fmt"

	"github.com/cenkcorapci/odak/internal/domain"
)

// ExperimentRepository implements domain.ExperimentRepository using PostgreSQL
type ExperimentRepository struct {
	db *DB
}

// NewExperimentRepository creates a new PostgreSQL experiment repository
func NewExperimentRepository(db *DB) *ExperimentRepository {
	return &ExperimentRepository{db: db}
}

// Create creates a new experiment
func (r *ExperimentRepository) Create(exp *domain.Experiment) error {
	variantsJSON, err := jsonToString(exp.Variants)
	if err != nil {
		return fmt.Errorf("failed to marshal variants: %w", err)
	}

	query := `
		INSERT INTO experiments (id, name, description, status, variants, traffic_pct, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.conn.Exec(query,
		exp.ID, exp.Name, exp.Description, exp.Status, variantsJSON,
		exp.TrafficPct, exp.CreatedAt, exp.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create experiment: %w", err)
	}

	return nil
}

// GetByID retrieves an experiment by ID
func (r *ExperimentRepository) GetByID(id string) (*domain.Experiment, error) {
	query := `
		SELECT id, name, description, status, variants, traffic_pct, created_at, updated_at
		FROM experiments WHERE id = $1
	`

	exp := &domain.Experiment{}
	var variantsJSON string

	err := r.db.conn.QueryRow(query, id).Scan(
		&exp.ID, &exp.Name, &exp.Description, &exp.Status, &variantsJSON,
		&exp.TrafficPct, &exp.CreatedAt, &exp.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("experiment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	if err := stringToJSON(variantsJSON, &exp.Variants); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variants: %w", err)
	}

	return exp, nil
}

// List retrieves all experiments
func (r *ExperimentRepository) List() ([]*domain.Experiment, error) {
	query := `
		SELECT id, name, description, status, variants, traffic_pct, created_at, updated_at
		FROM experiments ORDER BY created_at DESC
	`

	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list experiments: %w", err)
	}
	defer rows.Close()

	var experiments []*domain.Experiment
	for rows.Next() {
		exp := &domain.Experiment{}
		var variantsJSON string

		err := rows.Scan(&exp.ID, &exp.Name, &exp.Description, &exp.Status, &variantsJSON,
			&exp.TrafficPct, &exp.CreatedAt, &exp.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan experiment: %w", err)
		}

		if err := stringToJSON(variantsJSON, &exp.Variants); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variants: %w", err)
		}

		experiments = append(experiments, exp)
	}

	return experiments, nil
}

// Update updates an existing experiment
func (r *ExperimentRepository) Update(exp *domain.Experiment) error {
	variantsJSON, err := jsonToString(exp.Variants)
	if err != nil {
		return fmt.Errorf("failed to marshal variants: %w", err)
	}

	query := `
		UPDATE experiments
		SET name = $2, description = $3, status = $4, variants = $5, traffic_pct = $6, updated_at = $7
		WHERE id = $1
	`

	result, err := r.db.conn.Exec(query,
		exp.ID, exp.Name, exp.Description, exp.Status, variantsJSON,
		exp.TrafficPct, exp.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update experiment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("experiment not found")
	}

	return nil
}

// Delete deletes an experiment
func (r *ExperimentRepository) Delete(id string) error {
	query := `DELETE FROM experiments WHERE id = $1`

	result, err := r.db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete experiment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("experiment not found")
	}

	return nil
}
