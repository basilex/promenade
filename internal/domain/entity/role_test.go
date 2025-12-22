package entity

import (
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestRole_Validate(t *testing.T) {
	tests := []struct {
		name    string
		role    Role
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid role",
			role: Role{
				ID:          uuidv7.New(),
				Name:        "admin",
				DisplayName: "Administrator",
			},
			wantErr: false,
		},
		{
			name: "empty id",
			role: Role{
				Name:        "admin",
				DisplayName: "Administrator",
			},
			wantErr: true,
			errMsg:  "role id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.role.Validate()
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
