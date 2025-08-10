package line

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"

	line "github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"

	"github.com/flarexio/core/endpoint"
	"github.com/flarexio/talkix"
	"github.com/flarexio/talkix/config"
	"github.com/flarexio/talkix/identity"
	"github.com/flarexio/talkix/user"
)

var (
	cfg config.Config
	bot *line.MessagingApiAPI
)

func Init(config config.Config) error {
	api, err := line.NewMessagingApiAPI(config.Line.Messaging.ChannelToken)
	if err != nil {
		return err
	}

	cfg = config
	bot = api
	return nil
}

func MessageHandler(endpoint endpoint.Endpoint, directUser identity.DirectUser) gin.HandlerFunc {
	return func(c *gin.Context) {
		cb, err := webhook.ParseRequest(cfg.Line.Messaging.ChannelSecret, c.Request)
		if err != nil {
			if errors.Is(err, webhook.ErrInvalidSignature) {
				c.String(http.StatusBadRequest, err.Error())
				c.Error(err)
				c.Abort()
				return
			}

			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		for _, event := range cb.Events {
			switch e := event.(type) {
			case webhook.MessageEvent:
				u := new(user.User)

				// Determine the source of the event
				switch source := e.Source.(type) {
				case webhook.UserSource:
					u.ID = source.UserId

					profile, _, err := directUser(source.UserId)
					if err == nil {
						u.ID = profile.ID
						u.Profile = profile
						u.Verified = true
					}

					go bot.ShowLoadingAnimation(&line.ShowLoadingAnimationRequest{
						ChatId:         source.UserId,
						LoadingSeconds: 20,
					})

				default:
					err := errors.New("unsupported source type")
					c.String(http.StatusBadRequest, err.Error())
					c.Error(err)
					c.Abort()
					return
				}

				// Handle the message based on its type
				var reply talkix.Message
				switch msg := e.Message.(type) {
				case webhook.TextMessageContent:
					req := talkix.NewTextMessage(msg.Text)
					req.SetTimestamp(time.UnixMilli(e.Timestamp))

					ctx := context.Background()
					ctx = context.WithValue(ctx, talkix.UserKey, u)

					resp, err := endpoint(ctx, req)
					if err != nil {
						c.String(http.StatusInternalServerError, err.Error())
						c.Error(err)
						c.Abort()
						return
					}

					r, ok := resp.(talkix.Message)
					if !ok {
						err := errors.New("expected message type in response")
						c.String(http.StatusInternalServerError, err.Error())
						c.Error(err)
						c.Abort()
						return
					}

					reply = r

				case webhook.LocationMessageContent:
					locationText := fmt.Sprintf("Title: %s\nAddress: %s\nLatitude: %.6f\nLongitude: %.6f",
						msg.Title, msg.Address, msg.Latitude, msg.Longitude)

					req := talkix.NewTextMessage(locationText)
					req.SetTimestamp(time.UnixMilli(e.Timestamp))

					ctx := context.Background()
					ctx = context.WithValue(ctx, talkix.UserKey, u)

					resp, err := endpoint(ctx, req)
					if err != nil {
						c.String(http.StatusInternalServerError, err.Error())
						c.Error(err)
						c.Abort()
						return
					}

					r, ok := resp.(talkix.Message)
					if !ok {
						err := errors.New("expected message type in response")
						c.String(http.StatusInternalServerError, err.Error())
						c.Error(err)
						c.Abort()
						return
					}

					reply = r

				default:
					err := errors.New("unsupported message type")
					c.String(http.StatusBadRequest, err.Error())
					c.Error(err)
					c.Abort()
					return
				}

				// Reply to the message

				// Prepare quick replies if available
				items := make([]line.QuickReplyItem, 0)
				for _, qr := range reply.QuickReply() {
					items = append(items, line.QuickReplyItem{
						Type: "action",
						Action: line.MessageAction{
							Label: qr,
							Text:  qr,
						},
					})
				}

				// Send the reply based on the message type
				var lineMsg line.MessageInterface
				switch replyMsg := reply.(type) {
				case *talkix.TextMessage:
					lineMsg = line.TextMessage{
						Sender: &line.Sender{
							Name:    cfg.LLM.Model,
							IconUrl: "https://openai.com/favicon.ico",
						},
						Text: replyMsg.Text,
						QuickReply: &line.QuickReply{
							Items: items,
						},
					}

				case *talkix.FlexMessage:
					container, err := line.UnmarshalFlexContainer(replyMsg.Flex)
					if err != nil {
						lineMsg = line.TextMessage{
							Sender: &line.Sender{
								Name:    cfg.LLM.Model,
								IconUrl: "https://openai.com/favicon.ico",
							},
							Text: replyMsg.AltText,
							QuickReply: &line.QuickReply{
								Items: items,
							},
						}
					} else {
						lineMsg = line.FlexMessage{
							Sender: &line.Sender{
								Name:    cfg.LLM.Model,
								IconUrl: "https://openai.com/favicon.ico",
							},
							AltText:  replyMsg.AltText,
							Contents: container,
							QuickReply: &line.QuickReply{
								Items: items,
							},
						}
					}

				default:
					err := errors.New("expected message type in response")
					c.String(http.StatusInternalServerError, err.Error())
					c.Error(err)
					c.Abort()
					return
				}

				if _, err := bot.ReplyMessage(&line.ReplyMessageRequest{
					ReplyToken: e.ReplyToken,
					Messages:   []line.MessageInterface{lineMsg},
				}); err != nil {
					c.String(http.StatusInternalServerError, err.Error())
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
