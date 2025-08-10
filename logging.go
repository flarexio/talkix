package talkix

import (
	"context"

	"go.uber.org/zap"

	"github.com/flarexio/talkix/session"
)

func LoggingMiddleware(name string) ServiceMiddleware {
	return func(next Service) Service {
		log := zap.L().With(
			zap.String("service", name),
		)

		log.Info("service initialized", zap.String("name", name))

		return &loggingMiddleware{
			log:  log,
			next: next,
		}
	}
}

type loggingMiddleware struct {
	log  *zap.Logger
	next Service
}

func (mw *loggingMiddleware) Name() string {
	return mw.next.Name()
}

func (mw *loggingMiddleware) ReplyMessage(ctx context.Context, msg Message) (Message, error) {
	log := mw.log.With(
		zap.String("action", "reply_message"),
		zap.String("content", msg.Content()),
		zap.Time("timestamp", msg.Timestamp()),
	)

	log.Info("replying to message")

	reply, err := mw.next.ReplyMessage(ctx, msg)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Info("message replied", zap.String("reply_content", reply.Content()))
	return reply, nil
}

func SessionLoggingMiddleware() SessionServiceMiddleware {
	return func(next SessionService) SessionService {
		log := zap.L().With(
			zap.String("service", "session"),
		)

		log.Info("session service initialized")

		return &sessionLoggingMiddleware{
			log:  log,
			next: next,
		}
	}
}

type sessionLoggingMiddleware struct {
	log  *zap.Logger
	next SessionService
}

func (mw *sessionLoggingMiddleware) ListSessions(ctx context.Context) ([]*session.Session, string, error) {
	log := mw.log.With(
		zap.String("action", "list_sessions"),
	)

	sessions, selectedSessionID, err := mw.next.ListSessions(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, "", err
	}

	log.Info("sessions listed",
		zap.Int("count", len(sessions)),
		zap.String("selected_session", selectedSessionID))

	return sessions, selectedSessionID, nil
}

func (mw *sessionLoggingMiddleware) Session(ctx context.Context, sessionID string) (*session.Session, error) {
	log := mw.log.With(
		zap.String("action", "get_session"),
		zap.String("session_id", sessionID),
	)

	session, err := mw.next.Session(ctx, sessionID)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Info("session retrieved")
	return session, nil
}

func (mw *sessionLoggingMiddleware) CreateSession(ctx context.Context) (*session.Session, error) {
	log := mw.log.With(
		zap.String("action", "create_session"),
	)

	session, err := mw.next.CreateSession(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Info("session created", zap.String("session_id", session.ID))
	return session, nil
}

func (mw *sessionLoggingMiddleware) SwitchSession(ctx context.Context, sessionID string) error {
	log := mw.log.With(
		zap.String("action", "switch_session"),
		zap.String("session", sessionID),
	)

	err := mw.next.SwitchSession(ctx, sessionID)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	log.Info("session selected")
	return nil
}

func (mw *sessionLoggingMiddleware) DeleteSession(ctx context.Context, sessionID string) error {
	log := mw.log.With(
		zap.String("action", "delete_session"),
		zap.String("session", sessionID),
	)

	err := mw.next.DeleteSession(ctx, sessionID)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	log.Info("session deleted")
	return nil
}
