package talkix

import (
	"context"
	"errors"

	"github.com/flarexio/core/endpoint"
	"github.com/flarexio/talkix/session"
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

type ListSessionsResponse struct {
	Sessions          []*session.Session `json:"sessions"`
	SelectedSessionID string             `json:"selected_session_id"`
}

func ListSessionsEndpoint(service SessionService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		sessions, selectedSessionID, err := service.ListSessions(ctx)
		if err != nil {
			return nil, err
		}

		resp := &ListSessionsResponse{
			Sessions:          sessions,
			SelectedSessionID: selectedSessionID,
		}

		return resp, nil
	}
}

func SessionEndpoint(service SessionService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		sessionID, ok := request.(string)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		return service.Session(ctx, sessionID)
	}
}

func CreateSessionEndpoint(service SessionService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		session, err := service.CreateSession(ctx)
		if err != nil {
			return nil, err
		}

		return session, nil
	}
}

func SwitchSessionEndpoint(service SessionService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		sessionID, ok := request.(string)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		err := service.SwitchSession(ctx, sessionID)
		return nil, err
	}
}

func DeleteSessionEndpoint(service SessionService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		sessionID, ok := request.(string)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		err := service.DeleteSession(ctx, sessionID)
		return nil, err
	}
}
