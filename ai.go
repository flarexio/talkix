package talkix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"text/template"
	"time"

	"github.com/flarexio/talkix/message"
	"github.com/flarexio/talkix/templates"
)

const SYSTEM_PROMPT = `You are a helpful assistant for the LINE messaging platform.

<user_profile>
{{ .UserProfile }}
</user_profile>
If <user_profile> is (NULL), please prompt the user to bind their LINE account before proceeding.

Instructions:
1. Only return a single JSON object, not an array.
2. The object must have a "type" field, which is either "text" or "flex".
3. If "type" is "text", fill the "text" field with an object containing a "text" string. Set "flex" to null.
4. If "type" is "flex", fill the "flex" field with an object containing at least "altText" (string).
   - If using a Flex Message JSON, set "flex" to the JSON string and "templateSpec" to null.
   - If using a template, set "flex" to an empty string, and fill "templateSpec" with the template name and required "values" object. Set unused template values to null.
   - Set "text" to null.
5. For templates, "templateSpec" must include:
   - "template": one of "login", "weather", or "restaurant"
   - "values": an object with keys "login", "weather", "restaurant". Only fill the one matching the template, others set to null.
6. Do not include any extra fields or properties.
7. When a tool is available for a query, always use the tool to get the latest information. Do not rely on your own internal knowledge.

Available tools:
- Time: Query the current time.
- Weather: Query weather information.
- Google Maps: Query map and location information.
  - When using the maps_search_places tool, always optimize the query for the best search result by combining the user's intent and any specific place name or context mentioned in the question.

Supported templates:
- login: Use when you want to prompt the user to log in or authorize an action. Includes a title and description.
- weather: Use when you want to show weather information for a location. Includes weather icon, temperature, humidity, wind speed, last updated time, and an "extraInfo" field for AI suggestions or additional information.
- restaurant: Use when you want to display restaurant details. Includes restaurant name, rating (stars), address, and opening hours.

Field definitions:
- type: "text" or "flex"
- text: object with "text" (string), only for type "text"
- flex: object with "altText" (string), "flex" (string or empty), "templateSpec" (object or null), only for type "flex"
- templateSpec: object with "template" (string), "values" (object with keys "login", "weather", "restaurant")
- values: for the selected template, fill required fields; others set to null

Notes:
- For the weather template, always fill the "ExtraInfo" field with an AI-generated suggestion or any additional information relevant to the user's query. If there is nothing extra to add, set it to an empty string.`

func SystemPrompt() (PromptFormat, error) {
	tmpl, err := template.New("system_prompt").Parse(SYSTEM_PROMPT)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) (string, error) {
		var userProfile string

		user, ok := ctx.Value(ContextKeyUser).(*User)
		if !ok {
			userProfile = "(NULL)"
		}

		bs, err := json.Marshal(user)
		if err != nil {
			return "", err
		}

		userProfile = string(bs)

		values := map[string]any{
			"UserProfile": userProfile,
		}

		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, values); err != nil {
			return "", err
		}

		return buf.String(), nil
	}, nil
}

func NewAIService(cfg LineConfig, tools []Tool) (Service, error) {
	prompt, err := SystemPrompt()
	if err != nil {
		return nil, err
	}

	llm, err := NewLLM("openai:gpt-4.1",
		WithPromptFormat(prompt),
		WithTools(tools),
		WithStructuredOutput(LineMessage{}),
	)

	if err != nil {
		return nil, err
	}

	templates := map[string]*template.Template{
		"login":      templates.LoginTemplate(cfg.Login.AuthURL),
		"weather":    templates.WeatherTemplate(),
		"restaurant": templates.RestaurantTemplate(),
	}

	return &aiService{
		llm:       llm,
		templates: templates,
	}, nil
}

type aiService struct {
	llm       *LLM
	templates map[string]*template.Template
}

func (svc *aiService) ReplyMessage(ctx context.Context, msg message.Message) (message.Message, error) {
	m, ok := msg.(*message.TextMessage)
	if !ok {
		return nil, errors.New("invalid message type")
	}

	var reply LineMessage
	if err := svc.llm.Invoke(ctx, m.Text, &reply); err != nil {
		return nil, err
	}

	switch reply.Type {
	case "text":
		if reply.Text == nil {
			return nil, errors.New("text message content is empty")
		}

		return message.NewTextMessage(reply.Text.Text), nil

	case "flex":
		if reply.Flex == nil {
			return nil, errors.New("flex message content is empty")
		}

		flexMsg := &message.FlexMessage{
			AltText:   reply.Flex.AltText,
			Flex:      reply.Flex.Flex,
			CreatedAt: reply.Flex.CreatedAt,
		}

		if templateSpec := reply.Flex.TemplateSpec; templateSpec != nil {
			name := templateSpec.Template

			tmpl, ok := svc.templates[name]
			if !ok {
				return nil, errors.New("unknown template: " + name)
			}

			values, ok := templateSpec.Values[name]
			if !ok {
				return nil, errors.New("missing values for template: " + name)
			}

			buf := &bytes.Buffer{}
			if err := tmpl.Execute(buf, values); err != nil {
				return nil, err
			}

			flexMsg.Flex = buf.Bytes()
		}

		return flexMsg, nil

	default:
		return nil, errors.New("unknown message type: " + reply.Type)
	}
}

type TemplateSpec struct {
	Template string         `json:"template"`
	Values   map[string]any `json:"values"`
}

type FlexMessage struct {
	message.FlexMessage
	TemplateSpec *TemplateSpec `json:"templateSpec,omitempty"`
}

type LineMessage struct {
	Type string               `json:"type"`
	Text *message.TextMessage `json:"text,omitempty"`
	Flex *FlexMessage         `json:"flex,omitempty"`
}

func (msg *LineMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type string               `json:"type"`
		Text *message.TextMessage `json:"text,omitempty"`
		Flex *struct {
			AltText      string `json:"altText"`
			Flex         string `json:"flex"`
			Timestamp    int64  `json:"timestamp"`
			TemplateSpec *struct {
				Template string         `json:"template"`
				Values   map[string]any `json:"values"`
			} `json:"templateSpec"`
		} `json:"flex,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	msg.Type = raw.Type

	switch raw.Type {
	case "text":
		msg.Text = raw.Text

	case "flex":
		if raw.Flex == nil {
			return errors.New("flex message content is empty")
		}

		flexMsg := &FlexMessage{
			FlexMessage: message.FlexMessage{
				AltText:   raw.Flex.AltText,
				CreatedAt: time.Now(),
			},
		}

		if raw.Flex.Timestamp > 0 {
			flexMsg.CreatedAt = time.UnixMilli(raw.Flex.Timestamp)
		}

		if raw.Flex.Flex == "" && raw.Flex.TemplateSpec == nil {
			return errors.New("flex message content is empty")
		}

		if raw.Flex.Flex != "" {
			flexMsg.Flex = json.RawMessage(raw.Flex.Flex)
		}

		if raw.Flex.TemplateSpec != nil {
			flexMsg.TemplateSpec = &TemplateSpec{
				Template: raw.Flex.TemplateSpec.Template,
				Values:   raw.Flex.TemplateSpec.Values,
			}
		}

		msg.Flex = flexMsg
	}

	return nil
}

func (msg LineMessage) Name() string {
	return "LineMessage"
}

func (msg LineMessage) Description() string {
	return "A LINE message object that can be either a text message or a Flex Message for rich content display."
}

func (msg LineMessage) Schema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "A LINE message object. Only one of 'text' or 'flex' should be present.",
		"properties": map[string]any{
			"type": map[string]any{
				"type":        "string",
				"description": "The type of message, either 'text' or 'flex'.",
				"enum":        []string{"text", "flex"},
			},
			"text": map[string]any{
				"type":        []string{"object", "null"},
				"description": "Text message object. Only use this when sending a simple text message.",
				"properties": map[string]any{
					"text": map[string]any{
						"type":        "string",
						"description": "The message text.",
					},
				},
				"required":             []string{"text"},
				"additionalProperties": false,
			},
			"flex": map[string]any{
				"type":        []string{"object", "null"},
				"description": "Flex message object. Only use this when sending a Flex Message.",
				"properties": map[string]any{
					"altText": map[string]any{
						"type":        "string",
						"description": "The altText for the Flex Message.",
					},
					"flex": map[string]any{
						"type":        []string{"string", "null"},
						"description": "The Flex Message JSON string (leave empty if using templateSpec).",
					},
					"templateSpec": map[string]any{
						"type":        []string{"object", "null"},
						"description": "Optional template specification for Flex Message.",
						"properties": map[string]any{
							"template": map[string]any{
								"type":        "string",
								"description": "The name of the template to use.",
								"enum":        []string{"login", "weather", "restaurant"},
							},
							"values": map[string]any{
								"type":        "object",
								"description": "Key-value pairs required by the template.",
								"properties": map[string]any{
									"login":      templates.LoginValuesSchema,
									"weather":    templates.WeatherValuesSchema,
									"restaurant": templates.RestaurantValuesSchema,
								},
								"required":             []string{"login", "weather", "restaurant"},
								"additionalProperties": false,
							},
						},
						"required":             []string{"template", "values"},
						"additionalProperties": false,
					},
				},
				"required":             []string{"altText", "flex", "templateSpec"},
				"additionalProperties": false,
			},
		},
		"required":             []string{"type", "text", "flex"},
		"additionalProperties": false,
	}
}
