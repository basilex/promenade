package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockPostCommentRepository is a mock implementation of repository.PostCommentRepository
type MockPostCommentRepository struct {
	mock.Mock
}

func (m *MockPostCommentRepository) Create(ctx context.Context, comment *entity.PostComment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockPostCommentRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.PostComment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PostComment), args.Error(1)
}

func (m *MockPostCommentRepository) Update(ctx context.Context, comment *entity.PostComment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockPostCommentRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostCommentRepository) SoftDelete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostCommentRepository) Restore(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostCommentRepository) GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error) {
	args := m.Called(ctx, postID, limit, offset)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entity.PostComment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockPostCommentRepository) GetCommentReplies(ctx context.Context, parentID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error) {
	args := m.Called(ctx, parentID, limit, offset)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entity.PostComment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockPostCommentRepository) GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entity.PostComment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockPostCommentRepository) CountPostComments(ctx context.Context, postID uuidv7.UUID) (int, error) {
	args := m.Called(ctx, postID)
	return args.Int(0), args.Error(1)
}

func (m *MockPostCommentRepository) CountUserComments(ctx context.Context, userID uuidv7.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockPostCommentRepository) IncrementLikes(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostCommentRepository) DecrementLikes(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostCommentRepository) IncrementReplies(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostCommentRepository) DecrementReplies(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// AddLike mocks the AddLike method
func (m *MockPostCommentRepository) AddLike(ctx context.Context, commentID, userID uuidv7.UUID) error {
	args := m.Called(ctx, commentID, userID)
	return args.Error(0)
}

// RemoveLike mocks the RemoveLike method
func (m *MockPostCommentRepository) RemoveLike(ctx context.Context, commentID, userID uuidv7.UUID) error {
	args := m.Called(ctx, commentID, userID)
	return args.Error(0)
}

// HasUserLiked mocks the HasUserLiked method
func (m *MockPostCommentRepository) HasUserLiked(ctx context.Context, commentID, userID uuidv7.UUID) (bool, error) {
	args := m.Called(ctx, commentID, userID)
	return args.Bool(0), args.Error(1)
}

// GetCommentLikers mocks the GetCommentLikers method
func (m *MockPostCommentRepository) GetCommentLikers(ctx context.Context, commentID uuidv7.UUID, limit, offset int) ([]uuidv7.UUID, error) {
	args := m.Called(ctx, commentID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuidv7.UUID), args.Error(1)
}
