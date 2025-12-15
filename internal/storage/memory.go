package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/basilex/promenade/internal/models"
	"github.com/google/uuid"
)

var (
	// ErrResourceNotFound is returned when a resource is not found
	ErrResourceNotFound = errors.New("resource not found")
)

// MemoryStorage provides in-memory storage for resources
type MemoryStorage struct {
	mu        sync.RWMutex
	resources map[string]*models.Resource
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		resources: make(map[string]*models.Resource),
	}
}

// Create adds a new resource to storage
func (s *MemoryStorage) Create(req *models.CreateResourceRequest) (*models.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resource := &models.Resource{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Status:      req.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.resources[resource.ID] = resource
	return resource, nil
}

// GetAll returns all resources
func (s *MemoryStorage) GetAll() []*models.Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]*models.Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		resources = append(resources, resource)
	}
	return resources
}

// GetByID returns a resource by ID
func (s *MemoryStorage) GetByID(id string) (*models.Resource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resource, exists := s.resources[id]
	if !exists {
		return nil, ErrResourceNotFound
	}
	return resource, nil
}

// Update updates an existing resource
func (s *MemoryStorage) Update(id string, req *models.UpdateResourceRequest) (*models.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resource, exists := s.resources[id]
	if !exists {
		return nil, ErrResourceNotFound
	}

	if req.Name != "" {
		resource.Name = req.Name
	}
	if req.Description != "" {
		resource.Description = req.Description
	}
	if req.Type != "" {
		resource.Type = req.Type
	}
	if req.Status != "" {
		resource.Status = req.Status
	}
	resource.UpdatedAt = time.Now()

	return resource, nil
}

// Delete removes a resource from storage
func (s *MemoryStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.resources[id]; !exists {
		return ErrResourceNotFound
	}

	delete(s.resources, id)
	return nil
}
