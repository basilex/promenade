package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/modules/notifications/domain/entity"
	"github.com/basilex/promenade/internal/modules/notifications/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// DTOs

// SendNotificationRequest is the request DTO for sending a notification
type SendNotificationRequest struct {
	Type     string         `json:"type" binding:"required,oneof=system security marketing product social"`
	Channel  string         `json:"channel" binding:"required,oneof=email sms push in_app"`
	Template string         `json:"template" binding:"omitempty"`
	Subject  string         `json:"subject" binding:"required_if=Channel email"`
	Content  string         `json:"content" binding:"required"`
	Data     map[string]any `json:"data,omitempty"`
}

// NotificationResponse is the response DTO for a notification
type NotificationResponse struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Type      string         `json:"type"`
	Channel   string         `json:"channel"`
	Status    string         `json:"status"`
	Template  string         `json:"template"`
	Subject   string         `json:"subject"`
	Content   string         `json:"content"`
	Data      map[string]any `json:"data,omitempty"`
	SentAt    *string        `json:"sent_at,omitempty"`
	OpenedAt  *string        `json:"opened_at,omitempty"`
	ClickedAt *string        `json:"clicked_at,omitempty"`
	CreatedAt string         `json:"created_at"`
}

// UpdatePreferencesRequest is the request DTO for updating user preferences
type UpdatePreferencesRequest struct {
	EmailEnabled     *bool   `json:"email_enabled"`
	SMSEnabled       *bool   `json:"sms_enabled"`
	PushEnabled      *bool   `json:"push_enabled"`
	InAppEnabled     *bool   `json:"in_app_enabled"`
	SystemEnabled    *bool   `json:"system_enabled"`
	SecurityEnabled  *bool   `json:"security_enabled"`
	MarketingEnabled *bool   `json:"marketing_enabled"`
	ProductEnabled   *bool   `json:"product_enabled"`
	SocialEnabled    *bool   `json:"social_enabled"`
	QuietHoursStart  *string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd    *string `json:"quiet_hours_end,omitempty"`
	Timezone         *string `json:"timezone,omitempty"`
}

// PreferenceResponse is the response DTO for user preferences
type PreferenceResponse struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	EmailEnabled     bool    `json:"email_enabled"`
	SMSEnabled       bool    `json:"sms_enabled"`
	PushEnabled      bool    `json:"push_enabled"`
	InAppEnabled     bool    `json:"in_app_enabled"`
	SystemEnabled    bool    `json:"system_enabled"`
	SecurityEnabled  bool    `json:"security_enabled"`
	MarketingEnabled bool    `json:"marketing_enabled"`
	ProductEnabled   bool    `json:"product_enabled"`
	SocialEnabled    bool    `json:"social_enabled"`
	QuietHoursStart  *string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd    *string `json:"quiet_hours_end,omitempty"`
	Timezone         string  `json:"timezone"`
}

// NotificationHandler handles notification HTTP requests
type NotificationHandler struct {
	useCase usecase.INotificationUseCase
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(useCase usecase.INotificationUseCase) *NotificationHandler {
	return &NotificationHandler{
		useCase: useCase,
	}
}

// SendNotification handles POST /notifications
// @Summary Send a notification
// @Description Send a notification to the authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SendNotificationRequest true "Notification data"
// @Success 201 {object} response.Response{data=NotificationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications [post]
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	ctx := c.Request.Context()

	// Get authenticated user ID from context
	userIDValue := c.GetString("user_id")
	if userIDValue == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", errors.New("user not authenticated"))
		return
	}

	userID, err := uuidv7.Parse(userIDValue)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", err)
		return
	}

	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	notification, err := h.useCase.SendNotification(
		ctx,
		userID,
		entity.NotificationType(req.Type),
		entity.NotificationChannel(req.Channel),
		req.Template,
		req.Subject,
		req.Content,
		req.Data,
	)
	if err != nil {
		switch err {
		case usecase.ErrChannelDisabled:
			response.Error(c, http.StatusBadRequest, "CHANNEL_DISABLED", err)
		case usecase.ErrTypeDisabled:
			response.Error(c, http.StatusBadRequest, "TYPE_DISABLED", err)
		case usecase.ErrQuietHours:
			response.Error(c, http.StatusBadRequest, "QUIET_HOURS", err)
		default:
			response.Error(c, http.StatusInternalServerError, "SEND_FAILED", err)
		}
		return
	}

	response.Success(c, http.StatusCreated, toNotificationResponse(notification))
}

// GetNotifications handles GET /notifications
// @Summary Get user notifications
// @Description Get paginated list of notifications for authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} response.Response{data=[]NotificationResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	ctx := c.Request.Context()

	userIDValue := c.GetString("user_id")
	if userIDValue == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", errors.New("user not authenticated"))
		return
	}

	userID, err := uuidv7.Parse(userIDValue)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", err)
		return
	}

	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)
	offset := (page - 1) * pageSize

	notifications, total, err := h.useCase.GetUserNotifications(ctx, userID, pageSize, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", err)
		return
	}

	notificationResponses := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		notificationResponses[i] = toNotificationResponse(n)
	}

	response.SuccessWithPagination(c, http.StatusOK, notificationResponses, total, page, pageSize)
}

// GetNotification handles GET /notifications/:id
// @Summary Get notification by ID
// @Description Get a specific notification by ID
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID (UUID)"
// @Success 200 {object} response.Response{data=NotificationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id} [get]
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	notificationID, err := uuidv7.Parse(idParam)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", errors.New("invalid notification ID format"))
		return
	}

	notification, err := h.useCase.GetNotification(ctx, notificationID)
	if err != nil {
		if err == usecase.ErrNotificationNotFound {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", errors.New("notification not found"))
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", err)
		return
	}

	response.Success(c, http.StatusOK, toNotificationResponse(notification))
}

// GetUnreadCount handles GET /notifications/unread-count
// @Summary Get unread notification count
// @Description Get count of unread notifications for authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=map[string]int}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	ctx := c.Request.Context()

	userIDValue := c.GetString("user_id")
	if userIDValue == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", errors.New("user not authenticated"))
		return
	}

	userID, err := uuidv7.Parse(userIDValue)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", err)
		return
	}

	count, err := h.useCase.GetUnreadCount(ctx, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", err)
		return
	}

	response.Success(c, http.StatusOK, map[string]int{"count": count})
}

// MarkAsOpened handles POST /notifications/:id/opened
// @Summary Mark notification as opened
// @Description Mark a notification as opened
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID (UUID)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id}/opened [post]
func (h *NotificationHandler) MarkAsOpened(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	notificationID, err := uuidv7.Parse(idParam)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", errors.New("invalid notification ID format"))
		return
	}

	if err := h.useCase.MarkAsOpened(ctx, notificationID); err != nil {
		if err == usecase.ErrNotificationNotFound {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", errors.New("notification not found"))
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "notification marked as opened"})
}

// GetPreferences handles GET /notifications/preferences
// @Summary Get user notification preferences
// @Description Get notification preferences for authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=PreferenceResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/preferences [get]
func (h *NotificationHandler) GetPreferences(c *gin.Context) {
	ctx := c.Request.Context()

	userIDValue := c.GetString("user_id")
	if userIDValue == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", errors.New("user not authenticated"))
		return
	}

	userID, err := uuidv7.Parse(userIDValue)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", err)
		return
	}

	prefs, err := h.useCase.GetUserPreferences(ctx, userID)
	if err != nil {
		if err == usecase.ErrPreferenceNotFound {
			// Create default preferences
			prefs, err = h.useCase.CreateDefaultPreferences(ctx, userID)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", err)
				return
			}
		} else {
			response.Error(c, http.StatusInternalServerError, "GET_FAILED", err)
			return
		}
	}

	response.Success(c, http.StatusOK, toPreferenceResponse(prefs))
}

// UpdatePreferences handles PUT /notifications/preferences
// @Summary Update user notification preferences
// @Description Update notification preferences for authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdatePreferencesRequest true "Preference data"
// @Success 200 {object} response.Response{data=PreferenceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/preferences [put]
func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	ctx := c.Request.Context()

	userIDValue := c.GetString("user_id")
	if userIDValue == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", errors.New("user not authenticated"))
		return
	}

	userID, err := uuidv7.Parse(userIDValue)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", err)
		return
	}

	var req UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	// Get existing preferences
	prefs, err := h.useCase.GetUserPreferences(ctx, userID)
	if err != nil {
		if err == usecase.ErrPreferenceNotFound {
			// Create default preferences first
			prefs, err = h.useCase.CreateDefaultPreferences(ctx, userID)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", err)
				return
			}
		} else {
			response.Error(c, http.StatusInternalServerError, "GET_FAILED", err)
			return
		}
	}

	// Update only provided fields
	if req.EmailEnabled != nil {
		prefs.EmailEnabled = *req.EmailEnabled
	}
	if req.SMSEnabled != nil {
		prefs.SMSEnabled = *req.SMSEnabled
	}
	if req.PushEnabled != nil {
		prefs.PushEnabled = *req.PushEnabled
	}
	if req.InAppEnabled != nil {
		prefs.InAppEnabled = *req.InAppEnabled
	}
	if req.SystemEnabled != nil {
		prefs.SystemEnabled = *req.SystemEnabled
	}
	if req.SecurityEnabled != nil {
		prefs.SecurityEnabled = *req.SecurityEnabled
	}
	if req.MarketingEnabled != nil {
		prefs.MarketingEnabled = *req.MarketingEnabled
	}
	if req.ProductEnabled != nil {
		prefs.ProductEnabled = *req.ProductEnabled
	}
	if req.SocialEnabled != nil {
		prefs.SocialEnabled = *req.SocialEnabled
	}
	if req.Timezone != nil {
		prefs.Timezone = *req.Timezone
	}

	if err := h.useCase.UpdateUserPreferences(ctx, userID, prefs); err != nil {
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", err)
		return
	}

	response.Success(c, http.StatusOK, toPreferenceResponse(prefs))
}

// Helper functions to convert entities to response DTOs
func toNotificationResponse(n *entity.Notification) NotificationResponse {
	resp := NotificationResponse{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		Type:      string(n.Type),
		Channel:   string(n.Channel),
		Status:    string(n.Status),
		Template:  n.Template,
		Subject:   n.Subject,
		Content:   n.Content,
		Data:      n.Data,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if n.SentAt != nil {
		sentAt := n.SentAt.Format("2006-01-02T15:04:05Z07:00")
		resp.SentAt = &sentAt
	}
	if n.OpenedAt != nil {
		openedAt := n.OpenedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.OpenedAt = &openedAt
	}
	if n.ClickedAt != nil {
		clickedAt := n.ClickedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ClickedAt = &clickedAt
	}

	return resp
}

func toPreferenceResponse(p *entity.UserPreference) PreferenceResponse {
	resp := PreferenceResponse{
		ID:               p.ID.String(),
		UserID:           p.UserID.String(),
		EmailEnabled:     p.EmailEnabled,
		SMSEnabled:       p.SMSEnabled,
		PushEnabled:      p.PushEnabled,
		InAppEnabled:     p.InAppEnabled,
		SystemEnabled:    p.SystemEnabled,
		SecurityEnabled:  p.SecurityEnabled,
		MarketingEnabled: p.MarketingEnabled,
		ProductEnabled:   p.ProductEnabled,
		SocialEnabled:    p.SocialEnabled,
		Timezone:         p.Timezone,
	}

	if p.QuietHoursStart != nil {
		start := p.QuietHoursStart.Format("15:04")
		resp.QuietHoursStart = &start
	}
	if p.QuietHoursEnd != nil {
		end := p.QuietHoursEnd.Format("15:04")
		resp.QuietHoursEnd = &end
	}

	return resp
}

// Helper to parse time string (HH:MM format)
func parseTimeString(s string) (int, error) {
	return strconv.Atoi(s)
}
