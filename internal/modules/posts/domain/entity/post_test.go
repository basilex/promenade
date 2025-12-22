package entity

import (
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestPostStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status PostStatus
		want   bool
	}{
		{"draft status", PostStatusDraft, true},
		{"published status", PostStatusPublished, true},
		{"archived status", PostStatusArchived, true},
		{"scheduled status", PostStatusScheduled, true},
		{"invalid status", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestUserPost_Creation(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test Post",
		Slug:    "test-post",
		Content: "This is test content",
		Status:  PostStatusDraft,
	}

	assert.NotNil(t, post)
	assert.Equal(t, "Test Post", post.Title)
	assert.Equal(t, PostStatusDraft, post.Status)
}
