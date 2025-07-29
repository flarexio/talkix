package talkix

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/flarexio/talkix/message"
)

type ReplyMessageRequest = message.Message

type ReplyMessageResponse = message.Message

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
