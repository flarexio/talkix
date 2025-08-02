package talkix

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type ReplyMessageRequest = Message
type ReplyMessageResponse = Message

func ReplyMessageEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ReplyMessageRequest)

		reply, err := service.ReplyMessage(ctx, req)
		if err != nil {
			return nil, err
		}

		return reply, nil
	}
}
