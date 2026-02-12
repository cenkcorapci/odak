package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
)

// DB wraps the database connection
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection
func NewDB(connStr string) (*DB, error) {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// InitSchema initializes the database schema
func (db *DB) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS experiments (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		status VARCHAR(50) NOT NULL,
		variants JSONB NOT NULL,
		traffic_pct INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_experiments_status ON experiments(status);

	CREATE TABLE IF NOT EXISTS assignments (
		experiment_id VARCHAR(255) NOT NULL,
		entity_id VARCHAR(255) NOT NULL,
		variant_id VARCHAR(255) NOT NULL,
		assigned_at TIMESTAMP NOT NULL,
		PRIMARY KEY (experiment_id, entity_id)
	);

	CREATE INDEX IF NOT EXISTS idx_assignments_variant ON assignments(variant_id);
	CREATE INDEX IF NOT EXISTS idx_assignments_experiment ON assignments(experiment_id);

	CREATE TABLE IF NOT EXISTS metric_events (
		id VARCHAR(255) PRIMARY KEY,
		experiment_id VARCHAR(255) NOT NULL,
		variant_id VARCHAR(255) NOT NULL,
		entity_id VARCHAR(255) NOT NULL,
		metric_name VARCHAR(255) NOT NULL,
		metric_value DOUBLE PRECISION NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		properties JSONB
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_experiment ON metric_events(experiment_id);
	CREATE INDEX IF NOT EXISTS idx_metrics_variant ON metric_events(variant_id);
	CREATE INDEX IF NOT EXISTS idx_metrics_name ON metric_events(metric_name);
	CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metric_events(timestamp);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// jsonToString converts a value to JSON string
func jsonToString(v interface{}) (string, error) {
	if v == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// stringToJSON converts a JSON string to a value
func stringToJSON(s string, v interface{}) error {
	if s == "" || s == "{}" {
		return nil
	}
	return json.Unmarshal([]byte(s), v)
}
