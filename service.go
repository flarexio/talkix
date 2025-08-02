package talkix

import (
	"context"
)

type Service interface {
	ReplyMessage(ctx context.Context, msg Message) (reply Message, err error)
}

type ServiceMiddleware func(Service) Service
