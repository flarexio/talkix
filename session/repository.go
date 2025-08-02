package session

import "errors"

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Repository interface {
	Find(id string) (*Session, error)
	Save(s *Session) error
	Delete(id string) error
}
