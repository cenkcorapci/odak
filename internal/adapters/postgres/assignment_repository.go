package postgres

import (
	"database/sql"
	"fmt"

	"github.com/cenkcorapci/odak/internal/domain"
)

// AssignmentRepository implements domain.AssignmentRepository using PostgreSQL
type AssignmentRepository struct {
	db *DB
}

// NewAssignmentRepository creates a new PostgreSQL assignment repository
func NewAssignmentRepository(db *DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// Create creates a new assignment
func (r *AssignmentRepository) Create(assignment *domain.Assignment) error {
	query := `
		INSERT INTO assignments (experiment_id, entity_id, variant_id, assigned_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (experiment_id, entity_id) DO NOTHING
	`

	_, err := r.db.conn.Exec(query,
		assignment.ExperimentID, assignment.EntityID, assignment.VariantID, assignment.AssignedAt)
	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	return nil
}

// GetByExperimentAndEntity retrieves an assignment by experiment and entity ID
func (r *AssignmentRepository) GetByExperimentAndEntity(experimentID, entityID string) (*domain.Assignment, error) {
	query := `
		SELECT experiment_id, entity_id, variant_id, assigned_at
		FROM assignments WHERE experiment_id = $1 AND entity_id = $2
	`

	assignment := &domain.Assignment{}
	err := r.db.conn.QueryRow(query, experimentID, entityID).Scan(
		&assignment.ExperimentID, &assignment.EntityID, &assignment.VariantID, &assignment.AssignedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	return assignment, nil
}

// ListByExperiment retrieves all assignments for an experiment
func (r *AssignmentRepository) ListByExperiment(experimentID string) ([]*domain.Assignment, error) {
	query := `
		SELECT experiment_id, entity_id, variant_id, assigned_at
		FROM assignments WHERE experiment_id = $1
	`

	rows, err := r.db.conn.Query(query, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.Assignment
	for rows.Next() {
		assignment := &domain.Assignment{}
		err := rows.Scan(&assignment.ExperimentID, &assignment.EntityID,
			&assignment.VariantID, &assignment.AssignedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}
