package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type PostCommentHandler struct {
	commentUC usecase.PostCommentUseCase
}

// NewPostCommentHandler creates a new post comment handler
func NewPostCommentHandler(commentUC usecase.PostCommentUseCase) *PostCommentHandler {
	return &PostCommentHandler{
		commentUC: commentUC,
	}
}

// CreateComment godoc
// @Summary Create a new comment
// @Description Create a comment on a post (or reply to another comment)
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCommentRequest true "Comment data"
// @Success 201 {object} response.Response{data=dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /comments [post]
func (h *PostCommentHandler) CreateComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	postID, err := uuidv7.Parse(req.PostID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post_id", err)
		return
	}

	var parentID *uuidv7.UUID
	if req.ParentID != nil {
		pid, err := uuidv7.Parse(*req.ParentID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid parent_id", err)
			return
		}
		parentID = &pid
	}

	uid := userID.(uuidv7.UUID)
	comment, err := h.commentUC.CreateComment(c.Request.Context(), postID, uid, req.Content, parentID)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCommentData) || errors.Is(err, usecase.ErrCannotReplyToDeleted) {
			response.Error(c, http.StatusBadRequest, err.Error(), err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to create comment", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToCommentResponse(comment))
}

// GetComment godoc
// @Summary Get a comment by ID
// @Description Get comment details
// @Tags comments
// @Produce json
// @Param id path string true "Comment ID (UUID)"
// @Success 200 {object} response.Response{data=dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /comments/{id} [get]
func (h *PostCommentHandler) GetComment(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid comment id", err)
		return
	}

	comment, err := h.commentUC.GetComment(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrCommentNotFound) {
			response.Error(c, http.StatusNotFound, "comment not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get comment", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCommentResponse(comment))
}

// UpdateComment godoc
// @Summary Update a comment
// @Description Update comment content (marks as edited)
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Comment ID (UUID)"
// @Param request body dto.UpdateCommentRequest true "Updated content"
// @Success 200 {object} response.Response{data=dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /comments/{id} [put]
func (h *PostCommentHandler) UpdateComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid comment id", err)
		return
	}

	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	uid := userID.(uuidv7.UUID)
	comment, err := h.commentUC.UpdateComment(c.Request.Context(), id, uid, req.Content)
	if err != nil {
		if errors.Is(err, usecase.ErrCommentNotFound) {
			response.Error(c, http.StatusNotFound, "comment not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorizedComment) {
			response.Error(c, http.StatusForbidden, "unauthorized to modify this comment", err)
			return
		}
		if errors.Is(err, usecase.ErrInvalidCommentData) || errors.Is(err, usecase.ErrCommentDeleted) {
			response.Error(c, http.StatusBadRequest, err.Error(), err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update comment", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCommentResponse(comment))
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Soft delete a comment
// @Tags comments
// @Security BearerAuth
// @Param id path string true "Comment ID (UUID)"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /comments/{id} [delete]
func (h *PostCommentHandler) DeleteComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid comment id", err)
		return
	}

	uid := userID.(uuidv7.UUID)
	if err := h.commentUC.DeleteComment(c.Request.Context(), id, uid); err != nil {
		if errors.Is(err, usecase.ErrCommentNotFound) {
			response.Error(c, http.StatusNotFound, "comment not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorizedComment) {
			response.Error(c, http.StatusForbidden, "unauthorized to delete this comment", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete comment", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetPostComments godoc
// @Summary Get comments for a post
// @Description Get paginated top-level comments for a post
// @Tags comments
// @Produce json
// @Param post_id query string true "Post ID (UUID)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.CommentListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /comments [get]
func (h *PostCommentHandler) GetPostComments(c *gin.Context) {
	postIDStr := c.Query("post_id")
	if postIDStr == "" {
		response.Error(c, http.StatusBadRequest, "post_id query parameter is required", nil)
		return
	}

	postID, err := uuidv7.Parse(postIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post id", err)
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 20
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if _, err := fmt.Sscanf(o, "%d", &offset); err != nil {
			offset = 0
		}
	}

	comments, meta, err := h.commentUC.GetPostComments(c.Request.Context(), postID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get comments", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCommentListResponse(comments, meta))
}

// GetCommentReplies godoc
// @Summary Get replies to a comment
// @Description Get paginated replies to a specific comment
// @Tags comments
// @Produce json
// @Param id path string true "Comment ID (UUID)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.CommentListResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /comments/{id}/replies [get]
func (h *PostCommentHandler) GetCommentReplies(c *gin.Context) {
	commentID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid comment id", err)
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 20
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if _, err := fmt.Sscanf(o, "%d", &offset); err != nil {
			offset = 0
		}
	}

	comments, meta, err := h.commentUC.GetCommentReplies(c.Request.Context(), commentID, limit, offset)
	if err != nil {
		if errors.Is(err, usecase.ErrCommentNotFound) {
			response.Error(c, http.StatusNotFound, "comment not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get replies", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCommentListResponse(comments, meta))
}

// GetUserComments godoc
// @Summary Get user's comments
// @Description Get paginated comments by a specific user
// @Tags comments
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.CommentListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{user_id}/comments [get]
func (h *PostCommentHandler) GetUserComments(c *gin.Context) {
	userID, err := uuidv7.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id", err)
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 20
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if _, err := fmt.Sscanf(o, "%d", &offset); err != nil {
			offset = 0
		}
	}

	comments, meta, err := h.commentUC.GetUserComments(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get comments", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCommentListResponse(comments, meta))
}

// LikeComment godoc
// @Summary Like a comment
// @Description Like a comment
// @Tags comments
// @Security BearerAuth
// @Param id path string true "Comment ID (UUID)"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /comments/{id}/like [post]
func (h *PostCommentHandler) LikeComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	commentID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid comment id", err)
		return
	}

	uid := userID.(uuidv7.UUID)
	if err := h.commentUC.LikeComment(c.Request.Context(), commentID, uid); err != nil {
		if errors.Is(err, usecase.ErrCommentNotFound) {
			response.Error(c, http.StatusNotFound, "comment not found", err)
			return
		}
		if errors.Is(err, usecase.ErrCommentDeleted) {
			response.Error(c, http.StatusBadRequest, "comment is deleted", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to like comment", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UnlikeComment godoc
// @Summary Unlike a comment
// @Description Remove like from a comment
// @Tags comments
// @Security BearerAuth
// @Param id path string true "Comment ID (UUID)"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /comments/{id}/like [delete]
func (h *PostCommentHandler) UnlikeComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	commentID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid comment id", err)
		return
	}

	uid := userID.(uuidv7.UUID)
	if err := h.commentUC.UnlikeComment(c.Request.Context(), commentID, uid); err != nil {
		if errors.Is(err, usecase.ErrCommentNotFound) {
			response.Error(c, http.StatusNotFound, "comment not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to unlike comment", err)
		return
	}

	c.Status(http.StatusNoContent)
}
