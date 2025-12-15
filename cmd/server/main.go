package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/basilex/promenade/internal/handlers"
	"github.com/basilex/promenade/internal/storage"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "[promenade] ", log.LstdFlags)

	// Initialize storage
	store := storage.NewMemoryStorage()

	// Initialize handlers
	resourceHandler := handlers.NewResourceHandler(store, logger)

	// Setup router
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/resources", resourceHandler.GetResources).Methods("GET")
	api.HandleFunc("/resources", resourceHandler.CreateResource).Methods("POST")
	api.HandleFunc("/resources/{id}", resourceHandler.GetResource).Methods("GET")
	api.HandleFunc("/resources/{id}", resourceHandler.UpdateResource).Methods("PUT")
	api.HandleFunc("/resources/{id}", resourceHandler.DeleteResource).Methods("DELETE")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	// Middleware for logging
	router.Use(loggingMiddleware(logger))

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configure server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	logger.Printf("Starting server on port %s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server failed to start: %v", err)
	}
}

// loggingMiddleware logs each incoming request
func loggingMiddleware(logger *log.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger.Printf("%s %s", r.Method, r.RequestURI)
			next.ServeHTTP(w, r)
			logger.Printf("%s %s - completed in %v", r.Method, r.RequestURI, time.Since(start))
		})
	}
}
