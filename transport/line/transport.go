package line

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"

	line "github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"

	"github.com/flarexio/talkix"
	"github.com/flarexio/talkix/config"
	"github.com/flarexio/talkix/user"
)

var (
	cfg config.LineConfig
	bot *line.MessagingApiAPI
)

func Init(config config.LineConfig) error {
	api, err := line.NewMessagingApiAPI(config.Messaging.ChannelToken)
	if err != nil {
		return err
	}

	cfg = config
	bot = api
	return nil
}

type DirectIdentityUser func(username string) (*user.UserProfile, error)

func MessageHandler(endpoint endpoint.Endpoint, directIdentityUser DirectIdentityUser) gin.HandlerFunc {
	return func(c *gin.Context) {
		cb, err := webhook.ParseRequest(cfg.Messaging.ChannelSecret, c.Request)
		if err != nil {
			if errors.Is(err, webhook.ErrInvalidSignature) {
				c.String(http.StatusBadRequest, err.Error())
				c.Error(err)
				c.Abort()
			} else {
				c.String(http.StatusInternalServerError, err.Error())
				c.Error(err)
				c.Abort()
			}
			return
		}

		for _, event := range cb.Events {
			switch e := event.(type) {
			case webhook.MessageEvent:
				u := new(user.User)

				switch source := e.Source.(type) {
				case webhook.UserSource:
					u.ID = source.UserId

					profile, err := directIdentityUser(source.UserId)
					if err == nil {
						u.Profile = profile
						u.Verified = true
					}

				default:
					err := fmt.Errorf("unsupported source type: %T", source)
					c.String(http.StatusBadRequest, err.Error())
					c.Error(err)
					c.Abort()
					return
				}

				switch msg := e.Message.(type) {
				case webhook.TextMessageContent:
					req := talkix.NewTextMessage(msg.Text)
					req.SetTimestamp(time.UnixMilli(e.Timestamp))

					ctx := context.Background()
					ctx = context.WithValue(ctx, talkix.UserKey, u)

					reply, err := endpoint(ctx, req)
					if err != nil {
						c.String(http.StatusInternalServerError, err.Error())
						c.Error(err)
						c.Abort()
						return
					}

					switch replyMsg := reply.(type) {
					case *talkix.TextMessage:
						_, err := bot.ReplyMessage(&line.ReplyMessageRequest{
							ReplyToken: e.ReplyToken,
							Messages: []line.MessageInterface{
								line.TextMessage{
									Text: replyMsg.Text,
								},
							},
						})

						if err != nil {
							c.String(http.StatusInternalServerError, err.Error())
							c.Error(err)
							c.Abort()
							return
						}

					case *talkix.FlexMessage:
						container, err := line.UnmarshalFlexContainer(replyMsg.Flex)
						if err != nil {
							c.String(http.StatusInternalServerError, err.Error())
							c.Error(err)
							c.Abort()
							return
						}

						if _, err := bot.ReplyMessage(&line.ReplyMessageRequest{
							ReplyToken: e.ReplyToken,
							Messages: []line.MessageInterface{
								line.FlexMessage{
									AltText:  replyMsg.AltText,
									Contents: container,
								},
							},
						}); err != nil {
							c.String(http.StatusInternalServerError, err.Error())
							c.Error(err)
							c.Abort()
							return
						}

					default:
						err := errors.New("expected Message type in response")
						c.String(http.StatusInternalServerError, err.Error())
						c.Error(err)
						c.Abort()
						return
					}

				default:
					err := errors.New("unsupported message type")
					c.String(http.StatusBadRequest, err.Error())
					c.Error(err)
					c.Abort()
					return
				}

			default:
				err := errors.New("unsupported event type")
				c.String(http.StatusBadRequest, err.Error())
				c.Error(err)
				c.Abort()
				return
			}
		}
	}
}
