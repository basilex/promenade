package storage

import (
	"testing"

	"github.com/basilex/promenade/internal/models"
)

func TestCreate(t *testing.T) {
	store := NewMemoryStorage()

	req := &models.CreateResourceRequest{
		Name:        "Test Resource",
		Description: "Test Description",
		Type:        "test",
		Status:      "active",
	}

	resource, err := store.Create(req)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	if resource.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if resource.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, resource.Name)
	}

	if resource.Description != req.Description {
		t.Errorf("Expected description %s, got %s", req.Description, resource.Description)
	}
}

func TestGetAll(t *testing.T) {
	store := NewMemoryStorage()

	// Create multiple resources
	store.Create(&models.CreateResourceRequest{
		Name:   "Resource 1",
		Type:   "test",
		Status: "active",
	})
	store.Create(&models.CreateResourceRequest{
		Name:   "Resource 2",
		Type:   "test",
		Status: "inactive",
	})

	resources := store.GetAll()

	if len(resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(resources))
	}
}

func TestGetByID(t *testing.T) {
	store := NewMemoryStorage()

	// Create a resource
	created, _ := store.Create(&models.CreateResourceRequest{
		Name:   "Test Resource",
		Type:   "test",
		Status: "active",
	})

	// Retrieve the resource
	resource, err := store.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get resource: %v", err)
	}

	if resource.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, resource.ID)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.GetByID("nonexistent")
	if err != ErrResourceNotFound {
		t.Errorf("Expected ErrResourceNotFound, got %v", err)
	}
}

func TestUpdate(t *testing.T) {
	store := NewMemoryStorage()

	// Create a resource
	created, _ := store.Create(&models.CreateResourceRequest{
		Name:   "Test Resource",
		Type:   "test",
		Status: "active",
	})

	// Update the resource
	updateReq := &models.UpdateResourceRequest{
		Name:   "Updated Resource",
		Status: "inactive",
	}

	updated, err := store.Update(created.ID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update resource: %v", err)
	}

	if updated.Name != updateReq.Name {
		t.Errorf("Expected name %s, got %s", updateReq.Name, updated.Name)
	}

	if updated.Status != updateReq.Status {
		t.Errorf("Expected status %s, got %s", updateReq.Status, updated.Status)
	}

	if updated.Type != created.Type {
		t.Errorf("Expected type %s to remain unchanged, got %s", created.Type, updated.Type)
	}
}

func TestUpdateNotFound(t *testing.T) {
	store := NewMemoryStorage()

	updateReq := &models.UpdateResourceRequest{
		Name: "Updated Resource",
	}

	_, err := store.Update("nonexistent", updateReq)
	if err != ErrResourceNotFound {
		t.Errorf("Expected ErrResourceNotFound, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	store := NewMemoryStorage()

	// Create a resource
	created, _ := store.Create(&models.CreateResourceRequest{
		Name:   "Test Resource",
		Type:   "test",
		Status: "active",
	})

	// Delete the resource
	err := store.Delete(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete resource: %v", err)
	}

	// Verify it's deleted
	_, err = store.GetByID(created.ID)
	if err != ErrResourceNotFound {
		t.Error("Expected resource to be deleted")
	}
}

func TestDeleteNotFound(t *testing.T) {
	store := NewMemoryStorage()

	err := store.Delete("nonexistent")
	if err != ErrResourceNotFound {
		t.Errorf("Expected ErrResourceNotFound, got %v", err)
	}
}
