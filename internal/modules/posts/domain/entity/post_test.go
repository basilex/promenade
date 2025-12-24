package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/pkg/uuidv7"
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

func TestNewUserPost(t *testing.T) {
	userID := uuidv7.New()
	post, err := NewUserPost(userID, "My First Post", "my-first-post", "This is the content of my first post.")

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, userID, post.UserID)
	assert.Equal(t, "My First Post", post.Title)
	assert.Equal(t, "my-first-post", post.Slug)
	assert.Equal(t, PostStatusDraft, post.Status)
	assert.True(t, post.IsPublic)
	assert.True(t, post.IsCommentsEnabled)
}

func TestUserPost_Validate(t *testing.T) {
	userID := uuidv7.New()

	tests := []struct {
		name    string
		post    *UserPost
		wantErr bool
	}{
		{
			name: "valid post",
			post: &UserPost{
				ID:      uuidv7.New(),
				UserID:  userID,
				Title:   "Valid Post",
				Slug:    "valid-post",
				Content: "This is valid content",
				Status:  PostStatusDraft,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			post: &UserPost{
				ID:      uuidv7.New(),
				Title:   "Test",
				Slug:    "test",
				Content: "content",
				Status:  PostStatusDraft,
			},
			wantErr: true,
		},
		{
			name: "empty title",
			post: &UserPost{
				ID:      uuidv7.New(),
				UserID:  userID,
				Slug:    "test",
				Content: "content",
				Status:  PostStatusDraft,
			},
			wantErr: true,
		},
		{
			name: "short content",
			post: &UserPost{
				ID:      uuidv7.New(),
				UserID:  userID,
				Title:   "Test",
				Slug:    "test",
				Content: "short",
				Status:  PostStatusDraft,
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

func TestUserPost_Publish(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test",
		Slug:    "test",
		Content: "content here",
		Status:  PostStatusDraft,
	}

	post.Publish()
	assert.Equal(t, PostStatusPublished, post.Status)
	assert.NotNil(t, post.PublishedAt)
	assert.Nil(t, post.ScheduledAt)
	assert.True(t, post.IsPublished())
}

func TestUserPost_Unpublish(t *testing.T) {
	now := time.Now()
	post := &UserPost{
		ID:          uuidv7.New(),
		UserID:      uuidv7.New(),
		Title:       "Test",
		Slug:        "test",
		Content:     "content",
		Status:      PostStatusPublished,
		PublishedAt: &now,
	}

	post.Unpublish()
	assert.Equal(t, PostStatusDraft, post.Status)
	assert.Nil(t, post.PublishedAt)
}

func TestUserPost_Archive(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test",
		Slug:    "test",
		Content: "content",
		Status:  PostStatusPublished,
	}

	post.Archive()
	assert.Equal(t, PostStatusArchived, post.Status)
}

func TestUserPost_Schedule(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test",
		Slug:    "test",
		Content: "content",
		Status:  PostStatusDraft,
	}

	future := time.Now().Add(24 * time.Hour)
	err := post.Schedule(future)
	assert.NoError(t, err)
	assert.Equal(t, PostStatusScheduled, post.Status)
	assert.NotNil(t, post.ScheduledAt)
	assert.True(t, post.IsScheduled())
}

func TestUserPost_SchedulePastTime(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test",
		Slug:    "test",
		Content: "content",
		Status:  PostStatusDraft,
	}

	past := time.Now().Add(-1 * time.Hour)
	err := post.Schedule(past)
	assert.Error(t, err)
}

func TestUserPost_IncrementCounters(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test",
		Slug:    "test",
		Content: "content",
		Status:  PostStatusDraft,
	}

	assert.Equal(t, 0, post.ViewCount)
	post.IncrementViews()
	assert.Equal(t, 1, post.ViewCount)

	assert.Equal(t, 0, post.LikeCount)
	post.IncrementLikes()
	assert.Equal(t, 1, post.LikeCount)
	post.DecrementLikes()
	assert.Equal(t, 0, post.LikeCount)

	assert.Equal(t, 0, post.CommentCount)
	post.IncrementComments()
	assert.Equal(t, 1, post.CommentCount)
	post.DecrementComments()
	assert.Equal(t, 0, post.CommentCount)

	assert.Equal(t, 0, post.ShareCount)
	post.IncrementShares()
	assert.Equal(t, 1, post.ShareCount)
}

func TestUserPost_SoftDelete(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test",
		Slug:    "test",
		Content: "content",
		Status:  PostStatusDraft,
	}

	assert.False(t, post.IsDeleted())
	post.SoftDelete()
	assert.True(t, post.IsDeleted())
	assert.NotNil(t, post.DeletedAt)

	post.Restore()
	assert.False(t, post.IsDeleted())
	assert.Nil(t, post.DeletedAt)
}

func TestUserPost_ToggleFeatured(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test",
		Slug:    "test",
		Content: "content",
		Status:  PostStatusDraft,
	}

	assert.False(t, post.IsFeatured)
	post.ToggleFeatured()
	assert.True(t, post.IsFeatured)
	post.ToggleFeatured()
	assert.False(t, post.IsFeatured)
}

func TestUserPost_ToggleComments(t *testing.T) {
	post := &UserPost{
		ID:                uuidv7.New(),
		UserID:            uuidv7.New(),
		Title:             "Test",
		Slug:              "test",
		Content:           "content",
		Status:            PostStatusDraft,
		IsCommentsEnabled: true,
	}

	assert.True(t, post.IsCommentsEnabled)
	post.ToggleComments()
	assert.False(t, post.IsCommentsEnabled)
}

func TestUserPost_IsOwnedBy(t *testing.T) {
	userID := uuidv7.New()
	otherID := uuidv7.New()

	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  userID,
		Title:   "Test",
		Slug:    "test",
		Content: "content",
		Status:  PostStatusDraft,
	}

	assert.True(t, post.IsOwnedBy(userID))
	assert.False(t, post.IsOwnedBy(otherID))
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"simple", "Hello World", "hello-world"},
		{"special chars", "Hello, World!", "hello-world"},
		{"multiple spaces", "Hello   World", "hello-world"},
		{"hyphens", "Hello-World-Test", "hello-world-test"},
		{"numbers", "Test 123 Post", "test-123-post"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GenerateSlug(tt.title))
		})
	}
}

func TestUserPost_GenerateSlug(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "My Great Post Title",
		Content: "content",
	}

	slug := post.GenerateSlug()
	assert.Equal(t, "my-great-post-title", slug)
}

func TestPostStatus_Validate(t *testing.T) {
	tests := []struct {
		name    string
		status  PostStatus
		wantErr bool
	}{
		{"valid draft", PostStatusDraft, false},
		{"valid published", PostStatusPublished, false},
		{"invalid", "invalid", true},
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

func TestUserPost_ShouldPublishNow(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	post := &UserPost{
		ID:          uuidv7.New(),
		UserID:      uuidv7.New(),
		Title:       "Test",
		Slug:        "test",
		Content:     "content",
		Status:      PostStatusScheduled,
		ScheduledAt: &past,
	}

	assert.True(t, post.ShouldPublishNow())
}

func TestMarshalUnmarshalFeaturedImage(t *testing.T) {
	img := &FeaturedImage{
		URL:    "https://example.com/image.jpg",
		Alt:    "Test Image",
		Width:  800,
		Height: 600,
	}

	data, err := MarshalFeaturedImage(img)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	unmarshaled, err := UnmarshalFeaturedImage(data)
	assert.NoError(t, err)
	assert.Equal(t, img.URL, unmarshaled.URL)
	assert.Equal(t, img.Alt, unmarshaled.Alt)
}

func TestMarshalFeaturedImageNil(t *testing.T) {
	data, err := MarshalFeaturedImage(nil)
	assert.NoError(t, err)
	assert.Nil(t, data)
}

func TestUnmarshalFeaturedImageEmpty(t *testing.T) {
	img, err := UnmarshalFeaturedImage([]byte{})
	assert.NoError(t, err)
	assert.Nil(t, img)
}

func TestUserPost_ValidateTitle(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "ab",
		Slug:    "test",
		Content: "content here",
		Status:  PostStatusDraft,
	}

	err := post.ValidateTitle()
	assert.Error(t, err)

	post.Title = "Good Title"
	err = post.ValidateTitle()
	assert.NoError(t, err)
}

func TestUserPost_ValidateSlug(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test",
		Slug:    "INVALID-SLUG",
		Content: "content",
		Status:  PostStatusDraft,
	}

	err := post.ValidateSlug()
	assert.Error(t, err)

	post.Slug = "valid-slug"
	err = post.ValidateSlug()
	assert.NoError(t, err)
}

func TestUserPost_ValidateContent(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test",
		Slug:    "test",
		Content: "short",
		Status:  PostStatusDraft,
	}

	err := post.ValidateContent()
	assert.Error(t, err)

	post.Content = "This is valid content with enough characters"
	err = post.ValidateContent()
	assert.NoError(t, err)
}

func TestUserPost_ValidateSEO(t *testing.T) {
	longTitle := string(make([]byte, 71))
	post := &UserPost{
		ID:        uuidv7.New(),
		UserID:    uuidv7.New(),
		Title:     "Test",
		Slug:      "test",
		Content:   "content here",
		Status:    PostStatusDraft,
		MetaTitle: &longTitle,
	}

	err := post.ValidateSEO()
	assert.Error(t, err)

	goodTitle := "Good Meta Title"
	post.MetaTitle = &goodTitle
	err = post.ValidateSEO()
	assert.NoError(t, err)
}

func TestUserPost_DecrementCountersEdgeCase(t *testing.T) {
	post := &UserPost{
		ID:      uuidv7.New(),
		UserID:  uuidv7.New(),
		Title:   "Test",
		Slug:    "test",
		Content: "content",
		Status:  PostStatusDraft,
	}

	// Should not go below zero
	post.DecrementLikes()
	assert.Equal(t, 0, post.LikeCount)

	post.DecrementComments()
	assert.Equal(t, 0, post.CommentCount)
}
