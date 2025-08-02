package kv

import (
	"encoding/json"
	"errors"

	"github.com/dgraph-io/badger/v4"

	"github.com/flarexio/talkix/user"
)

func NewUserRepository(db *badger.DB) user.Repository {
	return &userRepository{db}
}

type userRepository struct {
	db *badger.DB
}

func (repo *userRepository) Find(id string) (*user.User, error) {
	var u *user.User

	err := repo.db.View(func(txn *badger.Txn) error {
		key := []byte("user:" + id)

		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &u)
		})
	})

	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, user.ErrUserNotFound
		}

		return nil, err
	}

	return u, nil
}

func (repo *userRepository) Save(u *user.User) error {
	return repo.db.Update(func(txn *badger.Txn) error {
		key := []byte("user:" + u.ID)

		val, err := json.Marshal(&u)
		if err != nil {
			return err
		}

		return txn.Set(key, val)
	})
}
