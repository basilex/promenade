package entity

import (
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestComment_Validate(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid comment",
			content: "This is a valid comment",
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: true,
			errMsg:  "content is required",
		},
		{
			name:    "content too long",
			content: string(make([]byte, 2001)),
			wantErr: true,
			errMsg:  "content must be at most 2000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment := &Comment{Content: tt.content}
			err := comment.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestComment_IsTopLevel(t *testing.T) {
	tests := []struct {
		name     string
		parentID *uuidv7.UUID
		want     bool
	}{
		{"top-level comment", nil, true},
		{"reply comment", ptrUUID(uuidv7.New()), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment := &Comment{ParentID: tt.parentID}
			assert.Equal(t, tt.want, comment.IsTopLevel())
		})
	}
}

func TestNewComment(t *testing.T) {
	postID := uuidv7.New()
	userID := uuidv7.New()
	content := "Test comment"

	comment, err := NewComment(postID, userID, content, nil)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, userID, comment.UserID)
	assert.Equal(t, content, comment.Content)
}

func ptrUUID(id uuidv7.UUID) *uuidv7.UUID {
	return &id
}
