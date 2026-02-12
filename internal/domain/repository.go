package domain

// ExperimentRepository defines the interface for experiment storage
type ExperimentRepository interface {
	Create(experiment *Experiment) error
	GetByID(id string) (*Experiment, error)
	List() ([]*Experiment, error)
	Update(experiment *Experiment) error
	Delete(id string) error
}

// AssignmentRepository defines the interface for assignment storage
type AssignmentRepository interface {
	Create(assignment *Assignment) error
	GetByExperimentAndEntity(experimentID, entityID string) (*Assignment, error)
	ListByExperiment(experimentID string) ([]*Assignment, error)
}

// MetricRepository defines the interface for metric storage
type MetricRepository interface {
	Create(event *MetricEvent) error
	GetByExperiment(experimentID string) ([]*MetricEvent, error)
	GetAggregated(experimentID, variantID, metricName string) (float64, int, error)
}
