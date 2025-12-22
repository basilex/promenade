package handler

import (
	"net/http"
	"strconv"

	"github.com/basilex/promenade/internal/modules/posts/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/posts/usecase"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

// CommentHandler handles comment-related HTTP requests
type CommentHandler struct {
	commentUC usecase.CommentUseCase
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(commentUC usecase.CommentUseCase) *CommentHandler {
	return &CommentHandler{
		commentUC: commentUC,
	}
}

// CreateComment godoc
// @Summary Create a new comment on a post
// @Description Create a new comment on a specific post (can be a top-level comment or a reply)
// @Tags comments
// @Accept json
// @Produce json
// @Param postId path string true "Post ID"
// @Param request body dto.CreateCommentRequest true "Comment details"
// @Success 201 {object} response.Response{data=dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /posts/{postId}/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	// Get post ID from path
	postIDStr := c.Param("postId")
	postID, err := uuidv7.Parse(postIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}
	userID := userIDInterface.(uuidv7.UUID)

	// Bind request
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Create comment
	comment, err := h.commentUC.CreateComment(c.Request.Context(), postID, userID, req.Content, req.ParentID)
	if err != nil {
		log.Error("Failed to create comment", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to create comment", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToCommentResponse(comment))
}

// GetComment godoc
// @Summary Get a comment by ID
// @Description Retrieve details of a specific comment
// @Tags comments
// @Produce json
// @Param id path string true "Comment ID"
// @Success 200 {object} response.Response{data=dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /comments/{id} [get]
func (h *CommentHandler) GetComment(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	// Get comment ID from path
	commentIDStr := c.Param("id")
	commentID, err := uuidv7.Parse(commentIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid comment ID", err)
		return
	}

	// Get comment
	comment, err := h.commentUC.GetComment(c.Request.Context(), commentID)
	if err != nil {
		log.Error("Failed to get comment", "error", err)
		response.Error(c, http.StatusNotFound, "comment not found", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCommentResponse(comment))
}

// UpdateComment godoc
// @Summary Update a comment
// @Description Update the content of an existing comment (only by the author)
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param request body dto.UpdateCommentRequest true "Updated comment content"
// @Success 200 {object} response.Response{data=dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /comments/{id} [put]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	// Get comment ID from path
	commentIDStr := c.Param("id")
	commentID, err := uuidv7.Parse(commentIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid comment ID", err)
		return
	}

	// Get user ID from context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}
	userID := userIDInterface.(uuidv7.UUID)

	// Bind request
	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Update comment
	comment, err := h.commentUC.UpdateComment(c.Request.Context(), userID, commentID, req.Content)
	if err != nil {
		log.Error("Failed to update comment", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to update comment", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCommentResponse(comment))
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Soft-delete a comment (only by the author)
// @Tags comments
// @Produce json
// @Param id path string true "Comment ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	// Get comment ID from path
	commentIDStr := c.Param("id")
	commentID, err := uuidv7.Parse(commentIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid comment ID", err)
		return
	}

	// Get user ID from context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}
	userID := userIDInterface.(uuidv7.UUID)

	// Delete comment
	if err := h.commentUC.DeleteComment(c.Request.Context(), userID, commentID); err != nil {
		log.Error("Failed to delete comment", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to delete comment", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

// GetPostComments godoc
// @Summary Get comments for a post
// @Description Retrieve all top-level comments for a specific post with pagination
// @Tags comments
// @Produce json
// @Param postId path string true "Post ID"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} response.PaginatedResponse{items=[]dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/{postId}/comments [get]
func (h *CommentHandler) GetPostComments(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	// Get post ID from path
	postIDStr := c.Param("postId")
	postID, err := uuidv7.Parse(postIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	// Get pagination params
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Get comments
	comments, meta, err := h.commentUC.GetPostComments(c.Request.Context(), postID, limit, offset)
	if err != nil {
		log.Error("Failed to get post comments", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to fetch comments", err)
		return
	}

	page := offset/limit + 1
	paginatedResp := response.NewPaginatedResponse(dto.ToCommentResponses(comments), page, limit, meta.Total)
	c.JSON(http.StatusOK, paginatedResp)
}

// GetCommentReplies godoc
// @Summary Get replies to a comment
// @Description Retrieve all replies to a specific comment with pagination
// @Tags comments
// @Produce json
// @Param id path string true "Comment ID"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} response.PaginatedResponse{items=[]dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /comments/{id}/replies [get]
func (h *CommentHandler) GetCommentReplies(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	// Get comment ID from path
	commentIDStr := c.Param("id")
	commentID, err := uuidv7.Parse(commentIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid comment ID", err)
		return
	}

	// Get pagination params
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Get replies
	replies, meta, err := h.commentUC.GetCommentReplies(c.Request.Context(), commentID, limit, offset)
	if err != nil {
		log.Error("Failed to get comment replies", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to fetch replies", err)
		return
	}

	page := offset/limit + 1
	paginatedResp := response.NewPaginatedResponse(dto.ToCommentResponses(replies), page, limit, meta.Total)
	c.JSON(http.StatusOK, paginatedResp)
}

// GetUserComments godoc
// @Summary Get comments by user
// @Description Retrieve all comments made by a specific user with pagination
// @Tags comments
// @Produce json
// @Param userId path string true "User ID"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} response.PaginatedResponse{items=[]dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{userId}/comments [get]
func (h *CommentHandler) GetUserComments(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	// Get user ID from path
	userIDStr := c.Param("userId")
	userID, err := uuidv7.Parse(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID", err)
		return
	}

	// Get pagination params
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Get comments
	comments, meta, err := h.commentUC.GetUserComments(c.Request.Context(), userID, limit, offset)
	if err != nil {
		log.Error("Failed to get user comments", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to fetch comments", err)
		return
	}

	page := offset/limit + 1
	paginatedResp := response.NewPaginatedResponse(dto.ToCommentResponses(comments), page, limit, meta.Total)
	c.JSON(http.StatusOK, paginatedResp)
}
