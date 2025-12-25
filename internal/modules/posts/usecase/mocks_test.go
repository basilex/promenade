package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// mockUserPostRepository implements IUserPostRepository for testing
type mockUserPostRepository struct {
	mock.Mock
}

func (m *mockUserPostRepository) Create(ctx context.Context, post *entity.UserPost) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *mockUserPostRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserPost), args.Error(1)
}

func (m *mockUserPostRepository) GetBySlug(ctx context.Context, userID uuidv7.UUID, slug string) (*entity.UserPost, error) {
	args := m.Called(ctx, userID, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserPost), args.Error(1)
}

func (m *mockUserPostRepository) Update(ctx context.Context, post *entity.UserPost) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *mockUserPostRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserPostRepository) SoftDelete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserPostRepository) Restore(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserPostRepository) List(ctx context.Context, params repository.ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*pagination.Metadata), args.Error(2)
	}
	return args.Get(0).([]*entity.UserPost), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *mockUserPostRepository) GetUserPosts(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *mockUserPostRepository) GetPublishedPosts(ctx context.Context, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *mockUserPostRepository) GetFeaturedPosts(ctx context.Context, limit int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *mockUserPostRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *mockUserPostRepository) GetByTag(ctx context.Context, tag string, limit, offset int) ([]*entity.UserPost, error) {
	args := m.Called(ctx, tag, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *mockUserPostRepository) GetScheduledPosts(ctx context.Context) ([]*entity.UserPost, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPost), args.Error(1)
}

func (m *mockUserPostRepository) IncrementViews(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserPostRepository) IncrementLikes(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserPostRepository) DecrementLikes(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserPostRepository) IncrementComments(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserPostRepository) DecrementComments(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserPostRepository) IncrementShares(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserPostRepository) UpdateStatus(ctx context.Context, id uuidv7.UUID, status entity.PostStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockUserPostRepository) PublishScheduledPost(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// mockCommentRepository implements ICommentRepository for testing
type mockCommentRepository struct {
	mock.Mock
}

func (m *mockCommentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *mockCommentRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Comment), args.Error(1)
}

func (m *mockCommentRepository) Update(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *mockCommentRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCommentRepository) GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	args := m.Called(ctx, postID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*pagination.Metadata), args.Error(2)
	}
	return args.Get(0).([]*entity.Comment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *mockCommentRepository) GetReplies(ctx context.Context, parentID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	args := m.Called(ctx, parentID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*pagination.Metadata), args.Error(2)
	}
	return args.Get(0).([]*entity.Comment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *mockCommentRepository) GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*pagination.Metadata), args.Error(2)
	}
	return args.Get(0).([]*entity.Comment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *mockCommentRepository) CountPostComments(ctx context.Context, postID uuidv7.UUID) (int, error) {
	args := m.Called(ctx, postID)
	return args.Int(0), args.Error(1)
}

func (m *mockCommentRepository) IncrementRepliesCount(ctx context.Context, parentID uuidv7.UUID) error {
	args := m.Called(ctx, parentID)
	return args.Error(0)
}

func (m *mockCommentRepository) DecrementRepliesCount(ctx context.Context, parentID uuidv7.UUID) error {
	args := m.Called(ctx, parentID)
	return args.Error(0)
}
