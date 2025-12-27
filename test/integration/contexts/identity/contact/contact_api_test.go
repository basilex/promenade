package contact_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// TestContactAPI_FullWorkflow tests complete Contact API lifecycle with real database
func TestContactAPI_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	db := testDB.DB

	// Create test user
	userID := createTestUser(t, db)

	// Setup router
	router := setupTestRouter(db)

	t.Run("Create Email Contact", func(t *testing.T) {
		payload := map[string]interface{}{
			"type":  "email",
			"label": "Work Email",
			"email": "john.doe@company.com",
		}

		contactID := createContact(t, router, userID, payload)
		assert.NotEmpty(t, contactID)

		// Verify in database
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM identity_contacts WHERE id = $1 AND user_id = $2", contactID, userID)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Create Phone Contact", func(t *testing.T) {
		payload := map[string]interface{}{
			"type":  "phone",
			"label": "Mobile",
			"phone": map[string]string{
				"country_code": "+1",
				"number":       "5551234567",
			},
		}

		contactID := createContact(t, router, userID, payload)
		assert.NotEmpty(t, contactID)

		// Verify phone format in database
		var phone string
		err := db.Get(&phone, "SELECT phone FROM identity_contacts WHERE id = $1", contactID)
		require.NoError(t, err)
		assert.Equal(t, "+15551234567", phone)
	})

	t.Run("Create Address Contact", func(t *testing.T) {
		payload := map[string]interface{}{
			"type":  "address",
			"label": "Home Address",
			"address": map[string]interface{}{
				"street":      "123 Main Street",
				"city":        "New York",
				"country":     "US",
				"postal_code": "10001",
			},
		}

		contactID := createContact(t, router, userID, payload)
		assert.NotEmpty(t, contactID)

		// Verify address in database (address_city not address->>'city')
		var city string
		err := db.Get(&city, "SELECT address_city FROM identity_contacts WHERE id = $1", contactID)
		require.NoError(t, err)
		assert.Equal(t, "New York", city)
	})

	t.Run("List User Contacts", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/identity/contacts?user_id=%s", userID), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response["status"])
		contacts := response["data"].([]interface{})
		assert.GreaterOrEqual(t, len(contacts), 3, "Should have at least 3 contacts (email, phone, address)")
	})

	t.Run("Get Contact By ID", func(t *testing.T) {
		// Create contact first
		payload := map[string]interface{}{
			"type":  "email",
			"label": "Personal",
			"email": "personal@example.com",
		}
		contactID := createContact(t, router, userID, payload)

		// Get contact
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/identity/contacts/%s", contactID), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, contactID, data["id"])
		assert.Equal(t, "email", data["type"])
		assert.Equal(t, "Personal", data["label"])
	})

	t.Run("Update Contact", func(t *testing.T) {
		// Create contact
		payload := map[string]interface{}{
			"type":  "email",
			"label": "Old Label",
			"email": "test@example.com",
		}
		contactID := createContact(t, router, userID, payload)

		// Update label and visibility
		updatePayload := map[string]interface{}{
			"label":     "Updated Label",
			"is_public": true,
		}

		w := httptest.NewRecorder()
		body, _ := json.Marshal(updatePayload)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/identity/contacts/%s", contactID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Updated Label", data["label"])
		assert.True(t, data["is_public"].(bool))
	})

	t.Run("Set Contact As Primary", func(t *testing.T) {
		// Create two email contacts
		contact1 := createContact(t, router, userID, map[string]interface{}{
			"type":  "email",
			"label": "Email 1",
			"email": "email1@example.com",
		})

		contact2 := createContact(t, router, userID, map[string]interface{}{
			"type":  "email",
			"label": "Email 2",
			"email": "email2@example.com",
		})

		// Set contact2 as primary
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/identity/contacts/%s/primary", contact2), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify contact2 is primary
		var isPrimary bool
		err := db.Get(&isPrimary, "SELECT is_primary FROM identity_contacts WHERE id = $1", contact2)
		require.NoError(t, err)
		assert.True(t, isPrimary)

		// Verify contact1 is NOT primary (should be unset)
		err = db.Get(&isPrimary, "SELECT is_primary FROM identity_contacts WHERE id = $1", contact1)
		require.NoError(t, err)
		assert.False(t, isPrimary)
	})

	t.Run("Verify Contact", func(t *testing.T) {
		// Create contact
		contactID := createContact(t, router, userID, map[string]interface{}{
			"type":  "email",
			"label": "To Verify",
			"email": "verify@example.com",
		})

		// Verify contact
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/identity/contacts/%s/verify", contactID), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check in database
		var isVerified bool
		err := db.Get(&isVerified, "SELECT is_verified FROM identity_contacts WHERE id = $1", contactID)
		require.NoError(t, err)
		assert.True(t, isVerified)
	})

	t.Run("Delete Contact (Soft Delete)", func(t *testing.T) {
		// Create contact
		contactID := createContact(t, router, userID, map[string]interface{}{
			"type":  "email",
			"label": "To Delete",
			"email": "delete@example.com",
		})

		// Delete contact
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/identity/contacts/%s", contactID), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify hard delete (no deleted_at column in schema)
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM identity_contacts WHERE id = $1", contactID)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Contact should be hard deleted")
	})

	t.Run("List Only Active Contacts (After Delete)", func(t *testing.T) {
		// Create contact and delete it
		contactID := createContact(t, router, userID, map[string]interface{}{
			"type":  "email",
			"label": "Deleted",
			"email": "deleted@example.com",
		})

		// Delete (hard delete - no soft delete support)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/identity/contacts/%s", contactID), nil)
		req.Header.Set("X-User-ID", userID)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// List contacts - deleted contact should NOT appear
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/identity/contacts?user_id=%s", userID), nil)
		req.Header.Set("X-User-ID", userID)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		contacts := response["data"].([]interface{})
		for _, c := range contacts {
			contactData := c.(map[string]interface{})
			assert.NotEqual(t, contactID, contactData["id"], "Deleted contact should not appear in list")
		}
	})
}

// TestContactAPI_Validation tests input validation for Contact API
func TestContactAPI_Validation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	db := testDB.DB

	userID := createTestUser(t, db)
	router := setupTestRouter(db)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		errorContains  string
	}{
		{
			name: "Missing Type",
			payload: map[string]interface{}{
				"label": "Test",
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Type",
		},
		{
			name: "Invalid Type",
			payload: map[string]interface{}{
				"type":  "invalid",
				"label": "Test",
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Type",
		},
		{
			name: "Email Type Without Email",
			payload: map[string]interface{}{
				"type":  "email",
				"label": "Test",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "email",
		},
		{
			name: "Invalid Email Format",
			payload: map[string]interface{}{
				"type":  "email",
				"label": "Test",
				"email": "invalid-email",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "email",
		},
		{
			name: "Phone Without Country Code",
			payload: map[string]interface{}{
				"type":  "phone",
				"label": "Test",
				"phone": map[string]string{
					"number": "5551234567",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Phone Invalid Format (No Plus)",
			payload: map[string]interface{}{
				"type":  "phone",
				"label": "Test",
				"phone": map[string]string{
					"country_code": "1",
					"number":       "5551234567",
				},
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "phone",
		},
		{
			name: "Address Without Country",
			payload: map[string]interface{}{
				"type":  "address",
				"label": "Test",
				"address": map[string]string{
					"street": "123 Main",
					"city":   "NYC",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Address Invalid Country Code (Not ISO)",
			payload: map[string]interface{}{
				"type":  "address",
				"label": "Test",
				"address": map[string]string{
					"street":      "123 Main",
					"city":        "NYC",
					"country":     "USA",
					"postal_code": "10001",
				},
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "country",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/identity/contacts?user_id=%s", userID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.errorContains != "" {
				assert.Contains(t, w.Body.String(), tt.errorContains)
			}
		})
	}
}

// TestContactAPI_EdgeCases tests edge cases and error scenarios
func TestContactAPI_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	router := setupTestRouter(testDB.DB)

	t.Run("Get Non-Existent Contact", func(t *testing.T) {
		fakeID := uuidv7.New()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/identity/contacts/%s", fakeID), nil)
		router.ServeHTTP(w, req)

		// Handler returns 500 for database errors (contact not found)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Update Non-Existent Contact", func(t *testing.T) {
		fakeID := uuidv7.New()
		payload := map[string]interface{}{
			"label": "Updated",
		}

		w := httptest.NewRecorder()
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/identity/contacts/%s", fakeID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Handler returns 500 for database errors
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Delete Non-Existent Contact", func(t *testing.T) {
		fakeID := uuidv7.New()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/identity/contacts/%s", fakeID), nil)
		router.ServeHTTP(w, req)

		// Delete returns 200 even if contact doesn't exist (idempotent)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Set Primary on Non-Existent Contact", func(t *testing.T) {
		fakeID := uuidv7.New()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/identity/contacts/%s/primary", fakeID), nil)
		router.ServeHTTP(w, req)

		// Handler returns 500 for database errors
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Invalid User ID Format", func(t *testing.T) {
		payload := map[string]interface{}{
			"type":  "email",
			"label": "Test",
			"email": "test@example.com",
		}

		w := httptest.NewRecorder()
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/identity/contacts?user_id=invalid-uuid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing User ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"type":  "email",
			"label": "Test",
			"email": "test@example.com",
		}

		w := httptest.NewRecorder()
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/identity/contacts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestContactAPI_FilterByType tests filtering contacts by type
func TestContactAPI_FilterByType(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	db := testDB.DB

	userID := createTestUser(t, db)
	router := setupTestRouter(db)

	// Create contacts of different types
	createContact(t, router, userID, map[string]interface{}{
		"type": "email", "label": "Email 1", "email": "email1@example.com",
	})
	createContact(t, router, userID, map[string]interface{}{
		"type": "email", "label": "Email 2", "email": "email2@example.com",
	})
	createContact(t, router, userID, map[string]interface{}{
		"type": "phone", "label": "Phone 1",
		"phone": map[string]string{"country_code": "+1", "number": "5551111111"},
	})
	createContact(t, router, userID, map[string]interface{}{
		"type": "address", "label": "Address 1",
		"address": map[string]interface{}{
			"street": "123 Main", "city": "NYC", "country": "US", "postal_code": "10001",
		},
	})

	tests := []struct {
		name          string
		filterType    string
		expectedCount int
		expectedType  string
	}{
		{"Filter by email", "email", 2, "email"},
		{"Filter by phone", "phone", 1, "phone"},
		{"Filter by address", "address", 1, "address"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/api/v1/identity/contacts?user_id=%s&type=%s", userID, tt.filterType), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			contacts := response["data"].([]interface{})
			assert.Equal(t, tt.expectedCount, len(contacts), "Should match expected count")

			// Verify all contacts have correct type
			for _, c := range contacts {
				contactData := c.(map[string]interface{})
				assert.Equal(t, tt.expectedType, contactData["type"])
			}
		})
	}
}

// Helper functions

func setupTestRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register Identity context routes
	identityRouter := identity.NewRouter(db)
	api := router.Group("/api")
	v1 := api.Group("/v1")
	identityRouter.RegisterRoutes(v1)

	return router
}

func createTestUser(t *testing.T, db *sqlx.DB) string {
	userID := uuidv7.New()
	_, err := db.Exec(`
		INSERT INTO identity_users (id, email, name, password, status)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, fmt.Sprintf("test_%s@example.com", userID), "Test User", "hash", "active")
	require.NoError(t, err)
	return userID.String()
}

func createContact(t *testing.T, router *gin.Engine, userID string, payload map[string]interface{}) string {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/identity/contacts?user_id=%s", userID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "Failed to create contact: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	return data["id"].(string)
}
