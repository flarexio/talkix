package user

import (
	"time"
)

type User struct {
	ID       string       `json:"id"`
	Profile  *UserProfile `json:"-"`
	Verified bool         `json:"-"`

	SessionIDs        []string `json:"session_ids"`
	SelectedSessionID string   `json:"selected_session_id"`
}

type UserProfile struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
