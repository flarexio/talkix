package talkix

import (
	"context"
	"errors"

	"github.com/flarexio/talkix/session"
	"github.com/flarexio/talkix/user"
)

type Service interface {
	Name() string
	ReplyMessage(ctx context.Context, msg Message) (reply Message, err error)
}

type ServiceMiddleware func(Service) Service

type SessionService interface {
	ListSessions(ctx context.Context) (sessions []*session.Session, selectedSessionID string, err error)
	Session(ctx context.Context, sessionID string) (*session.Session, error)
	CreateSession(ctx context.Context) (*session.Session, error)
	SwitchSession(ctx context.Context, sessionID string) error
	DeleteSession(ctx context.Context, sessionID string) error
}

type SessionServiceMiddleware func(SessionService) SessionService

func NewSessionService(users user.Repository, sessions session.Repository) SessionService {
	return &sessionService{
		users:    users,
		sessions: sessions,
	}
}

type sessionService struct {
	users    user.Repository
	sessions session.Repository
}

func (svc *sessionService) ListSessions(ctx context.Context) ([]*session.Session, string, error) {
	userCtx, ok := ctx.Value(UserKey).(*user.User)
	if !ok {
		return nil, "", errors.New("user not found in context")
	}

	u, err := svc.users.Find(userCtx.ID)
	if err != nil {
		return nil, "", errors.New(err.Error())
	}

	sessions := make([]*session.Session, 0)
	for _, id := range u.SessionIDs {
		session, err := svc.sessions.Find(id)
		if err != nil {
			return nil, "", errors.New(err.Error())
		}

		sessions = append(sessions, session)
	}

	return sessions, u.SelectedSessionID, nil
}

func (svc *sessionService) Session(ctx context.Context, sessionID string) (*session.Session, error) {
	userCtx, ok := ctx.Value(UserKey).(*user.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}

	session, err := svc.sessions.Find(sessionID)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	if session.UserID != userCtx.ID {
		return nil, errors.New("session does not belong to user")
	}

	return session, nil
}

func (svc *sessionService) CreateSession(ctx context.Context) (*session.Session, error) {
	userCtx, ok := ctx.Value(UserKey).(*user.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}

	u, err := svc.users.Find(userCtx.ID)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	newSession := session.NewSession(u.ID)
	if err := svc.sessions.Save(newSession); err != nil {
		return nil, errors.New(err.Error())
	}

	u.AddSessionID(newSession.ID)

	if err := svc.users.Save(u); err != nil {
		return nil, errors.New(err.Error())
	}

	return newSession, nil
}

func (svc *sessionService) SwitchSession(ctx context.Context, sessionID string) error {
	userCtx, ok := ctx.Value(UserKey).(*user.User)
	if !ok {
		return errors.New("user not found in context")
	}

	u, err := svc.users.Find(userCtx.ID)
	if err != nil {
		return errors.New(err.Error())
	}

	session, err := svc.sessions.Find(sessionID)
	if err != nil {
		return errors.New(err.Error())
	}

	u.SelectedSessionID = session.ID

	return svc.users.Save(u)
}

func (svc *sessionService) DeleteSession(ctx context.Context, sessionID string) error {
	userCtx, ok := ctx.Value(UserKey).(*user.User)
	if !ok {
		return errors.New("user not found in context")
	}

	u, err := svc.users.Find(userCtx.ID)
	if err != nil {
		return errors.New(err.Error())
	}

	session, err := svc.sessions.Find(sessionID)
	if err != nil {
		return errors.New(err.Error())
	}

	if session.UserID != u.ID {
		return errors.New("session does not belong to user")
	}

	if err := u.RemoveSessionID(session.ID); err != nil {
		return err
	}

	if err := svc.users.Save(u); err != nil {
		return err
	}

	return svc.sessions.Delete(session.ID)
}
