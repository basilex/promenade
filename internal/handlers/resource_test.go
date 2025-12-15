package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/basilex/promenade/internal/models"
	"github.com/basilex/promenade/internal/storage"
	"github.com/gorilla/mux"
)

func setupTestHandler() (*ResourceHandler, *storage.MemoryStorage) {
	store := storage.NewMemoryStorage()
	logger := log.New(os.Stdout, "[test] ", log.LstdFlags)
	handler := NewResourceHandler(store, logger)
	return handler, store
}

func TestCreateResource(t *testing.T) {
	handler, _ := setupTestHandler()

	reqBody := models.CreateResourceRequest{
		Name:        "Test Resource",
		Description: "A test resource",
		Type:        "test",
		Status:      "active",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/resources", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.CreateResource(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rec.Code)
	}

	var resource models.Resource
	if err := json.NewDecoder(rec.Body).Decode(&resource); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resource.Name != reqBody.Name {
		t.Errorf("Expected name %s, got %s", reqBody.Name, resource.Name)
	}

	if resource.ID == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestCreateResourceWithoutName(t *testing.T) {
	handler, _ := setupTestHandler()

	reqBody := models.CreateResourceRequest{
		Description: "A test resource",
		Type:        "test",
		Status:      "active",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/resources", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.CreateResource(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetResources(t *testing.T) {
	handler, store := setupTestHandler()

	// Create test resources
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

	req := httptest.NewRequest("GET", "/api/resources", nil)
	rec := httptest.NewRecorder()

	handler.GetResources(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var resources []*models.Resource
	if err := json.NewDecoder(rec.Body).Decode(&resources); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(resources))
	}
}

func TestGetResource(t *testing.T) {
	handler, store := setupTestHandler()

	// Create a test resource
	resource, _ := store.Create(&models.CreateResourceRequest{
		Name:   "Test Resource",
		Type:   "test",
		Status: "active",
	})

	req := httptest.NewRequest("GET", "/api/resources/"+resource.ID, nil)
	rec := httptest.NewRecorder()

	// Setup mux vars
	req = mux.SetURLVars(req, map[string]string{"id": resource.ID})

	handler.GetResource(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var retrievedResource models.Resource
	if err := json.NewDecoder(rec.Body).Decode(&retrievedResource); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if retrievedResource.ID != resource.ID {
		t.Errorf("Expected ID %s, got %s", resource.ID, retrievedResource.ID)
	}
}

func TestGetResourceNotFound(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest("GET", "/api/resources/nonexistent", nil)
	rec := httptest.NewRecorder()

	// Setup mux vars
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})

	handler.GetResource(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUpdateResource(t *testing.T) {
	handler, store := setupTestHandler()

	// Create a test resource
	resource, _ := store.Create(&models.CreateResourceRequest{
		Name:   "Test Resource",
		Type:   "test",
		Status: "active",
	})

	updateReq := models.UpdateResourceRequest{
		Name:   "Updated Resource",
		Status: "inactive",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/api/resources/"+resource.ID, bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	// Setup mux vars
	req = mux.SetURLVars(req, map[string]string{"id": resource.ID})

	handler.UpdateResource(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var updatedResource models.Resource
	if err := json.NewDecoder(rec.Body).Decode(&updatedResource); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updatedResource.Name != updateReq.Name {
		t.Errorf("Expected name %s, got %s", updateReq.Name, updatedResource.Name)
	}

	if updatedResource.Status != updateReq.Status {
		t.Errorf("Expected status %s, got %s", updateReq.Status, updatedResource.Status)
	}
}

func TestUpdateResourceNotFound(t *testing.T) {
	handler, _ := setupTestHandler()

	updateReq := models.UpdateResourceRequest{
		Name: "Updated Resource",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/api/resources/nonexistent", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	// Setup mux vars
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})

	handler.UpdateResource(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestDeleteResource(t *testing.T) {
	handler, store := setupTestHandler()

	// Create a test resource
	resource, _ := store.Create(&models.CreateResourceRequest{
		Name:   "Test Resource",
		Type:   "test",
		Status: "active",
	})

	req := httptest.NewRequest("DELETE", "/api/resources/"+resource.ID, nil)
	rec := httptest.NewRecorder()

	// Setup mux vars
	req = mux.SetURLVars(req, map[string]string{"id": resource.ID})

	handler.DeleteResource(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, rec.Code)
	}

	// Verify the resource is deleted
	_, err := store.GetByID(resource.ID)
	if err != storage.ErrResourceNotFound {
		t.Error("Expected resource to be deleted")
	}
}

func TestDeleteResourceNotFound(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest("DELETE", "/api/resources/nonexistent", nil)
	rec := httptest.NewRecorder()

	// Setup mux vars
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})

	handler.DeleteResource(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rec.Code)
	}
}
