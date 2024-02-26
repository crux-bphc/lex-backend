package database

import (
	"time"

	"github.com/surrealdb/surrealdb.go"
)

type impartusRepository struct {
	DB *surrealdb.DB
}

type ImpartusUser struct {
	surrealdb.Basemodel `table:"user"`
	ID                  string    `json:"id,omitempty"`
	EMail               string    `json:"email,omitempty"`
	Jwt                 string    `json:"jwt,omitempty"`
	Password            string    `json:"password,omitempty"`
	UpdatedAt           time.Time `json:"updated_at,omitempty"`
}
