package talkix

import (
	"bytes"
	"context"
	"errors"
	"text/template"

	"github.com/flarexio/talkix/message"
	"github.com/flarexio/talkix/templates"
)

func NewSimpleService(cfg LineConfig) Service {
	templates := map[string]*template.Template{
		"login": templates.LoginTemplate(cfg.Login.AuthURL),
	}

	return &simpleService{
		cfg:       cfg,
		templates: templates,
	}
}

type simpleService struct {
	cfg       LineConfig
	templates map[string]*template.Template
}

func (svc *simpleService) ReplyMessage(ctx context.Context, msg message.Message) (message.Message, error) {
	m, ok := msg.(*message.TextMessage)
	if !ok {
		return nil, errors.New("invalid message type")
	}

	if m.Text == "LOGIN" {
		tmpl, ok := svc.templates["login"]
		if !ok {
			return nil, errors.New("login template not found")
		}

		values := map[string]string{
			"Title":       "Please Login to Continue",
			"Description": "You need to login to access this feature.",
		}

		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, values); err != nil {
			return nil, err
		}

		return message.NewFlexMessage(
			"Please Login to Continue",
			buf.Bytes(),
		), nil
	}

	return message.NewTextMessage("Copy cat: " + m.Text), nil
}
