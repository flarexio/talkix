package talkix

import (
	"bytes"
	"context"
	"errors"
	"text/template"

	"github.com/flarexio/talkix/config"
	"github.com/flarexio/talkix/templates"
)

func NewSimpleService(cfg config.LineConfig) Service {
	templates := map[string]*template.Template{
		"login": templates.LoginTemplate(cfg.Login.AuthURL),
	}

	return &simpleService{
		cfg:       cfg,
		templates: templates,
	}
}

type simpleService struct {
	cfg       config.LineConfig
	templates map[string]*template.Template
}

func (svc *simpleService) ReplyMessage(ctx context.Context, msg Message) (Message, error) {
	m, ok := msg.(*TextMessage)
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

		return NewFlexMessage(
			"Please Login to Continue",
			buf.Bytes(),
		), nil
	}

	return NewTextMessage("Copy cat: " + m.Text), nil
}
