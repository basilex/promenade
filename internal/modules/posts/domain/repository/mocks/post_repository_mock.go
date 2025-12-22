package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockUserPostRepository is a mock implementation of repository.UserPostRepository
//  This mock is fully independent and uses only module-internal types
type MockUserPostRepository struct {
	mock.Mock
}

// Ensure MockUserPostRepository implements repository.UserPostRepository
var _ repository.UserPostRepository = (*MockUserPostRepository)(nil)

func (m *MockUserPostRepository) Create(ctx context.Context, post *entity.UserPost) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockUserPostRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserPost), args.Error(1)
}

func (m *MockUserPostRepository) GetBySlug(ctx context.Context, userID uuidv7.UUID, slug string) (*entity.UserPost, error) {
	args := m.Called(ctx, userID, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserPost), args.Error(1)
}

func (m *MockUserPostRepository) Update(ctx context.Context, post *entity.UserPost) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockUserPostRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserPostRepository) SoftDelete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserPostRepository) Restore(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserPostRepository) List(ctx context.Context, params repository.ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entity.UserPost), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockUserPostRepository) GetUserPosts(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostRepository) GetPublishedPosts(ctx context.Context, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostRepository) GetFeaturedPosts(ctx context.Context, limit int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostRepository) GetByTag(ctx context.Context, tag string, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, tag, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostRepository) GetScheduledPosts(ctx context.Context) ([]*entity.UserPost, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *MockUserPostRepository) IncrementViews(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserPostRepository) IncrementLikes(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserPostRepository) DecrementLikes(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserPostRepository) IncrementComments(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserPostRepository) DecrementComments(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserPostRepository) IncrementShares(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserPostRepository) UpdateStatus(ctx context.Context, id uuidv7.UUID, status entity.PostStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockUserPostRepository) PublishScheduledPost(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
