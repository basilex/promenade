package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestPostStatus_Validate(t *testing.T) {
	tests := []struct {
		name    string
		status  PostStatus
		wantErr bool
	}{
		{"Valid Draft", PostStatusDraft, false},
		{"Valid Published", PostStatusPublished, false},
		{"Valid Archived", PostStatusArchived, false},
		{"Valid Scheduled", PostStatusScheduled, false},
		{"Invalid Status", PostStatus("invalid"), true},
		{"Empty Status", PostStatus(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.status.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewUserPost(t *testing.T) {
	userID := uuidv7.New()

	tests := []struct {
		name    string
		userID  uuidv7.UUID
		title   string
		slug    string
		content string
		wantErr bool
	}{
		{
			name:    "Valid Post",
			userID:  userID,
			title:   "My First Post",
			slug:    "my-first-post",
			content: "This is my first blog post content.",
			wantErr: false,
		},
		{
			name:    "Empty Title",
			userID:  userID,
			title:   "",
			slug:    "slug",
			content: "content",
			wantErr: true,
		},
		{
			name:    "Empty Slug",
			userID:  userID,
			title:   "Title",
			slug:    "",
			content: "content",
			wantErr: true,
		},
		{
			name:    "Empty Content",
			userID:  userID,
			title:   "Title",
			slug:    "slug",
			content: "",
			wantErr: true,
		},
		{
			name:    "Title Too Long",
			userID:  userID,
			title:   string(make([]byte, 201)),
			slug:    "slug",
			content: "content",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := NewUserPost(tt.userID, tt.title, tt.slug, tt.content)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, post)
				assert.Equal(t, tt.userID, post.UserID)
				assert.Equal(t, tt.title, post.Title)
				assert.Equal(t, tt.slug, post.Slug)
				assert.Equal(t, tt.content, post.Content)
				assert.Equal(t, PostStatusDraft, post.Status)
				assert.True(t, post.IsPublic)
				assert.False(t, post.IsFeatured)
				assert.True(t, post.IsCommentsEnabled)
			}
		})
	}
}

func TestUserPost_Publish(t *testing.T) {
	userID := uuidv7.New()
	post, err := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	assert.NoError(t, err)
	assert.NotNil(t, post)

	assert.Equal(t, PostStatusDraft, post.Status)
	assert.Nil(t, post.PublishedAt)

	post.Publish()

	assert.Equal(t, PostStatusPublished, post.Status)
	assert.NotNil(t, post.PublishedAt)
	assert.Nil(t, post.ScheduledAt)
}

func TestUserPost_Unpublish(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)
	post.Publish()

	assert.Equal(t, PostStatusPublished, post.Status)

	post.Unpublish()

	assert.Equal(t, PostStatusDraft, post.Status)
}

func TestUserPost_Archive(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	post.Archive()

	assert.Equal(t, PostStatusArchived, post.Status)
}

func TestUserPost_Schedule(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	futureTime := time.Now().Add(24 * time.Hour)
	post.Schedule(futureTime)

	assert.Equal(t, PostStatusScheduled, post.Status)
	assert.NotNil(t, post.ScheduledAt)
	assert.Equal(t, futureTime, *post.ScheduledAt)
}

func TestUserPost_ToggleFeatured(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	assert.False(t, post.IsFeatured)

	post.ToggleFeatured()
	assert.True(t, post.IsFeatured)

	post.ToggleFeatured()
	assert.False(t, post.IsFeatured)
}

func TestUserPost_ToggleComments(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	assert.True(t, post.IsCommentsEnabled)

	post.ToggleComments()
	assert.False(t, post.IsCommentsEnabled)

	post.ToggleComments()
	assert.True(t, post.IsCommentsEnabled)
}

func TestUserPost_IncrementViews(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	assert.Equal(t, 0, post.ViewCount)

	post.IncrementViews()
	assert.Equal(t, 1, post.ViewCount)

	post.IncrementViews()
	assert.Equal(t, 2, post.ViewCount)
}

func TestUserPost_IncrementLikes(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	assert.Equal(t, 0, post.LikeCount)

	post.IncrementLikes()
	assert.Equal(t, 1, post.LikeCount)
}

func TestUserPost_DecrementLikes(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)
	post.IncrementLikes()
	post.IncrementLikes()

	assert.Equal(t, 2, post.LikeCount)

	post.DecrementLikes()
	assert.Equal(t, 1, post.LikeCount)

	// Should not go below 0
	post.DecrementLikes()
	post.DecrementLikes()
	assert.Equal(t, 0, post.LikeCount)
}

func TestUserPost_IncrementComments(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	post.IncrementComments()
	assert.Equal(t, 1, post.CommentCount)
}

func TestUserPost_DecrementComments(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)
	post.IncrementComments()
	post.IncrementComments()

	post.DecrementComments()
	assert.Equal(t, 1, post.CommentCount)

	// Should not go below 0
	post.DecrementComments()
	post.DecrementComments()
	assert.Equal(t, 0, post.CommentCount)
}

func TestUserPost_IncrementShares(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	post.IncrementShares()
	assert.Equal(t, 1, post.ShareCount)
}

func TestUserPost_SoftDelete(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	assert.Nil(t, post.DeletedAt)

	post.SoftDelete()

	assert.NotNil(t, post.DeletedAt)
}

func TestUserPost_Restore(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)
	post.SoftDelete()

	assert.NotNil(t, post.DeletedAt)

	post.Restore()

	assert.Nil(t, post.DeletedAt)
}

func TestUserPost_IsPublished(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	assert.False(t, post.IsPublished())

	post.Publish()
	assert.True(t, post.IsPublished())

	post.Unpublish()
	assert.False(t, post.IsPublished())
}

func TestUserPost_IsScheduled(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	assert.False(t, post.IsScheduled())

	post.Schedule(time.Now().Add(24 * time.Hour))
	assert.True(t, post.IsScheduled())

	post.Publish()
	assert.False(t, post.IsScheduled())
}

func TestUserPost_ShouldPublishNow(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Test Post", "test-post", "This is test content for the post")
	require.NotNil(t, post)

	// Not scheduled
	assert.False(t, post.ShouldPublishNow())

	// Scheduled in future
	post.Schedule(time.Now().Add(24 * time.Hour))
	assert.False(t, post.ShouldPublishNow())

	// Create new post for past schedule test
	post2, _ := NewUserPost(userID, "Test Post 2", "test-post-2", "This is test content for the post")
	require.NotNil(t, post2)

	// Scheduled in past (set directly since Schedule() rejects past times)
	pastTime := time.Now().Add(-1 * time.Hour)
	post2.Status = PostStatusScheduled
	post2.ScheduledAt = &pastTime

	assert.True(t, post2.ShouldPublishNow())
}

func TestUserPost_GenerateSlug(t *testing.T) {
	userID := uuidv7.New()
	post, _ := NewUserPost(userID, "Hello World", "hello-world", "This is test content for the post")
	require.NotNil(t, post)

	slug := post.GenerateSlug()
	assert.Equal(t, "hello-world", slug)
}

func TestUserPost_Validate(t *testing.T) {
	userID := uuidv7.New()

	tests := []struct {
		name    string
		post    *UserPost
		wantErr bool
	}{
		{
			name: "Valid Post",
			post: &UserPost{
				ID:      uuidv7.New(),
				UserID:  userID,
				Title:   "Valid Title",
				Slug:    "valid-slug",
				Content: "Valid content",
				Status:  PostStatusDraft,
			},
			wantErr: false,
		},
		{
			name: "Missing Title",
			post: &UserPost{
				ID:      uuidv7.New(),
				UserID:  userID,
				Title:   "",
				Slug:    "slug",
				Content: "content",
				Status:  PostStatusDraft,
			},
			wantErr: true,
		},
		{
			name: "Missing Slug",
			post: &UserPost{
				ID:      uuidv7.New(),
				UserID:  userID,
				Title:   "Title",
				Slug:    "",
				Content: "content",
				Status:  PostStatusDraft,
			},
			wantErr: true,
		},
		{
			name: "Missing Content",
			post: &UserPost{
				ID:      uuidv7.New(),
				UserID:  userID,
				Title:   "Title",
				Slug:    "slug",
				Content: "",
				Status:  PostStatusDraft,
			},
			wantErr: true,
		},
		{
			name: "Invalid Status",
			post: &UserPost{
				ID:      uuidv7.New(),
				UserID:  userID,
				Title:   "Title",
				Slug:    "slug",
				Content: "content",
				Status:  PostStatus("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.post.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFeaturedImage_MarshalUnmarshal(t *testing.T) {
	original := &FeaturedImage{
		URL:          "https://example.com/image.jpg",
		Alt:          "Test Image",
		Width:        800,
		Height:       600,
		ThumbnailURL: "https://example.com/thumb.jpg",
	}

	// Marshal
	data, err := MarshalFeaturedImage(original)
	assert.NoError(t, err)
	assert.NotNil(t, data)

	// Unmarshal
	unmarshaled, err := UnmarshalFeaturedImage(data)
	assert.NoError(t, err)
	assert.NotNil(t, unmarshaled)
	assert.Equal(t, original.URL, unmarshaled.URL)
	assert.Equal(t, original.Alt, unmarshaled.Alt)
	assert.Equal(t, original.Width, unmarshaled.Width)
	assert.Equal(t, original.Height, unmarshaled.Height)
	assert.Equal(t, original.ThumbnailURL, unmarshaled.ThumbnailURL)
}

func TestFeaturedImage_MarshalNil(t *testing.T) {
	data, err := MarshalFeaturedImage(nil)
	assert.NoError(t, err)
	assert.Nil(t, data)
}

func TestFeaturedImage_UnmarshalNil(t *testing.T) {
	img, err := UnmarshalFeaturedImage(nil)
	assert.NoError(t, err)
	assert.Nil(t, img)
}
