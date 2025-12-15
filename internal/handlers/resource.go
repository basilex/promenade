package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/basilex/promenade/internal/models"
	"github.com/basilex/promenade/internal/storage"
	"github.com/gorilla/mux"
)

// ResourceHandler handles HTTP requests for resources
type ResourceHandler struct {
	storage *storage.MemoryStorage
	logger  *log.Logger
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler(storage *storage.MemoryStorage, logger *log.Logger) *ResourceHandler {
	return &ResourceHandler{
		storage: storage,
		logger:  logger,
	}
}

// CreateResource handles POST /api/resources
func (h *ResourceHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
	var req models.CreateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Error decoding request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	resource, err := h.storage.Create(&req)
	if err != nil {
		h.logger.Printf("Error creating resource: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create resource")
		return
	}

	h.logger.Printf("Created resource: %s", resource.ID)
	respondWithJSON(w, http.StatusCreated, resource)
}

// GetResources handles GET /api/resources
func (h *ResourceHandler) GetResources(w http.ResponseWriter, r *http.Request) {
	resources := h.storage.GetAll()
	h.logger.Printf("Retrieved %d resources", len(resources))
	respondWithJSON(w, http.StatusOK, resources)
}

// GetResource handles GET /api/resources/{id}
func (h *ResourceHandler) GetResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	resource, err := h.storage.GetByID(id)
	if err != nil {
		if err == storage.ErrResourceNotFound {
			respondWithError(w, http.StatusNotFound, "Resource not found")
			return
		}
		h.logger.Printf("Error retrieving resource: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve resource")
		return
	}

	h.logger.Printf("Retrieved resource: %s", id)
	respondWithJSON(w, http.StatusOK, resource)
}

// UpdateResource handles PUT /api/resources/{id}
func (h *ResourceHandler) UpdateResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Error decoding request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	resource, err := h.storage.Update(id, &req)
	if err != nil {
		if err == storage.ErrResourceNotFound {
			respondWithError(w, http.StatusNotFound, "Resource not found")
			return
		}
		h.logger.Printf("Error updating resource: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update resource")
		return
	}

	h.logger.Printf("Updated resource: %s", id)
	respondWithJSON(w, http.StatusOK, resource)
}

// DeleteResource handles DELETE /api/resources/{id}
func (h *ResourceHandler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.storage.Delete(id)
	if err != nil {
		if err == storage.ErrResourceNotFound {
			respondWithError(w, http.StatusNotFound, "Resource not found")
			return
		}
		h.logger.Printf("Error deleting resource: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to delete resource")
		return
	}

	h.logger.Printf("Deleted resource: %s", id)
	w.WriteHeader(http.StatusNoContent)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to marshal response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
