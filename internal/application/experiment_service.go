package application

import (
	"fmt"
	"time"

	"github.com/cenkcorapci/odak/internal/domain"
	"github.com/google/uuid"
)

// ExperimentService handles experiment-related business logic
type ExperimentService struct {
	experimentRepo  domain.ExperimentRepository
	assignmentRepo  domain.AssignmentRepository
	metricRepo      domain.MetricRepository
	bucketingService *domain.BucketingService
}

// NewExperimentService creates a new experiment service
func NewExperimentService(
	experimentRepo domain.ExperimentRepository,
	assignmentRepo domain.AssignmentRepository,
	metricRepo domain.MetricRepository,
	bucketingService *domain.BucketingService,
) *ExperimentService {
	return &ExperimentService{
		experimentRepo:   experimentRepo,
		assignmentRepo:   assignmentRepo,
		metricRepo:       metricRepo,
		bucketingService: bucketingService,
	}
}

// CreateExperiment creates a new experiment
func (s *ExperimentService) CreateExperiment(exp *domain.Experiment) error {
	exp.ID = uuid.New().String()
	exp.CreatedAt = time.Now()
	exp.UpdatedAt = time.Now()
	
	// Validate variants
	if len(exp.Variants) == 0 {
		return fmt.Errorf("experiment must have at least one variant")
	}
	
	totalTraffic := 0
	for i := range exp.Variants {
		if exp.Variants[i].ID == "" {
			exp.Variants[i].ID = uuid.New().String()
		}
		totalTraffic += exp.Variants[i].TrafficPct
	}
	
	if totalTraffic != 10000 {
		return fmt.Errorf("variant traffic percentages must sum to 10000 (got %d)", totalTraffic)
	}
	
	return s.experimentRepo.Create(exp)
}

// GetExperiment retrieves an experiment by ID
func (s *ExperimentService) GetExperiment(id string) (*domain.Experiment, error) {
	return s.experimentRepo.GetByID(id)
}

// ListExperiments retrieves all experiments
func (s *ExperimentService) ListExperiments() ([]*domain.Experiment, error) {
	return s.experimentRepo.List()
}

// UpdateExperiment updates an existing experiment
func (s *ExperimentService) UpdateExperiment(exp *domain.Experiment) error {
	exp.UpdatedAt = time.Now()
	return s.experimentRepo.Update(exp)
}

// AssignVariant assigns an entity to a variant
func (s *ExperimentService) AssignVariant(experimentID, entityID string) (*domain.Assignment, error) {
	// Check for existing assignment
	existing, err := s.assignmentRepo.GetByExperimentAndEntity(experimentID, entityID)
	if err == nil && existing != nil {
		return existing, nil
	}
	
	// Get experiment
	exp, err := s.experimentRepo.GetByID(experimentID)
	if err != nil {
		return nil, fmt.Errorf("experiment not found: %w", err)
	}
	
	if exp.Status != "active" {
		return nil, fmt.Errorf("experiment is not active")
	}
	
	// Use bucketing service to assign variant
	variant, err := s.bucketingService.AssignVariant(experimentID, entityID, exp.Variants)
	if err != nil {
		return nil, fmt.Errorf("failed to assign variant: %w", err)
	}
	
	// Create assignment
	assignment := &domain.Assignment{
		ExperimentID: experimentID,
		EntityID:     entityID,
		VariantID:    variant.ID,
		AssignedAt:   time.Now(),
	}
	
	err = s.assignmentRepo.Create(assignment)
	if err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}
	
	return assignment, nil
}

// TrackMetric records a metric event
func (s *ExperimentService) TrackMetric(event *domain.MetricEvent) error {
	event.ID = uuid.New().String()
	event.Timestamp = time.Now()
	return s.metricRepo.Create(event)
}

// GetMetricAggregation retrieves aggregated metrics for a variant
func (s *ExperimentService) GetMetricAggregation(experimentID, variantID, metricName string) (float64, int, error) {
	return s.metricRepo.GetAggregated(experimentID, variantID, metricName)
}
