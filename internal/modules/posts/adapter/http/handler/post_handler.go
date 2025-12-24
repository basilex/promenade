package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/modules/posts/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/domain/repository"
	"github.com/basilex/promenade/internal/modules/posts/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type UserPostHandler struct {
	postUC usecase.IUserPostUseCase
}

func NewUserPostHandler(postUC usecase.IUserPostUseCase) *UserPostHandler {
	return &UserPostHandler{
		postUC: postUC,
	}
}

// CreatePost godoc
// @Summary Create a new post
// @Description Create a new blog post
// @Tags posts
// @Accept json
// @Produce json
// @Param request body dto.CreatePostRequest true "Create post request"
// @Success 201 {object} response.Response{data=dto.PostResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /posts [post]
func (h *UserPostHandler) CreatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	post, err := h.postUC.CreatePost(c.Request.Context(), userID.(uuidv7.UUID), req.Title, req.Content, req.Excerpt, req.Tags, req.Categories)
	if err != nil {
		if err == usecase.ErrSlugAlreadyExists {
			response.Error(c, http.StatusConflict, "post with this slug already exists", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to create post", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToPostResponse(post))
}

// GetPost godoc
// @Summary Get post by ID
// @Description Get a post by its ID
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/{id} [get]
func (h *UserPostHandler) GetPost(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	post, err := h.postUC.GetPost(c.Request.Context(), postID)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "post not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get post", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPostResponse(post))
}

// UpdatePost godoc
// @Summary Update post
// @Description Update an existing post
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body dto.UpdatePostRequest true "Update post request"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /posts/{id} [put]
func (h *UserPostHandler) UpdatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	idStr := c.Param("id")
	postID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	updates := req.ToUpdateMap()

	post, err := h.postUC.UpdatePost(c.Request.Context(), userID.(uuidv7.UUID), postID, updates)
	if err != nil {
		if err == usecase.ErrUnauthorized {
			response.Error(c, http.StatusForbidden, "not authorized to update this post", err)
			return
		}
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "post not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update post", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPostResponse(post))
}

// DeletePost godoc
// @Summary Delete post
// @Description Soft delete a post
// @Tags posts
// @Param id path string true "Post ID"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /posts/{id} [delete]
func (h *UserPostHandler) DeletePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	idStr := c.Param("id")
	postID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	if err := h.postUC.SoftDeletePost(c.Request.Context(), userID.(uuidv7.UUID), postID); err != nil {
		if err == usecase.ErrUnauthorized {
			response.Error(c, http.StatusForbidden, "not authorized to delete this post", err)
			return
		}
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "post not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete post", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PublishPost godoc
// @Summary Publish post
// @Description Publish a draft post immediately or schedule for later
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body dto.PublishPostRequest false "Publish options"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /posts/{id}/publish [post]
func (h *UserPostHandler) PublishPost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	idStr := c.Param("id")
	postID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	var req dto.PublishPostRequest
	_ = c.ShouldBindJSON(&req)

	if req.ScheduledAt != nil {
		// Schedule for later
		err = h.postUC.SchedulePost(c.Request.Context(), userID.(uuidv7.UUID), postID, *req.ScheduledAt)
	} else {
		// Publish immediately
		err = h.postUC.PublishPost(c.Request.Context(), userID.(uuidv7.UUID), postID)
	}

	if err != nil {
		if err == usecase.ErrUnauthorized {
			response.Error(c, http.StatusForbidden, "not authorized to publish this post", err)
			return
		}
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "post not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to publish post", err)
		return
	}

	// Get updated post
	post, _ := h.postUC.GetPost(c.Request.Context(), postID)
	response.Success(c, http.StatusOK, dto.ToPostResponse(post))
}

// UnpublishPost godoc
// @Summary Unpublish post
// @Description Unpublish a published post
// @Tags posts
// @Param id path string true "Post ID"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /posts/{id}/unpublish [post]
func (h *UserPostHandler) UnpublishPost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	idStr := c.Param("id")
	postID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	if err := h.postUC.UnpublishPost(c.Request.Context(), userID.(uuidv7.UUID), postID); err != nil {
		if err == usecase.ErrUnauthorized {
			response.Error(c, http.StatusForbidden, "not authorized to unpublish this post", err)
			return
		}
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "post not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to unpublish post", err)
		return
	}

	// Get updated post
	post, _ := h.postUC.GetPost(c.Request.Context(), postID)
	response.Success(c, http.StatusOK, dto.ToPostResponse(post))
}

// GetUserPosts godoc
// @Summary Get user's posts
// @Description Get all posts created by a specific user
// @Tags posts
// @Produce json
// @Param user_id path string true "User ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.PostListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{user_id}/posts [get]
func (h *UserPostHandler) GetUserPosts(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuidv7.Parse(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID", err)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	posts, err := h.postUC.GetUserPosts(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get user posts", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPostListResponse(posts, nil))
}

// GetPublishedPosts godoc
// @Summary Get published posts
// @Description Get all published posts
// @Tags posts
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.PostListResponse}
// @Failure 500 {object} response.Response
// @Router /posts/published [get]
func (h *UserPostHandler) GetPublishedPosts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	posts, err := h.postUC.GetPublishedPosts(c.Request.Context(), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get published posts", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPostListResponse(posts, nil))
}

// GetFeaturedPosts godoc
// @Summary Get featured posts
// @Description Get featured posts
// @Tags posts
// @Produce json
// @Param limit query int false "Limit" default(5)
// @Success 200 {object} response.Response{data=dto.PostListResponse}
// @Failure 500 {object} response.Response
// @Router /posts/featured [get]
func (h *UserPostHandler) GetFeaturedPosts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	posts, err := h.postUC.GetFeaturedPosts(c.Request.Context(), limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get featured posts", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPostListResponse(posts, nil))
}

// SearchPosts godoc
// @Summary Search posts
// @Description Search posts by keyword
// @Tags posts
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.PostListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/search [get]
func (h *UserPostHandler) SearchPosts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "search query is required", nil)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	posts, err := h.postUC.SearchPosts(c.Request.Context(), query, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to search posts", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPostListResponse(posts, nil))
}

// GetPostsByTag godoc
// @Summary Get posts by tag
// @Description Get posts filtered by tag
// @Tags posts
// @Produce json
// @Param tag path string true "Tag name"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.PostListResponse}
// @Failure 500 {object} response.Response
// @Router /posts/tags/{tag} [get]
func (h *UserPostHandler) GetPostsByTag(c *gin.Context) {
	tag := c.Param("tag")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	posts, err := h.postUC.GetPostsByTag(c.Request.Context(), tag, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get posts by tag", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPostListResponse(posts, nil))
}

// ListPosts godoc
// @Summary List posts with filters
// @Description List posts with advanced filtering and pagination
// @Tags posts
// @Accept json
// @Produce json
// @Param request body dto.ListPostsRequest true "List posts request"
// @Success 200 {object} response.Response{data=dto.PostListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/list [post]
func (h *UserPostHandler) ListPosts(c *gin.Context) {
	var req dto.ListPostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	params := repository.ListPostsParams{
		Limit:      req.Limit,
		Offset:     req.Offset,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
		UserID:     req.UserID,
		Status:     req.Status,
		IsPublic:   req.IsPublic,
		IsFeatured: req.IsFeatured,
	}

	posts, meta, err := h.postUC.ListPosts(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list posts", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPostListResponse(posts, meta))
}

// ViewPost godoc
// @Summary View post (increment views)
// @Description Increment view count for a post
// @Tags posts
// @Param id path string true "Post ID"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/{id}/view [post]
func (h *UserPostHandler) ViewPost(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	if err := h.postUC.ViewPost(c.Request.Context(), postID); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to record view", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// LikePost godoc
// @Summary Like post
// @Description Increment like count for a post
// @Tags posts
// @Param id path string true "Post ID"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/{id}/like [post]
func (h *UserPostHandler) LikePost(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	if err := h.postUC.LikePost(c.Request.Context(), postID); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to like post", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UnlikePost godoc
// @Summary Unlike post
// @Description Decrement like count for a post
// @Tags posts
// @Param id path string true "Post ID"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/{id}/unlike [post]
func (h *UserPostHandler) UnlikePost(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	if err := h.postUC.UnlikePost(c.Request.Context(), postID); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to unlike post", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ToggleFeatured godoc
// @Summary Toggle featured status
// @Description Toggle the featured flag for a post
// @Tags posts
// @Param id path string true "Post ID"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /posts/{id}/featured [post]
func (h *UserPostHandler) ToggleFeatured(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	idStr := c.Param("id")
	postID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	if err := h.postUC.ToggleFeatured(c.Request.Context(), userID.(uuidv7.UUID), postID); err != nil {
		if err == usecase.ErrUnauthorized {
			response.Error(c, http.StatusForbidden, "not authorized", err)
			return
		}
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "post not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to toggle featured", err)
		return
	}

	post, _ := h.postUC.GetPost(c.Request.Context(), postID)
	response.Success(c, http.StatusOK, dto.ToPostResponse(post))
}

// ToggleComments godoc
// @Summary Toggle comments
// @Description Toggle comments enabled/disabled for a post
// @Tags posts
// @Param id path string true "Post ID"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /posts/{id}/comments [post]
func (h *UserPostHandler) ToggleComments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	idStr := c.Param("id")
	postID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post ID", err)
		return
	}

	if err := h.postUC.ToggleComments(c.Request.Context(), userID.(uuidv7.UUID), postID); err != nil {
		if err == usecase.ErrUnauthorized {
			response.Error(c, http.StatusForbidden, "not authorized", err)
			return
		}
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "post not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to toggle comments", err)
		return
	}

	post, _ := h.postUC.GetPost(c.Request.Context(), postID)
	response.Success(c, http.StatusOK, dto.ToPostResponse(post))
}
