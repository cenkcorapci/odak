package domain

import (
	"fmt"
	"github.com/spaolacci/murmur3"
)

// BucketingService handles deterministic assignment of entities to variants
type BucketingService struct {
	salt string
}

// NewBucketingService creates a new bucketing service
func NewBucketingService(salt string) *BucketingService {
	return &BucketingService{salt: salt}
}

// AssignVariant deterministically assigns an entity to a variant
func (s *BucketingService) AssignVariant(experimentID, entityID string, variants []Variant) (*Variant, error) {
	if len(variants) == 0 {
		return nil, fmt.Errorf("no variants available")
	}

	// Create hash input from experimentID, entityID, and salt
	hashInput := fmt.Sprintf("%s:%s:%s", experimentID, entityID, s.salt)
	
	// Use MurmurHash3 for consistent hashing
	hash := murmur3.Sum32([]byte(hashInput))
	
	// Normalize to 0-9999
	bucket := int(hash % 10000)
	
	// Assign to variant based on traffic distribution (in basis points)
	cumulative := 0
	for i := range variants {
		cumulative += variants[i].TrafficPct
		if bucket < cumulative {
			return &variants[i], nil
		}
	}
	
	// Default to last variant if something goes wrong
	return &variants[len(variants)-1], nil
}

// GetBucket returns the bucket number (0-9999) for an entity
func (s *BucketingService) GetBucket(experimentID, entityID string) int {
	hashInput := fmt.Sprintf("%s:%s:%s", experimentID, entityID, s.salt)
	hash := murmur3.Sum32([]byte(hashInput))
	return int(hash % 10000)
}
