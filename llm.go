package talkix

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/openai/openai-go"
)

func NewLLM(model string, opts ...Option) (*LLM, error) {
	client := openai.NewClient()

	model, ok := strings.CutPrefix(model, "openai:")
	if !ok {
		return nil, errors.New("model must start with 'openai:'")
	}

	llm := &LLM{
		client: client,
		model:  openai.ChatModel(model),
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
	client       openai.Client
	model        openai.ChatModel
	prompt       string
	promptFormat PromptFormat
	schema       Schema
	tools        []Tool
}

type Option interface {
	Apply(*LLM) error
}

func WithPrompt(prompt string) Option {
	return &llmWithPrompt{prompt: prompt}
}

func WithPromptFormat(format PromptFormat) Option {
	return &llmWithPrompt{promptFormat: format}
}

type PromptFormat func(ctx context.Context) (string, error)

type llmWithPrompt struct {
	prompt       string
	promptFormat PromptFormat
}

func (opt *llmWithPrompt) Apply(llm *LLM) error {
	if llm == nil {
		return errors.New("llm cannot be nil")
	}

	llm.prompt = opt.prompt
	llm.promptFormat = opt.promptFormat
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

	llm.schema = opt.schema
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

	llm.tools = opt.tools
	return nil
}

func (llm *LLM) Invoke(ctx context.Context, question string, answer any) error {
	var prompt string
	switch {
	case llm.prompt != "":
		prompt = llm.prompt

	case llm.promptFormat != nil:
		p, err := llm.promptFormat(ctx)
		if err != nil {
			return err
		}

		prompt = p

	default:
		prompt = "You are a helpful assistant."
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(prompt),
		// ...
		openai.UserMessage(question),
	}

	body := openai.ChatCompletionNewParams{
		Model:    llm.model,
		Messages: messages,
	}

	if schema := llm.schema; schema != nil {
		schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        schema.Name(),
			Description: openai.String(schema.Description()),
			Schema:      schema.Schema(),
			Strict:      openai.Bool(true),
		}

		body.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		}
	}

	toolMap := make(map[string]Tool)
	if len(llm.tools) > 0 {
		tools := make([]openai.ChatCompletionToolParam, len(llm.tools))
		for i, tool := range llm.tools {
			tools[i] = openai.ChatCompletionToolParam{
				Function: openai.FunctionDefinitionParam{
					Name:        tool.Name(),
					Description: openai.String(tool.Description()),
					Parameters:  openai.FunctionParameters(tool.Parameters()),
				},
			}

			toolMap[tool.Name()] = tool
		}

		body.Tools = tools
	}

	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		body.Messages = messages

		completion, err := llm.client.Chat.Completions.New(ctx, body)
		if err != nil {
			return err
		}

		if len(completion.Choices) == 0 {
			return errors.New("no choices returned from LLM")
		}

		choice := completion.Choices[0]

		if toolsCalls := choice.Message.ToolCalls; len(toolsCalls) > 0 {
			messages = append(messages, choice.Message.ToParam()) // append AIMessage

			for _, toolCall := range toolsCalls {
				toolName := toolCall.Function.Name
				tool, ok := toolMap[toolName]
				if !ok {
					return errors.New("unknown tool called: " + toolName)
				}

				var params map[string]any
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
					return err
				}

				// Call the tool with the parameters
				result, err := tool.Call(ctx, params)
				if err != nil {
					return err
				}

				messages = append(messages, openai.ToolMessage(result, toolCall.ID)) // append ToolMessage
			}

			continue // re-evaluate with updated messages
		}

		switch v := answer.(type) {
		case *string:
			*v = choice.Message.Content
			return nil

		case *LineMessage:
			var msg LineMessage
			if err := json.Unmarshal([]byte(choice.Message.Content), &msg); err != nil {
				return err
			}

			*v = msg
			return nil

		default:
			return errors.New("unsupported answer type")
		}
	}

	return errors.New("max iterations reached without valid response")
}

type Schema interface {
	Name() string
	Description() string
	Schema() map[string]any
}

type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	Call(ctx context.Context, params map[string]any) (string, error)
}
