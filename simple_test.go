package talkix

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flarexio/talkix/config"
)

func TestSimpleService(t *testing.T) {
	assert := assert.New(t)

	var cfg config.LineConfig
	svc := NewSimpleService(cfg)

	ctx := context.Background()

	msg := NewTextMessage("Hello, world!")

	reply, err := svc.ReplyMessage(ctx, msg)
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	replyMsg, ok := reply.(*TextMessage)
	if !ok {
		err := errors.New("expected TextMessage type")
		assert.Fail(err.Error())
		return
	}

	assert.Equal("Copy cat: Hello, world!", replyMsg.Content)
}
