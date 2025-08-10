package llm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLLM(t *testing.T) {
	assert := assert.New(t)

	llm, err := NewLLM("openai:gpt-4.1-mini")
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	ctx := context.Background()
	msgs, err := llm.Invoke(ctx, "What is the capital of France?")
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	resp := msgs[len(msgs)-1]

	result := resp.Content
	assert.Contains(result, "Paris", "Expected result to contain 'Paris'")
}

type example struct {
	City string `json:"city"`
}

func (e example) Name() string {
	return "example"
}

func (e example) Description() string {
	return "An example tool for testing"
}

func (e example) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"city": map[string]any{
				"type": "string",
			},
		},
		"required":             []string{"city"},
		"additionalProperties": false,
	}
}

func TestLLMWithStructuredOutput(t *testing.T) {
	assert := assert.New(t)

	llm, err := NewLLM("openai:gpt-4.1-mini",
		WithStructuredOutput(example{}),
	)

	if err != nil {
		assert.Fail(err.Error())
		return
	}

	ctx := context.Background()
	msgs, err := llm.Invoke(ctx, "What is the capital of France?")
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	resp := msgs[len(msgs)-1]

	var result example
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		assert.Fail(err.Error())
		return
	}

	assert.Contains(result.City, "Paris", "Expected result to contain 'Paris'")
}
