package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

const DefaultOTPTTL = 3 * time.Minute

type OTPStore struct {
	tokens map[string]OTPData
	sync.Mutex
}

type OTPData struct {
	UserID    string
	Action    string
	Data      map[string]any
	ExpiresAt time.Time
}

func NewOTPStore() *OTPStore {
	store := &OTPStore{
		tokens: make(map[string]OTPData),
	}

	go store.cleanupRoutine()

	return store
}

func (store *OTPStore) GenerateOTP(userID string, action string, data map[string]any) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(bytes)

	store.Lock()
	defer store.Unlock()

	store.tokens[token] = OTPData{
		UserID:    userID,
		Action:    action,
		Data:      data,
		ExpiresAt: time.Now().Add(DefaultOTPTTL),
	}

	return token, nil
}

func (store *OTPStore) Validate(token string) (OTPData, error) {
	store.Lock()
	defer store.Unlock()

	data, ok := store.tokens[token]
	if !ok {
		return OTPData{}, errors.New("invalid or expired token")
	}

	delete(store.tokens, token)

	if time.Now().After(data.ExpiresAt) {
		return OTPData{}, errors.New("token has expired")
	}

	return data, nil
}

func (store *OTPStore) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		store.cleanup()
	}
}

func (store *OTPStore) cleanup() {
	store.Lock()
	defer store.Unlock()

	now := time.Now()
	for token, data := range store.tokens {
		if now.After(data.ExpiresAt) {
			delete(store.tokens, token)
		}
	}
}
