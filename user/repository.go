package user

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
)

type Repository interface {
	Find(id string) (*User, error)
	Save(u *User) error
}
