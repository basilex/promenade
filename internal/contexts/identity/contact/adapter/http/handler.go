package http

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ContactHandler handles HTTP requests for contact operations
type ContactHandler struct {
	usecase contact.IUseCase
}

// NewContactHandler creates a new contact handler
func NewContactHandler(usecase contact.IUseCase) *ContactHandler {
	return &ContactHandler{
		usecase: usecase,
	}
}

// Create handles POST /contacts
func (h *ContactHandler) Create(c *gin.Context) {
	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: Get user ID from JWT context
	// For now, get from query parameter
	userID := c.Query("user_id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	// Convert DTO to entity
	contactEntity, err := CreateContactFromRequest(userID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create contact based on type
	var created *contact.Contact
	switch contactEntity.Type {
	case contact.ContactTypeEmail:
		created, err = h.usecase.CreateEmailContact(
			c.Request.Context(),
			contactEntity.UserID,
			contactEntity.Email.Value(),
			contactEntity.Label,
			false, // isPrimary
		)
	case contact.ContactTypePhone:
		created, err = h.usecase.CreatePhoneContact(
			c.Request.Context(),
			contactEntity.UserID,
			contactEntity.Phone.Value(),
			contactEntity.Label,
			false, // isPrimary
		)
	case contact.ContactTypeAddress:
		created, err = h.usecase.CreateAddressContact(
			c.Request.Context(),
			contactEntity.UserID,
			contactEntity.Address.Street,
			contactEntity.Address.City,
			contactEntity.Address.Country,
			contactEntity.Address.PostalCode,
			contactEntity.Label,
			false, // isPrimary
		)
	default:
		response.BadRequest(c, "invalid contact type")
		return
	}

	if err != nil {
		response.InternalError(c, "Failed to create contact")
		return
	}

	response.Created(c, ToContactResponse(created))
}

// GetByID handles GET /contacts/:id
func (h *ContactHandler) GetByID(c *gin.Context) {
	contactID := c.Param("id")
	if contactID == "" {
		response.BadRequest(c, "contact_id is required")
		return
	}

	contactUUID, err := uuidv7.Parse(contactID)
	if err != nil {
		response.BadRequest(c, "invalid contact_id")
		return
	}

	contactEntity, err := h.usecase.GetContact(c.Request.Context(), contactUUID)
	if err != nil {
		if errors.Is(err, contact.ErrNotFound) {
			response.NotFound(c, "contact not found")
			return
		}
		response.InternalError(c, "Failed to retrieve contact")
		return
	}

	response.Success(c, ToContactResponse(contactEntity))
}

// Update handles PUT /contacts/:id
func (h *ContactHandler) Update(c *gin.Context) {
	contactID := c.Param("id")
	if contactID == "" {
		response.BadRequest(c, "contact_id is required")
		return
	}

	contactUUID, err := uuidv7.Parse(contactID)
	if err != nil {
		response.BadRequest(c, "invalid contact_id")
		return
	}

	var req UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get existing contact
	existing, err := h.usecase.GetContact(c.Request.Context(), contactUUID)
	if err != nil {
		if errors.Is(err, contact.ErrNotFound) {
			response.NotFound(c, "contact not found")
			return
		}
		response.InternalError(c, "Failed to retrieve contact")
		return
	}

	// Apply updates
	if req.Label != nil {
		existing.Label = *req.Label
	}

	if req.Email != nil && existing.Type == contact.ContactTypeEmail {
		if err := existing.SetEmail(*req.Email); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}

	if req.Phone != nil && existing.Type == contact.ContactTypePhone {
		phoneStr := req.Phone.CountryCode + req.Phone.Number
		if err := existing.SetPhone(phoneStr); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}

	if req.Address != nil && existing.Type == contact.ContactTypeAddress {
		if err := existing.SetAddress(
			req.Address.Street,
			req.Address.City,
			req.Address.Country,
			req.Address.PostalCode,
		); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}

	if req.IsPublic != nil {
		existing.IsPublic = *req.IsPublic
	}

	// Update contact
	err = h.usecase.UpdateContact(c.Request.Context(), existing)
	if err != nil {
		response.InternalError(c, "Failed to update contact")
		return
	}

	response.Success(c, ToContactResponse(existing))
}

// Delete handles DELETE /contacts/:id
func (h *ContactHandler) Delete(c *gin.Context) {
	contactID := c.Param("id")
	if contactID == "" {
		response.BadRequest(c, "contact_id is required")
		return
	}

	contactUUID, err := uuidv7.Parse(contactID)
	if err != nil {
		response.BadRequest(c, "invalid contact_id")
		return
	}

	err = h.usecase.DeleteContact(c.Request.Context(), contactUUID)
	if err != nil {
		if errors.Is(err, contact.ErrNotFound) {
			response.NotFound(c, "contact not found")
			return
		}
		response.InternalError(c, "Failed to delete contact")
		return
	}

	response.Success(c, gin.H{"message": "contact deleted successfully"})
}

// List handles GET /contacts
func (h *ContactHandler) List(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userUUID, err := uuidv7.Parse(userID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	contactType := c.Query("type")

	var contacts []*contact.Contact

	if contactType != "" {
		contacts, err = h.usecase.GetUserContactsByType(
			c.Request.Context(),
			userUUID,
			contact.ContactType(contactType),
		)
	} else {
		contacts, err = h.usecase.GetUserContacts(c.Request.Context(), userUUID)
	}

	if err != nil {
		response.InternalError(c, "Failed to list contacts")
		return
	}

	response.Success(c, ToContactListResponse(contacts))
}

// Verify handles PUT /contacts/:id/verify
func (h *ContactHandler) Verify(c *gin.Context) {
	contactID := c.Param("id")
	if contactID == "" {
		response.BadRequest(c, "contact_id is required")
		return
	}

	contactUUID, err := uuidv7.Parse(contactID)
	if err != nil {
		response.BadRequest(c, "invalid contact_id")
		return
	}

	err = h.usecase.VerifyContact(c.Request.Context(), contactUUID)
	if err != nil {
		if errors.Is(err, contact.ErrNotFound) {
			response.NotFound(c, "contact not found")
			return
		}
		response.InternalError(c, "Failed to verify contact")
		return
	}

	// Get updated contact
	contactEntity, err := h.usecase.GetContact(c.Request.Context(), contactUUID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve updated contact")
		return
	}

	response.Success(c, ToContactResponse(contactEntity))
}

// SetPrimary handles PUT /contacts/:id/primary
func (h *ContactHandler) SetPrimary(c *gin.Context) {
	contactID := c.Param("id")
	if contactID == "" {
		response.BadRequest(c, "contact_id is required")
		return
	}

	contactUUID, err := uuidv7.Parse(contactID)
	if err != nil {
		response.BadRequest(c, "invalid contact_id")
		return
	}

	err = h.usecase.SetAsPrimary(c.Request.Context(), contactUUID)
	if err != nil {
		if errors.Is(err, contact.ErrNotFound) {
			response.NotFound(c, "contact not found")
			return
		}
		response.InternalError(c, "Failed to set contact as primary")
		return
	}

	// Get updated contact
	contactEntity, err := h.usecase.GetContact(c.Request.Context(), contactUUID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve updated contact")
		return
	}

	response.Success(c, ToContactResponse(contactEntity))
}
