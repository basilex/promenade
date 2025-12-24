package entity

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestAuditEventCreate_Validate(t *testing.T) {
	validUserID := uuidv7.New()
	validEntityID := uuidv7.New()

	t.Run("valid audit event", func(t *testing.T) {
		create := &AuditEventCreate{
			UserID:     validUserID,
			Action:     "user.created",
			EntityType: "user",
			EntityID:   validEntityID,
		}
		err := create.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid with optional fields", func(t *testing.T) {
		oldData := `{"status":"draft"}`
		newData := `{"status":"published"}`
		metadata := `{"editor":"admin"}`

		create := &AuditEventCreate{
			UserID:     validUserID,
			Action:     "post.updated",
			EntityType: "post",
			EntityID:   validEntityID,
			OldData:    &oldData,
			NewData:    &newData,
			Metadata:   &metadata,
			IPAddress:  "127.0.0.1",
			UserAgent:  "Browser/1.0",
			RequestID:  "req-123",
		}
		err := create.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing user ID", func(t *testing.T) {
		create := &AuditEventCreate{
			Action:     "user.created",
			EntityType: "user",
			EntityID:   validEntityID,
		}
		err := create.Validate()
		assert.Equal(t, ErrUserIDRequired, err)
	})

	t.Run("missing action", func(t *testing.T) {
		create := &AuditEventCreate{
			UserID:     validUserID,
			EntityType: "user",
			EntityID:   validEntityID,
		}
		err := create.Validate()
		assert.Equal(t, ErrActionRequired, err)
	})

	t.Run("missing entity type", func(t *testing.T) {
		create := &AuditEventCreate{
			UserID:   validUserID,
			Action:   "user.created",
			EntityID: validEntityID,
		}
		err := create.Validate()
		assert.Equal(t, ErrEntityTypeRequired, err)
	})

	t.Run("missing entity ID", func(t *testing.T) {
		create := &AuditEventCreate{
			UserID:     validUserID,
			Action:     "user.created",
			EntityType: "user",
		}
		err := create.Validate()
		assert.Equal(t, ErrEntityIDRequired, err)
	})
}

func TestAuditEvent_Creation(t *testing.T) {
	t.Run("basic audit event", func(t *testing.T) {
		event := &AuditEvent{
			ID:         uuidv7.New(),
			UserID:     uuidv7.New(),
			Action:     "user.login",
			EntityType: "user",
			EntityID:   uuidv7.New(),
			IPAddress:  "192.168.1.1",
			UserAgent:  "Mozilla/5.0",
			RequestID:  "req-abc-123",
			Signature:  "signature-hash",
			CreatedAt:  time.Now(),
		}

		assert.NotNil(t, event)
		assert.Equal(t, "user.login", event.Action)
		assert.Equal(t, "user", event.EntityType)
		assert.NotEmpty(t, event.Signature)
	})

	t.Run("audit event with old and new data", func(t *testing.T) {
		oldData := `{"name":"John"}`
		newData := `{"name":"Jane"}`

		event := &AuditEvent{
			ID:         uuidv7.New(),
			UserID:     uuidv7.New(),
			Action:     "user.updated",
			EntityType: "user",
			EntityID:   uuidv7.New(),
			OldData:    &oldData,
			NewData:    &newData,
			IPAddress:  "10.0.0.1",
			Signature:  "sig",
			CreatedAt:  time.Now(),
		}

		assert.NotNil(t, event.OldData)
		assert.NotNil(t, event.NewData)
		assert.Equal(t, oldData, *event.OldData)
		assert.Equal(t, newData, *event.NewData)
	})

	t.Run("audit event with metadata", func(t *testing.T) {
		metadata := `{"reason":"spam","priority":"high"}`

		event := &AuditEvent{
			ID:         uuidv7.New(),
			UserID:     uuidv7.New(),
			Action:     "post.deleted",
			EntityType: "post",
			EntityID:   uuidv7.New(),
			Metadata:   &metadata,
			IPAddress:  "127.0.0.1",
			Signature:  "sig",
			CreatedAt:  time.Now(),
		}

		assert.NotNil(t, event.Metadata)
		assert.Equal(t, metadata, *event.Metadata)
	})
}

func TestAuditError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *AuditError
		wantMsg string
	}{
		{
			name:    "user ID required",
			err:     ErrUserIDRequired,
			wantMsg: "User ID is required",
		},
		{
			name:    "action required",
			err:     ErrActionRequired,
			wantMsg: "Action is required",
		},
		{
			name:    "entity type required",
			err:     ErrEntityTypeRequired,
			wantMsg: "Entity type is required",
		},
		{
			name:    "entity ID required",
			err:     ErrEntityIDRequired,
			wantMsg: "Entity ID is required",
		},
		{
			name:    "not found",
			err:     ErrAuditEventNotFound,
			wantMsg: "Audit event not found",
		},
		{
			name:    "invalid signature",
			err:     ErrInvalidSignature,
			wantMsg: "Invalid audit event signature",
		},
		{
			name:    "immutable",
			err:     ErrAuditEventImmutable,
			wantMsg: "Audit events are immutable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMsg, tt.err.Error())
		})
	}
}
