package talkix

import (
	"encoding/json"
	"time"
)

type Message interface {
	Type() string
	Content() string
	Timestamp() time.Time
	SetTimestamp(time.Time)
}

func NewTextMessage(text string) Message {
	return &TextMessage{
		Text:      text,
		CreatedAt: time.Now(),
	}
}

type TextMessage struct {
	Text      string
	CreatedAt time.Time
}

func (m *TextMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Text      string `json:"text"`
		Timestamp int64  `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.Text = raw.Text

	m.CreatedAt = time.Now()
	if raw.Timestamp > 0 {
		m.CreatedAt = time.UnixMilli(raw.Timestamp)
	}

	return nil
}

func (m *TextMessage) Type() string {
	return "text"
}

func (m *TextMessage) Content() string {
	return m.Text
}

func (m *TextMessage) Timestamp() time.Time {
	return m.CreatedAt
}

func (m *TextMessage) SetTimestamp(t time.Time) {
	m.CreatedAt = t
}

func NewFlexMessage(alt string, flex json.RawMessage) Message {
	return &FlexMessage{
		AltText:   alt,
		Flex:      flex,
		CreatedAt: time.Now(),
	}
}

type FlexMessage struct {
	AltText   string
	Flex      json.RawMessage
	CreatedAt time.Time
}

func (m *FlexMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		AltText   string `json:"altText"`
		Flex      string `json:"flex"`
		Timestamp int64  `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.AltText = raw.AltText

	if raw.Flex != "" {
		m.Flex = json.RawMessage(raw.Flex)
	}

	m.CreatedAt = time.Now()
	if raw.Timestamp > 0 {
		m.CreatedAt = time.UnixMilli(raw.Timestamp)
	}

	return nil
}

func (m *FlexMessage) Type() string {
	return "flex"
}

func (m *FlexMessage) Content() string {
	return m.AltText
}

func (m *FlexMessage) Timestamp() time.Time {
	return m.CreatedAt
}

func (m *FlexMessage) SetTimestamp(t time.Time) {
	m.CreatedAt = t
}
