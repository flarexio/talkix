package inmem

import (
	"sync"

	"github.com/flarexio/talkix/user"
)

func NewUserRepository() (user.Repository, error) {
	return &userRepository{
		users: make(map[string]*user.User),
	}, nil
}

type userRepository struct {
	users map[string]*user.User
	sync.RWMutex
}

func (repo *userRepository) Find(id string) (*user.User, error) {
	repo.RLock()
	defer repo.RUnlock()

	s, ok := repo.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return s, nil
}

func (repo *userRepository) Save(u *user.User) error {
	repo.Lock()
	defer repo.Unlock()

	repo.users[u.ID] = u
	return nil
}
