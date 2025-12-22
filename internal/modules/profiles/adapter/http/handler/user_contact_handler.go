package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/modules/profiles/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/internal/modules/profiles/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type UserContactHandler struct {
	contactUC usecase.UserContactUseCase
}

func NewUserContactHandler(contactUC usecase.UserContactUseCase) *UserContactHandler {
	return &UserContactHandler{
		contactUC: contactUC,
	}
}

// CreateContact godoc
// @Summary      Create user contact
// @Description  Creates a new contact for the authenticated user
// @Tags         user-contacts
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateUserContactRequest true "Contact data"
// @Success      201 {object} response.Response{data=dto.UserContactResponse}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      500 {object} response.Response
// @Security     BearerAuth
// @Router       /users/contacts [post]
func (h *UserContactHandler) CreateContact(c *gin.Context) {
	var req dto.CreateUserContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Get authenticated user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userUUID := userID.(uuidv7.UUID)

	// Parse time strings
	availableFrom, err := dto.ParseTimeString(req.AvailableFrom)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid available_from format (expected HH:MM:SS)", err)
		return
	}

	availableTo, err := dto.ParseTimeString(req.AvailableTo)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid available_to format (expected HH:MM:SS)", err)
		return
	}

	contact, err := h.contactUC.CreateContact(c.Request.Context(), userUUID,
		entity.ContactType(req.ContactType), req.ContactValue, req.Label, req.IsPublic,
		availableFrom, availableTo, req.AvailableDays, req.Timezone, req.Notes)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidContactType) {
			response.Error(c, http.StatusBadRequest, "invalid contact type", err)
			return
		}
		if errors.Is(err, usecase.ErrInvalidAvailability) {
			response.Error(c, http.StatusBadRequest, "invalid availability: 'from' time must be before 'to' time", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to create contact", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToUserContactResponse(contact))
}

// GetContact godoc
// @Summary      Get contact by ID
// @Description  Retrieves a specific contact by its ID
// @Tags         user-contacts
// @Accept       json
// @Produce      json
// @Param        id path string true "Contact ID (UUID)"
// @Success      200 {object} response.Response{data=dto.UserContactResponse}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Security     BearerAuth
// @Router       /users/contacts/{id} [get]
func (h *UserContactHandler) GetContact(c *gin.Context) {
	contactID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid contact ID", err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userUUID := userID.(uuidv7.UUID)

	contact, err := h.contactUC.GetContact(c.Request.Context(), contactID, userUUID)
	if err != nil {
		if errors.Is(err, usecase.ErrContactNotFound) {
			response.Error(c, http.StatusNotFound, "contact not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "unauthorized to access this contact", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get contact", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserContactResponse(contact))
}

// GetUserContacts godoc
// @Summary      Get user's contacts
// @Description  Retrieves all contacts for the authenticated user
// @Tags         user-contacts
// @Accept       json
// @Produce      json
// @Param        include_inactive query bool false "Include inactive contacts"
// @Success      200 {object} response.Response{data=[]dto.UserContactResponse}
// @Failure      401 {object} response.Response
// @Failure      500 {object} response.Response
// @Security     BearerAuth
// @Router       /users/contacts [get]
func (h *UserContactHandler) GetUserContacts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userUUID := userID.(uuidv7.UUID)

	includeInactive := c.Query("include_inactive") == "true"

	contacts, err := h.contactUC.GetUserContacts(c.Request.Context(), userUUID, userUUID, includeInactive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get contacts", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserContactListResponse(contacts))
}

// GetContactsByType godoc
// @Summary      Get contacts by type
// @Description  Retrieves user contacts filtered by type
// @Tags         user-contacts
// @Accept       json
// @Produce      json
// @Param        type path string true "Contact type" Enums(email, phone, telegram, whatsapp, viber, signal, skype, discord, linkedin, other)
// @Success      200 {object} response.Response{data=[]dto.UserContactResponse}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      500 {object} response.Response
// @Security     BearerAuth
// @Router       /users/contacts/type/{type} [get]
func (h *UserContactHandler) GetContactsByType(c *gin.Context) {
	contactType := entity.ContactType(c.Param("type"))
	if !contactType.IsValid() {
		response.Error(c, http.StatusBadRequest, "invalid contact type", nil)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userUUID := userID.(uuidv7.UUID)

	contacts, err := h.contactUC.GetUserContactsByType(c.Request.Context(), userUUID, contactType, userUUID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get contacts", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserContactListResponse(contacts))
}

// GetPrimaryContact godoc
// @Summary      Get primary contact by type
// @Description  Retrieves the primary contact for a specific type
// @Tags         user-contacts
// @Accept       json
// @Produce      json
// @Param        type path string true "Contact type" Enums(email, phone, telegram, whatsapp, viber, signal, skype, discord, linkedin, other)
// @Success      200 {object} response.Response{data=dto.UserContactResponse}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Security     BearerAuth
// @Router       /users/contacts/primary/{type} [get]
func (h *UserContactHandler) GetPrimaryContact(c *gin.Context) {
	contactType := entity.ContactType(c.Param("type"))
	if !contactType.IsValid() {
		response.Error(c, http.StatusBadRequest, "invalid contact type", nil)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userUUID := userID.(uuidv7.UUID)

	contact, err := h.contactUC.GetPrimaryContact(c.Request.Context(), userUUID, contactType, userUUID)
	if err != nil {
		if errors.Is(err, usecase.ErrContactNotFound) {
			response.Error(c, http.StatusNotFound, "primary contact not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get primary contact", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserContactResponse(contact))
}

// UpdateContact godoc
// @Summary      Update contact
// @Description  Updates an existing contact
// @Tags         user-contacts
// @Accept       json
// @Produce      json
// @Param        id path string true "Contact ID (UUID)"
// @Param        request body dto.UpdateUserContactRequest true "Updated contact data"
// @Success      200 {object} response.Response{data=dto.UserContactResponse}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Security     BearerAuth
// @Router       /users/contacts/{id} [put]
func (h *UserContactHandler) UpdateContact(c *gin.Context) {
	contactID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid contact ID", err)
		return
	}

	var req dto.UpdateUserContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userUUID := userID.(uuidv7.UUID)

	// Parse time strings
	availableFrom, err := dto.ParseTimeString(req.AvailableFrom)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid available_from format (expected HH:MM:SS)", err)
		return
	}

	availableTo, err := dto.ParseTimeString(req.AvailableTo)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid available_to format (expected HH:MM:SS)", err)
		return
	}

	contact, err := h.contactUC.UpdateContact(c.Request.Context(), contactID, userUUID,
		entity.ContactType(req.ContactType), req.ContactValue, req.Label, req.IsPublic,
		availableFrom, availableTo, req.AvailableDays, req.Timezone, req.Notes)
	if err != nil {
		if errors.Is(err, usecase.ErrContactNotFound) {
			response.Error(c, http.StatusNotFound, "contact not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "unauthorized to update this contact", err)
			return
		}
		if errors.Is(err, usecase.ErrInvalidContactType) {
			response.Error(c, http.StatusBadRequest, "invalid contact type", err)
			return
		}
		if errors.Is(err, usecase.ErrInvalidAvailability) {
			response.Error(c, http.StatusBadRequest, "invalid availability: 'from' time must be before 'to' time", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update contact", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserContactResponse(contact))
}

// DeleteContact godoc
// @Summary      Delete contact
// @Description  Deletes a contact
// @Tags         user-contacts
// @Accept       json
// @Produce      json
// @Param        id path string true "Contact ID (UUID)"
// @Success      204 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Security     BearerAuth
// @Router       /users/contacts/{id} [delete]
func (h *UserContactHandler) DeleteContact(c *gin.Context) {
	contactID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid contact ID", err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userUUID := userID.(uuidv7.UUID)

	if err := h.contactUC.DeleteContact(c.Request.Context(), contactID, userUUID); err != nil {
		if errors.Is(err, usecase.ErrContactNotFound) {
			response.Error(c, http.StatusNotFound, "contact not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "unauthorized to delete this contact", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete contact", err)
		return
	}

	response.Success(c, http.StatusNoContent, nil)
}

// SetPrimaryContact godoc
// @Summary      Set primary contact
// @Description  Sets a contact as the primary contact for its type
// @Tags         user-contacts
// @Accept       json
// @Produce      json
// @Param        id path string true "Contact ID (UUID)"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Security     BearerAuth
// @Router       /users/contacts/{id}/primary [post]
func (h *UserContactHandler) SetPrimaryContact(c *gin.Context) {
	contactID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid contact ID", err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userUUID := userID.(uuidv7.UUID)

	if err := h.contactUC.SetPrimaryContact(c.Request.Context(), contactID, userUUID); err != nil {
		if errors.Is(err, usecase.ErrContactNotFound) {
			response.Error(c, http.StatusNotFound, "contact not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "unauthorized to modify this contact", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to set primary contact", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "contact set as primary"})
}

// ToggleContactActive godoc
// @Summary      Toggle contact active status
// @Description  Toggles the active status of a contact
// @Tags         user-contacts
// @Accept       json
// @Produce      json
// @Param        id path string true "Contact ID (UUID)"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Security     BearerAuth
// @Router       /users/contacts/{id}/toggle [post]
func (h *UserContactHandler) ToggleContactActive(c *gin.Context) {
	contactID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid contact ID", err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userUUID := userID.(uuidv7.UUID)

	if err := h.contactUC.ToggleContactActive(c.Request.Context(), contactID, userUUID); err != nil {
		if errors.Is(err, usecase.ErrContactNotFound) {
			response.Error(c, http.StatusNotFound, "contact not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "unauthorized to modify this contact", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to toggle contact active status", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "contact active status toggled"})
}
