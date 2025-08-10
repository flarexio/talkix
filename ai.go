package talkix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/flarexio/talkix/auth"
	"github.com/flarexio/talkix/config"
	"github.com/flarexio/talkix/llm"
	"github.com/flarexio/talkix/llm/message"
	"github.com/flarexio/talkix/session"
	"github.com/flarexio/talkix/templates"
	"github.com/flarexio/talkix/user"
)

const MAIN_SYSTEM_PROMPT = `You are the primary AI assistant in a two-stage response system.

Stage 1 (Your Role): Content Generation
- You are responsible for generating helpful, accurate, and comprehensive responses to user queries
- Focus on providing complete information and analysis based on available tools and data
- Your response will be passed to a formatting agent, so prioritize content quality over formatting

Stage 2 (After You): LINE Message Formatting  
- Another AI agent will take your response and format it appropriately for LINE messaging
- The formatting agent will decide whether to use text format or structured templates based on your content
- You don't need to worry about LINE-specific formatting or structure

<UserProfile>
{{ .UserProfile }}
</UserProfile>
If <UserProfile> is (NULL), please prompt the user to bind their account before proceeding.

Instructions:
1. Provide helpful and accurate responses to user queries with complete information.
2. When a tool is available for a query, always use the tool to get the latest information. Do not rely on your own internal knowledge.
3. When a query requires a location, you MUST first use the maps_geocode tool to get the coordinates, then use the result for any further map or place queries. Do NOT guess or generate coordinates yourself.
4. The "query" field for maps_search_places or maps_place_details should include both the user's intent and any specific place name or context, and the "location" field MUST come from the maps_geocode tool result.
5. When you need location information from the user (for weather, nearby restaurants, directions, etc.), ask them politely to share their location.
6. Provide comprehensive and detailed responses that include all relevant information from tool results.
7. When using weather tools, include detailed weather information with specific data points (temperature, humidity, wind speed, conditions, etc.).
8. When using maps tools, provide complete place information including names, addresses, ratings, hours, and other relevant details.
9. Structure your responses clearly when presenting data-rich information (weather forecasts, place details, etc.) to help the formatting agent determine the best presentation format.

Content Guidelines for Tool Results:
- Weather Information: Present temperature, conditions, humidity, wind speed, and other metrics clearly
- Place Information: Include name, address, rating, business hours, and relevant details
- Session Management: Provide clear guidance about conversation management options
- Account Binding: Give clear instructions for account authentication and binding

Available tools:
- Time: Query the current time.
- Weather: Query weather information.
- Google Maps: Query map and location information.
  - When using the maps_search_places or maps_place_details tool, always optimize the query for the best search result by combining the user's intent and any specific place name or context mentioned in the question.

Usage Guidelines:
- When user asks for weather without specifying location, ask them to share their location or specify a city name
- When user asks for nearby restaurants, shops, or services, ask them to share their location
- When user asks for directions without providing starting point, ask them to share their location
- When user wants to manage conversations, provide session management guidance
- When user needs to bind their account, provide login instructions
- Always provide helpful and relevant information based on the context

Remember: Your primary focus is on content accuracy and completeness. The formatting agent will handle LINE-specific presentation based on your response content and the tools you used.
`

const LINE_SYSTEM_PROMPT = `You are a LINE message formatter. Your role is to process the output from another AI assistant and convert it into structured LINE message format.

You will receive the complete conversation flow including:
1. User messages
2. Tool calls and their outputs (weather data, place information, etc.)
3. AI assistant responses
4. The final AI response that needs LINE formatting

<Messages>
{{ .Messages }}
</Messages>

Your task: Format the FINAL AI response in the conversation into a proper LINE message structure.

Output Requirements:
- Return a single JSON object (not an array)
- Must include ALL required fields: "type", "text", "flex", "quickReply"
- "type" must be either "text" or "flex"

Format Rules:
IF type is "text":
- Fill "text" field with object containing "text" string
- Set "flex" to null

IF type is "flex":
- Fill "flex" field with object containing "altText", "flex", "templateSpec"
- Set "text" to null
- For templates: set "flex" to empty string, fill "templateSpec"
- For custom Flex JSON: set "flex" to JSON string, "templateSpec" to null

Template Selection Logic:
1. Examine tool outputs in the conversation
2. Check what structured data is available
3. Analyze the final AI response content and format
4. Consider user intent and response complexity

Use "weather" template when:
- Weather tool was called AND returned structured data
- Final AI response presents weather information in a structured way (temperature, conditions, etc.)
- AI response is primarily about weather data presentation

Use "place" template when:
- Maps/places tool was called AND returned place data
- Final AI response presents detailed place information (name, address, rating, hours)
- AI response is structured like a place recommendation or detailed place info
- AI response contains multiple specific place attributes (not just a simple mention)

Use "session_menu" template when:
- Final AI response mentions: "會話", "對話", "session", "conversation management"
- AI response is about managing or listing conversations

Use "login" template when:
- Final AI response mentions: login, account binding, authorization, "綁定帳號"
- AI response is about authentication or account setup

Use "text" template when:
- Simple conversational responses or explanations
- AI response is primarily text-based discussion or analysis
- AI response mentions places but focuses on opinions, reviews, or general discussion
- AI response provides recommendations in narrative form
- User seems to want textual information rather than structured data
- AI response doesn't present information in a structured format
- General information without specific structured presentation

Data Extraction Priority:
1. Tool outputs (primary source for structured data)
2. AI response content (for context and additional info)
3. User intent and response format preference

Template Structure:
- "templateSpec" requires "template" and "values"
- "values" must contain ALL keys: "login", "session_menu", "weather", "place"
- Only fill the matching template key, set others to null

QuickReply Rules:
- Always include 2-5 suggestions
- Max 20 characters each
- Can start with relevant emoji
- Base on conversation context and template type

Examples:
Weather: ["🌤️ 明天天氣", "📍 其他城市", "🌡️ 一週預報"]
Places: ["🍽️ 附近餐廳", "☕ 咖啡廳", "🛍️ 購物中心"]
Sessions: ["💬 管理對話", "➕ 新增會話", "🗂️ 歷史記錄"]
Login: ["🔐 登入綁定", "👤 個人資料", "⚙️ 設定"]
General: ["❓ 更多資訊", "🔄 重新查詢", "📋 相關建議"]

Template Specifications:
- login: title, description (from AI response about account binding)
- session_menu: no values needed (system generates URLs)
- weather: location, temperature, humidity, windSpeed, condition, lastUpdated, extraInfo
- place: name, rating, address (from tool outputs)

Key Points:
- Prefer "text" format for conversational, analytical, or opinion-based responses
- Use structured templates only when AI response is clearly presenting data in a structured format
- Consider user intent: do they want data presentation or discussion?
- If AI response reads like a conversation or explanation, use "text" format
- If AI response reads like a data card or structured information, consider templates
- When in doubt between text and template, choose text for better user experience
- Tool usage alone doesn't determine template choice - the AI response format matters more
`

func MainSystemPrompt(prompt string) (llm.PromptTemplate, error) {
	promptTemplate := MAIN_SYSTEM_PROMPT
	if prompt != "" {
		promptTemplate = prompt
	}

	tmpl, err := template.New("system_prompt").Parse(promptTemplate)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) ([]message.Message, error) {
		var userProfile string

		userCtx, ok := ctx.Value(UserKey).(*user.User)
		if !ok || !userCtx.Verified {
			userProfile = "(NULL)"
		} else {
			bs, err := json.Marshal(&userCtx.Profile)
			if err != nil {
				return nil, err
			}

			userProfile = string(bs)
		}

		values := map[string]any{
			"UserProfile": userProfile,
		}

		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, values); err != nil {
			return nil, err
		}

		msgs := []message.Message{
			message.SystemMessage(buf.String()),
		}

		s, ok := ctx.Value(SessionKey).(*session.Session)
		if !ok {
			return nil, errors.New("session not found in context")
		}

		if len(s.Conversations) > 0 {
			latestConv := s.Conversations[len(s.Conversations)-1]
			history := latestConv.TrimMessages(5)

			msgs = append(msgs, history...)
		}

		return msgs, nil
	}, nil
}

func LineSystemPrompt(prompt string) (llm.PromptTemplate, error) {
	promptTemplate := LINE_SYSTEM_PROMPT
	if prompt != "" {
		promptTemplate = prompt
	}

	tmpl, err := template.New("line_system_prompt").Parse(promptTemplate)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) ([]message.Message, error) {
		messages, ok := ctx.Value(MessagesKey).([]message.Message)
		if !ok {
			return nil, errors.New("messages not found in context")
		}

		// 將對話內容轉換為字串格式
		var messagesStr string
		for _, msg := range messages[1:] {
			messagesStr += msg.PrettyFormat()
		}

		values := map[string]any{
			"Messages": messagesStr,
		}

		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, values); err != nil {
			return nil, err
		}

		msgs := []message.Message{
			message.SystemMessage(buf.String()),
		}

		return msgs, nil
	}, nil
}

func NewAIService(cfg config.Config, tools []llm.Tool, otp *auth.OTPStore,
	users user.Repository, sessions session.Repository,
) (Service, error) {
	// 主要邏輯處理
	mainPrompt, err := MainSystemPrompt(cfg.LLM.Prompt)
	if err != nil {
		return nil, err
	}

	mainLLM, err := llm.NewLLM(cfg.LLM.Model,
		llm.WithPrompt(mainPrompt),
		llm.WithTools(tools),
	)

	if err != nil {
		return nil, err
	}

	// LINE 格式化 LLM
	linePrompt, err := LineSystemPrompt(cfg.LLM.Line.Prompt)
	if err != nil {
		return nil, err
	}

	lineLLM, err := llm.NewLLM(cfg.LLM.Line.Model,
		llm.WithPrompt(linePrompt),
		llm.WithStructuredOutput(LineMessage{}),
	)

	if err != nil {
		return nil, err
	}

	templates := map[string]*template.Template{
		"login":        templates.LoginTemplate(cfg.Line.Login.AuthURL),
		"session_menu": templates.SessionMenuTemplate(),
		"weather":      templates.WeatherTemplate(),
		"place":        templates.PlaceTemplate(),
	}

	return &aiService{
		cfg:       cfg,
		mainLLM:   mainLLM,
		lineLLM:   lineLLM,
		templates: templates,
		otp:       otp,
		users:     users,
		sessions:  sessions,
	}, nil
}

type aiService struct {
	cfg       config.Config
	mainLLM   *llm.LLM
	lineLLM   *llm.LLM
	templates map[string]*template.Template
	otp       *auth.OTPStore
	users     user.Repository
	sessions  session.Repository
}

func (svc *aiService) Name() string {
	return "ai"
}

func (svc *aiService) prepareContext(ctx context.Context) (context.Context, error) {
	userCtx, ok := ctx.Value(UserKey).(*user.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}

	u, err := svc.users.Find(userCtx.ID)
	if err != nil {
		u = userCtx
	} else {
		u.Profile = userCtx.Profile
		u.Verified = userCtx.Verified
	}

	ctx = context.WithValue(ctx, UserKey, u)

	var s *session.Session
	if u.SelectedSessionID == "" {
		s = session.NewSession(u.ID)
		if err := svc.sessions.Save(s); err != nil {
			return nil, err
		}

		u.AddSessionID(s.ID)
		if err := svc.users.Save(u); err != nil {
			return nil, err
		}

	} else {
		found, err := svc.sessions.Find(u.SelectedSessionID)
		if err != nil {
			return nil, err
		}

		s = found
	}

	ctx = context.WithValue(ctx, SessionKey, s)
	return ctx, nil
}

func (svc *aiService) saveConversation(ctx context.Context, conv *session.Conversation) error {
	s, ok := ctx.Value(SessionKey).(*session.Session)
	if !ok {
		return errors.New("session not found in context")
	}

	s.AddConversation(conv)
	return svc.sessions.Save(s)
}

func (svc *aiService) ReplyMessage(ctx context.Context, msg Message) (Message, error) {
	ctx, err := svc.prepareContext(ctx)
	if err != nil {
		return nil, err
	}

	u, ok := ctx.Value(UserKey).(*user.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}

	m, ok := msg.(*TextMessage)
	if !ok {
		return nil, errors.New("invalid message type")
	}

	msgs, err := svc.mainLLM.Invoke(ctx, m.Text)
	if err != nil {
		return nil, err
	}

	if len(msgs) == 0 {
		return nil, errors.New("no messages")
	}

	resp := msgs[len(msgs)-1]

	c := session.NewConversation()
	c.SetIO(m.Text, resp.Content)
	c.AddMessage(msgs...)

	ctx = context.WithValue(ctx, MessagesKey, msgs)

	msgs, err = svc.lineLLM.Invoke(ctx, m.Text)
	if err != nil {
		return nil, err
	}

	if len(msgs) == 0 {
		return nil, errors.New("no messages")
	}

	resp = msgs[len(msgs)-1]
	jsonBytes := []byte(resp.Content)

	c.SetFormat(jsonBytes)

	if err := svc.saveConversation(ctx, c); err != nil {
		return nil, err
	}

	var reply LineMessage
	if err := json.Unmarshal(jsonBytes, &reply); err != nil {
		return nil, err
	}

	switch reply.Type {
	case "text":
		if reply.Text == nil {
			return nil, errors.New("text message content is empty")
		}

		msg := NewTextMessage(reply.Text.Text)
		msg.AddQuickReply(reply.QuickReply...)
		return msg, nil

	case "flex":
		if reply.Flex == nil {
			return nil, errors.New("flex message content is empty")
		}

		flexMsg := &FlexMessage{
			AltText:      reply.Flex.AltText,
			Flex:         reply.Flex.Flex,
			CreatedAt:    reply.Flex.CreatedAt,
			QuickReplies: reply.QuickReply,
		}

		if templateSpec := reply.Flex.TemplateSpec; templateSpec != nil {
			name := templateSpec.Template

			tmpl, ok := svc.templates[name]
			if !ok {
				return nil, errors.New("unknown template: " + name)
			}

			var values any
			switch name {
			case "session_menu":
				otp, err := svc.otp.GenerateOTP(u.ID, "list_sessions", nil)
				if err != nil {
					return nil, err
				}

				url := fmt.Sprintf("%s/users/%s/session/list?token=%s", svc.cfg.BaseURL, u.Profile.Username, otp)

				values = templates.SessionMenuValues{
					ListSessionsURL: url,
				}

			default:
				vals, ok := templateSpec.Values[name]
				if !ok {
					return nil, errors.New("missing values for template: " + name)
				}

				values = vals
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

type FlexMessageWithTemplate struct {
	FlexMessage
	TemplateSpec *TemplateSpec `json:"templateSpec,omitempty"`
}

type LineMessage struct {
	Type       string                   `json:"type"`
	Text       *TextMessage             `json:"text,omitempty"`
	Flex       *FlexMessageWithTemplate `json:"flex,omitempty"`
	QuickReply []string                 `json:"quickReply,omitempty"`
}

func (msg *LineMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type string       `json:"type"`
		Text *TextMessage `json:"text,omitempty"`
		Flex *struct {
			AltText      string `json:"altText"`
			Flex         string `json:"flex"`
			TemplateSpec *struct {
				Template string         `json:"template"`
				Values   map[string]any `json:"values"`
			} `json:"templateSpec"`
		} `json:"flex,omitempty"`
		QuickReply []string `json:"quickReply,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	msg.Type = raw.Type
	msg.QuickReply = raw.QuickReply

	switch raw.Type {
	case "text":
		msg.Text = raw.Text

	case "flex":
		if raw.Flex == nil {
			return errors.New("flex message content is empty")
		}

		flexMsg := &FlexMessageWithTemplate{
			FlexMessage: FlexMessage{
				AltText:   raw.Flex.AltText,
				CreatedAt: time.Now(),
			},
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
								"enum": []string{
									"login",
									"session_menu",
									"weather",
									"place",
								},
							},
							"values": map[string]any{
								"type":        "object",
								"description": "Key-value pairs required by the template.",
								"properties": map[string]any{
									"login":        templates.LoginValuesSchema,
									"session_menu": templates.SessionMenuValuesSchema,
									"weather":      templates.WeatherValuesSchema,
									"place":        templates.PlaceValuesSchema,
								},
								"required":             []string{"login", "session_menu", "weather", "place"},
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
			"quickReply": map[string]any{
				"type":        "array",
				"description": "Optional array of quick reply text suggestions (max 13 items, can start with emoji).",
				"maxItems":    13,
				"items": map[string]any{
					"type":        "string",
					"description": "Quick reply text suggestion (max 20 chars, can start with emoji like 🌤️, 🍽️, 📍).",
					"maxLength":   20,
				},
			},
		},
		"required":             []string{"type", "text", "flex", "quickReply"},
		"additionalProperties": false,
	}
}
