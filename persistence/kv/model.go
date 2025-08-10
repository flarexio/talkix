package kv

import (
	"time"

	"github.com/flarexio/talkix/session"
)

func NewSession(s *session.Session) *Session {
	conversationIDs := make([]string, len(s.Conversations))
	for i, conv := range s.Conversations {
		conversationIDs[i] = conv.ID
	}

	return &Session{
		ID:              s.ID,
		UserID:          s.UserID,
		Summary:         s.Summary,
		ConversationIDs: conversationIDs,
		CreatedAt:       s.CreatedAt,

		conversations: s.Conversations,
	}
}

type Session struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Summary         string    `json:"summary"`
	ConversationIDs []string  `json:"conversation_ids"`
	CreatedAt       time.Time `json:"created_at"`

	conversations []*session.Conversation `json:"-"`
}

func (s *Session) reconstitute(convs []*session.Conversation) *session.Session {
	return &session.Session{
		ID:            s.ID,
		UserID:        s.UserID,
		Summary:       s.Summary,
		Conversations: convs,
		CreatedAt:     s.CreatedAt,
	}
}
