package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/domain/repository"
	"github.com/basilex/promenade/internal/modules/posts/usecase"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockUserPostUseCase is a mock implementation of IUserPostUseCase interface
//  IModule-independent: imports only module types, no core dependencies
type MockUserPostUseCase struct {
	mock.Mock
}

// Compile-time check to ensure MockUserPostUseCase implements IUserPostUseCase interface
var _ usecase.IUserPostUseCase = (*MockUserPostUseCase)(nil)

func (m *MockUserPostUseCase) CreatePost(ctx context.Context, userID uuidv7.UUID, title, content, excerpt string, tags, categories []string) (*entity.UserPost, error) {
	args := m.Called(ctx, userID, title, content, excerpt, tags, categories)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserPost), args.Error(1)
}

func (m *MockUserPostUseCase) GetPost(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserPost), args.Error(1)
}

func (m *MockUserPostUseCase) GetPostBySlug(ctx context.Context, userID uuidv7.UUID, slug string) (*entity.UserPost, error) {
	args := m.Called(ctx, userID, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserPost), args.Error(1)
}

func (m *MockUserPostUseCase) UpdatePost(ctx context.Context, userID, postID uuidv7.UUID, updates map[string]any) (*entity.UserPost, error) {
	args := m.Called(ctx, userID, postID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserPost), args.Error(1)
}

func (m *MockUserPostUseCase) DeletePost(ctx context.Context, userID, postID uuidv7.UUID) error {
	args := m.Called(ctx, userID, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) SoftDeletePost(ctx context.Context, userID, postID uuidv7.UUID) error {
	args := m.Called(ctx, userID, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) RestorePost(ctx context.Context, userID, postID uuidv7.UUID) error {
	args := m.Called(ctx, userID, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) PublishPost(ctx context.Context, userID, postID uuidv7.UUID) error {
	args := m.Called(ctx, userID, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) UnpublishPost(ctx context.Context, userID, postID uuidv7.UUID) error {
	args := m.Called(ctx, userID, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) ArchivePost(ctx context.Context, userID, postID uuidv7.UUID) error {
	args := m.Called(ctx, userID, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) SchedulePost(ctx context.Context, userID, postID uuidv7.UUID, scheduledAt time.Time) error {
	args := m.Called(ctx, userID, postID, scheduledAt)
	return args.Error(0)
}

func (m *MockUserPostUseCase) ToggleFeatured(ctx context.Context, userID, postID uuidv7.UUID) error {
	args := m.Called(ctx, userID, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) ToggleComments(ctx context.Context, userID, postID uuidv7.UUID) error {
	args := m.Called(ctx, userID, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) ViewPost(ctx context.Context, postID uuidv7.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) LikePost(ctx context.Context, postID uuidv7.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) UnlikePost(ctx context.Context, postID uuidv7.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockUserPostUseCase) GetUserPosts(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostUseCase) GetPublishedPosts(ctx context.Context, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostUseCase) GetFeaturedPosts(ctx context.Context, limit int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostUseCase) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostUseCase) GetPostsByTag(ctx context.Context, tag string, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, tag, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostUseCase) ListPosts(ctx context.Context, params repository.ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*entity.UserPost), nil, args.Error(2)
	}
	return args.Get(0).([]*entity.UserPost), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockUserPostUseCase) ProcessScheduledPosts(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
