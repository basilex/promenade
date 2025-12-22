package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockCommentRepository is a mock implementation of CommentRepository
//  Module-independent: imports only module types, no core dependencies
type MockCommentRepository struct {
	mock.Mock
}

// Compile-time check to ensure MockCommentRepository implements CommentRepository interface
var _ repository.CommentRepository = (*MockCommentRepository)(nil)

func (m *MockCommentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Comment), args.Error(1)
}

func (m *MockCommentRepository) Update(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentRepository) GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	args := m.Called(ctx, postID, limit, offset)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*entity.Comment), nil, args.Error(2)
	}
	return args.Get(0).([]*entity.Comment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockCommentRepository) GetReplies(ctx context.Context, parentID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	args := m.Called(ctx, parentID, limit, offset)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*entity.Comment), nil, args.Error(2)
	}
	return args.Get(0).([]*entity.Comment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockCommentRepository) GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*entity.Comment), nil, args.Error(2)
	}
	return args.Get(0).([]*entity.Comment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockCommentRepository) CountPostComments(ctx context.Context, postID uuidv7.UUID) (int, error) {
	args := m.Called(ctx, postID)
	return args.Int(0), args.Error(1)
}

func (m *MockCommentRepository) IncrementRepliesCount(ctx context.Context, parentID uuidv7.UUID) error {
	args := m.Called(ctx, parentID)
	return args.Error(0)
}

func (m *MockCommentRepository) DecrementRepliesCount(ctx context.Context, parentID uuidv7.UUID) error {
	args := m.Called(ctx, parentID)
	return args.Error(0)
}
