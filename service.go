package talkix

import (
	"context"
	"time"

	"github.com/flarexio/talkix/message"
)

type Service interface {
	ReplyMessage(ctx context.Context, msg message.Message) (reply message.Message, err error)
}

type ServiceMiddleware func(Service) Service

type ContextKey string

const (
	ContextKeyUser ContextKey = "user"
)

type User struct {
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
