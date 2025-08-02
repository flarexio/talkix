package kv

import (
	"encoding/json"
	"errors"

	"github.com/dgraph-io/badger/v4"

	"github.com/flarexio/talkix/session"
)

func NewSessionRepository(db *badger.DB) session.Repository {
	return &sessionRepository{db}
}

type sessionRepository struct {
	db *badger.DB
}

func (repo *sessionRepository) Find(id string) (*session.Session, error) {
	var s *session.Session

	err := repo.db.View(func(txn *badger.Txn) error {
		key := []byte("session:" + id)

		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		var ss *Session

		if err := item.Value(func(val []byte) error {
			return json.Unmarshal(val, &ss)
		}); err != nil {
			return err
		}

		convs := make([]*session.Conversation, len(ss.ConversationIDs))
		for i, id := range ss.ConversationIDs {
			key := []byte("conversation:" + id)
			convItem, err := txn.Get(key)
			if err != nil {
				return err
			}

			var conv *session.Conversation
			if err := convItem.Value(func(val []byte) error {
				return json.Unmarshal(val, &conv)
			}); err != nil {
				return err
			}

			convs[i] = conv
		}

		// reconstitute converts the data model back to the domain model
		s = ss.reconstitute(convs)
		return nil
	})

	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, session.ErrSessionNotFound
		}

		return nil, err
	}

	return s, nil
}

func (repo *sessionRepository) Save(s *session.Session) error {
	ss := NewSession(s) // convert domain to data model

	return repo.db.Update(func(txn *badger.Txn) error {
		for _, conv := range ss.conversations {
			key := []byte("conversation:" + conv.ID)

			val, err := json.Marshal(&conv)
			if err != nil {
				return err
			}

			if err := txn.Set(key, val); err != nil {
				return err
			}
		}

		key := []byte("session:" + ss.ID)

		val, err := json.Marshal(&ss)
		if err != nil {
			return err
		}

		return txn.Set(key, val)
	})
}

func (repo *sessionRepository) Delete(id string) error {
	return repo.db.Update(func(txn *badger.Txn) error {
		key := []byte("session:" + id)

		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		var s *Session

		if err := item.Value(func(val []byte) error {
			return json.Unmarshal(val, &s)
		}); err != nil {
			return err
		}

		for _, id := range s.ConversationIDs {
			key := []byte("conversation:" + id)
			if err := txn.Delete(key); err != nil {
				return err
			}
		}

		return txn.Delete(key)
	})
}
