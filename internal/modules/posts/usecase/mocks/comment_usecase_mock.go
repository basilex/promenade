package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/usecase"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockCommentUseCase is a mock implementation of ICommentUseCase interface
//  IModule-independent: imports only module types, no core dependencies
type MockCommentUseCase struct {
	mock.Mock
}

// Compile-time check to ensure MockCommentUseCase implements ICommentUseCase interface
var _ usecase.ICommentUseCase = (*MockCommentUseCase)(nil)

func (m *MockCommentUseCase) CreateComment(ctx context.Context, postID, userID uuidv7.UUID, content string, parentID *uuidv7.UUID) (*entity.Comment, error) {
	args := m.Called(ctx, postID, userID, content, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Comment), args.Error(1)
}

func (m *MockCommentUseCase) GetComment(ctx context.Context, id uuidv7.UUID) (*entity.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Comment), args.Error(1)
}

func (m *MockCommentUseCase) UpdateComment(ctx context.Context, userID, commentID uuidv7.UUID, content string) (*entity.Comment, error) {
	args := m.Called(ctx, userID, commentID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Comment), args.Error(1)
}

func (m *MockCommentUseCase) DeleteComment(ctx context.Context, userID, commentID uuidv7.UUID) error {
	args := m.Called(ctx, userID, commentID)
	return args.Error(0)
}

func (m *MockCommentUseCase) GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	args := m.Called(ctx, postID, limit, offset)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*entity.Comment), nil, args.Error(2)
	}
	return args.Get(0).([]*entity.Comment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockCommentUseCase) GetCommentReplies(ctx context.Context, commentID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	args := m.Called(ctx, commentID, limit, offset)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*entity.Comment), nil, args.Error(2)
	}
	return args.Get(0).([]*entity.Comment), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockCommentUseCase) GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*entity.Comment), nil, args.Error(2)
	}
	return args.Get(0).([]*entity.Comment), args.Get(1).(*pagination.Metadata), args.Error(2)
}
