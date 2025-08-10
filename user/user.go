package user

import (
	"errors"
	"time"
)

type User struct {
	ID       string       `json:"id"`
	Profile  *UserProfile `json:"-"`
	Verified bool         `json:"-"`

	SessionIDs        []string `json:"session_ids"`
	SelectedSessionID string   `json:"selected_session_id"`
}

func (u *User) AddSessionID(id string) {
	u.SessionIDs = append(u.SessionIDs, id)
	u.SelectedSessionID = id
}

func (u *User) RemoveSessionID(id string) error {
	if u.SelectedSessionID == id {
		return errors.New("cannot remove selected session")
	}

	for i, sessionID := range u.SessionIDs {
		if sessionID == id {
			u.SessionIDs = append(u.SessionIDs[:i], u.SessionIDs[i+1:]...)
			return nil
		}
	}

	return errors.New("id not found")
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
