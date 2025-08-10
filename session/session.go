package session

import (
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/flarexio/talkix/llm/message"
)

func NewSession(userID string) *Session {
	return &Session{
		ID:            ulid.Make().String(),
		UserID:        userID,
		Conversations: make([]*Conversation, 0),
		CreatedAt:     time.Now(),
	}
}

type Session struct {
	ID            string
	UserID        string
	Summary       string
	Conversations []*Conversation
	CreatedAt     time.Time
}

func (s *Session) AddConversation(conv *Conversation) {
	s.Conversations = append(s.Conversations, conv)

	summary, err := GenerateSummary(s.Summary, conv)
	if err != nil {
		return
	}

	s.Summary = summary
}

func NewConversation() *Conversation {
	return &Conversation{
		ID:        ulid.Make().String(),
		Messages:  make([]message.Message, 0),
		CreatedAt: time.Now(),
	}
}

type Conversation struct {
	ID        string
	Input     string
	Output    string
	Format    json.RawMessage
	Messages  []message.Message
	CreatedAt time.Time
}

func (c *Conversation) SetIO(input, output string) {
	c.Input = input
	c.Output = output
}

func (c *Conversation) SetFormat(format json.RawMessage) {
	c.Format = format
}

func (c *Conversation) AddMessage(message ...message.Message) {
	c.Messages = append(c.Messages, message...)
}

func (c *Conversation) TrimMessages(lastN int) []message.Message {
	type pair struct {
		h message.Message // Human
		a message.Message // AI
	}

	var lastHuman *message.Message

	pairs := make([]pair, 0)
	for _, m := range c.Messages {

		switch m.Role {
		case message.RoleHuman:
			lastHuman = &m

		case message.RoleAI:
			if lastHuman == nil {
				continue
			}

			if len(m.ToolCalls) > 0 {
				continue
			}

			pairs = append(pairs, pair{h: *lastHuman, a: m})
			lastHuman = nil
		}
	}

	start := 0
	if lastN > 0 && len(pairs) > lastN {
		start = len(pairs) - lastN
	}

	result := make([]message.Message, 0)
	for _, p := range pairs[start:] {
		result = append(result, p.h, p.a)
	}

	return result
}
