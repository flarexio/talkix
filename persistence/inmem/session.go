package inmem

import (
	"sync"

	"github.com/flarexio/talkix/session"
)

func NewSessionRepository() session.Repository {
	return &sessionRepository{
		sessions: make(map[string]*session.Session),
	}
}

type sessionRepository struct {
	sessions map[string]*session.Session
	sync.RWMutex
}

func (repo *sessionRepository) Find(id string) (*session.Session, error) {
	repo.RLock()
	defer repo.RUnlock()

	s, ok := repo.sessions[id]
	if !ok {
		return nil, session.ErrSessionNotFound
	}
	return s, nil
}

func (repo *sessionRepository) Save(s *session.Session) error {
	repo.Lock()
	defer repo.Unlock()

	repo.sessions[s.ID] = s
	return nil
}

func (repo *sessionRepository) Delete(id string) error {
	repo.Lock()
	defer repo.Unlock()

	_, ok := repo.sessions[id]
	if !ok {
		return session.ErrSessionNotFound
	}

	delete(repo.sessions, id)
	return nil
}
