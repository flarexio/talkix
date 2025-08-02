package kv

import (
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
		ConversationIDs: conversationIDs,

		conversations: s.Conversations,
	}
}

type Session struct {
	ID              string   `json:"id"`
	UserID          string   `json:"user_id"`
	ConversationIDs []string `json:"conversation_ids"`

	conversations []*session.Conversation `json:"-"`
}

func (s *Session) reconstitute(convs []*session.Conversation) *session.Session {
	return &session.Session{
		ID:            s.ID,
		UserID:        s.UserID,
		Conversations: convs,
	}
}
