package entity

import (
    "time"
    "github.com/google/uuid"
)

type Role struct {
    ID        uuid.UUID `db:"id" json:"id"`
    Name      string    `db:"name" json:"name"`
    Active    bool      `db:"active" json:"active"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (e *Role) IsActive() bool {
    return e.Active
}

func (e *Role) Activate() {
    e.Active = true
}

func (e *Role) Deactivate() {
    e.Active = false
}
