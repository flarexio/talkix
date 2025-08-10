package llm

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/openai/openai-go"

	"github.com/flarexio/talkix/llm/message"
)

func NewLLM(model string, opts ...Option) (*LLM, error) {
	client := openai.NewClient()

	model, ok := strings.CutPrefix(model, "openai:")
	if !ok {
		return nil, errors.New("model must start with 'openai:'")
	}

	llm := &LLM{
		client: client,
		body: openai.ChatCompletionNewParams{
			Model: openai.ChatModel(model),
		},
	}

	for _, opt := range opts {
		err := opt.Apply(llm)
		if err != nil {
			return nil, err
		}
	}

	return llm, nil
}

type LLM struct {
	client openai.Client
	body   openai.ChatCompletionNewParams
	prompt PromptTemplate
	tools  map[string]Tool
}

type Option interface {
	Apply(*LLM) error
}

type PromptTemplate func(ctx context.Context) ([]message.Message, error)

func WithPrompt(tmpl PromptTemplate) Option {
	return &llmWithPrompt{prompt: tmpl}
}

type llmWithPrompt struct {
	prompt PromptTemplate
}

func (opt *llmWithPrompt) Apply(llm *LLM) error {
	if llm == nil {
		return errors.New("llm cannot be nil")
	}

	llm.prompt = opt.prompt
	return nil
}

func WithStructuredOutput(schema Schema) Option {
	return &llmWithStructuredOutput{schema}
}

type llmWithStructuredOutput struct {
	schema Schema
}

func (opt *llmWithStructuredOutput) Apply(llm *LLM) error {
	if llm == nil {
		return errors.New("llm cannot be nil")
	}

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        opt.schema.Name(),
		Description: openai.String(opt.schema.Description()),
		Schema:      opt.schema.Schema(),
		Strict:      openai.Bool(true),
	}

	llm.body.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
			JSONSchema: schemaParam,
		},
	}

	return nil
}

func WithTools(tools []Tool) Option {
	return &llmWithTools{tools: tools}
}

type llmWithTools struct {
	tools []Tool
}

func (opt *llmWithTools) Apply(llm *LLM) error {
	if llm == nil {
		return errors.New("llm cannot be nil")
	}

	toolMap := make(map[string]Tool)

	tools := make([]openai.ChatCompletionToolParam, len(opt.tools))
	for i, tool := range opt.tools {
		tools[i] = openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name(),
				Description: openai.String(tool.Description()),
				Parameters:  openai.FunctionParameters(tool.Parameters()),
			},
		}

		toolMap[tool.Name()] = tool
	}

	llm.body.Tools = tools
	llm.tools = toolMap

	return nil
}

func (llm *LLM) Invoke(ctx context.Context, msg string) ([]message.Message, error) {
	msgs := []message.Message{
		message.SystemMessage("You are a helpful assistant."),
	}

	if llm.prompt != nil {
		messages, err := llm.prompt(ctx)
		if err != nil {
			return nil, err
		}

		msgs = messages
	}

	if msg != "" {
		msgs = append(msgs, message.HumanMessage(msg))
	}

	return llm.InvokeWithMessages(ctx, msgs)
}

func (llm *LLM) InvokeWithMessages(ctx context.Context, msgs []message.Message) ([]message.Message, error) {
	messages, err := convertToOpenAIMessages(msgs)
	if err != nil {
		return nil, err
	}

	body := llm.body
	body.Messages = messages

	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		body.Messages = messages

		completion, err := llm.client.Chat.Completions.New(ctx, body)
		if err != nil {
			return nil, err
		}

		if len(completion.Choices) == 0 {
			return nil, errors.New("no choices returned from LLM")
		}

		choice := completion.Choices[0]

		if toolsCalls := choice.Message.ToolCalls; len(toolsCalls) > 0 {
			messages = append(messages, choice.Message.ToParam())

			for _, toolCall := range toolsCalls {
				toolName := toolCall.Function.Name
				tool, ok := llm.tools[toolName]
				if !ok {
					return nil, errors.New("unknown tool called: " + toolName)
				}

				var params map[string]any
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
					return nil, err
				}

				// Call the tool with the parameters
				result, err := tool.Call(ctx, params)
				if err != nil {
					return nil, err
				}

				messages = append(messages, openai.ToolMessage(result, toolCall.ID))
			}

			continue // re-evaluate with updated messages
		}

		messages = append(messages, choice.Message.ToParam())

		msgs := make([]message.Message, len(messages))
		for i, msg := range messages {
			m, err := convertToMessage(msg)
			if err != nil {
				return nil, err
			}

			msgs[i] = m
		}

		return msgs, nil
	}

	return nil, errors.New("max iterations reached without valid response")
}

func convertToOpenAIMessages(msgs []message.Message) ([]openai.ChatCompletionMessageParamUnion, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, len(msgs))
	for i, msg := range msgs {
		var m openai.ChatCompletionMessageParamUnion

		switch msg.Role {
		case message.RoleSystem:
			m = openai.SystemMessage(msg.Content)

		case message.RoleHuman:
			m = openai.UserMessage(msg.Content)

		case message.RoleAI:
			m = openai.AssistantMessage(msg.Content)

			toolCalls := make([]openai.ChatCompletionMessageToolCallParam, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				bs, err := json.Marshal(tc.Arguments)
				if err != nil {
					return nil, err
				}

				toolCall := openai.ChatCompletionMessageToolCallParam{
					ID: tc.ID,
					Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      tc.Name,
						Arguments: string(bs),
					},
				}

				toolCalls[j] = toolCall
			}

		case message.RoleTool:
			m = openai.ToolMessage(msg.Content, msg.ToolCallID)

		default:
			continue // skip unknown roles
		}

		messages[i] = m
	}

	return messages, nil
}

func convertToMessage(msg openai.ChatCompletionMessageParamUnion) (message.Message, error) {
	var m message.Message

	switch {
	case msg.OfSystem != nil:
		m = message.Message{
			Role:    message.RoleSystem,
			Content: msg.OfSystem.Content.OfString.Value,
		}

	case msg.OfUser != nil:
		m = message.Message{
			Role:    message.RoleHuman,
			Content: msg.OfUser.Content.OfString.Value,
		}

	case msg.OfAssistant != nil:
		m = message.Message{
			Role:    message.RoleAI,
			Content: msg.OfAssistant.Content.OfString.Value,
		}

		toolCalls := make([]message.ToolCall, len(msg.OfAssistant.ToolCalls))
		for i, tc := range msg.OfAssistant.ToolCalls {
			args := make(map[string]any)

			err := json.Unmarshal([]byte(tc.Function.Arguments), &args)
			if err != nil {
				return message.Message{}, err
			}

			toolCalls[i] = message.ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: args,
			}
		}

		m.ToolCalls = toolCalls

	case msg.OfTool != nil:
		m = message.Message{
			Role:       message.RoleTool,
			Content:    msg.OfTool.Content.OfString.Value,
			ToolCallID: msg.OfTool.ToolCallID,
		}

	default:
		return message.Message{}, errors.New("unknown message type")
	}

	return m, nil
}
