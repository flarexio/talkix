package talkix

import (
	"context"

	"go.uber.org/zap"
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
