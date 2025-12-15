package entity

import (
    "time"
    "github.com/google/uuid"
)

type Product struct {
    ID        uuid.UUID `db:"id" json:"id"`
    Name      string    `db:"name" json:"name"`
    Active    bool      `db:"active" json:"active"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (e *Product) IsActive() bool {
    return e.Active
}

func (e *Product) Activate() {
    e.Active = true
}

func (e *Product) Deactivate() {
    e.Active = false
}
