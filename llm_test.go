package talkix

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLLM(t *testing.T) {
	assert := assert.New(t)

	llm, err := NewLLM("openai:gpt-4.1-mini",
		WithPrompt("You are a helpful assistant."),
	)

	if err != nil {
		assert.Fail(err.Error())
		return
	}

	var result string

	ctx := context.Background()
	if err := llm.Invoke(ctx, "What is the capital of France?", &result); err != nil {
		assert.Fail(err.Error())
		return
	}

	assert.Contains(result, "Paris", "Expected result to contain 'Paris'")
}

func TestLLMWithStructuredOutput(t *testing.T) {
	assert := assert.New(t)

	llm, err := NewLLM("openai:gpt-4.1-mini",
		WithPrompt("You are a helpful assistant."),
		WithStructuredOutput(LineMessage{}),
	)

	if err != nil {
		assert.Fail(err.Error())
		return
	}

	var result LineMessage

	ctx := context.Background()
	if err := llm.Invoke(ctx, "What is the capital of France?", &result); err != nil {
		assert.Fail(err.Error())
		return
	}

	assert.Equal("text", result.Type, "Expected message type to be 'text'")
	assert.Contains(result.Text.Text, "Paris", "Expected result to contain 'Paris'")
}

func TestLLMWithTools(t *testing.T) {
	assert := assert.New(t)

	apiKey, ok := os.LookupEnv("WEATHER_API_KEY")
	if !ok {
		t.Skip("WEATHER_API_KEY environment variable is not set")
		return
	}

	cfg := WeatherAPIConfig{
		BaseURL: "https://api.openweathermap.org",
		APIKey:  apiKey,
		Timeout: 10 * time.Second,
	}

	tools := []Tool{
		NewWeatherTool(cfg),
	}

	llm, err := NewLLM("openai:gpt-4.1-mini",
		WithPrompt("You are a helpful assistant."),
		WithTools(tools),
	)

	if err != nil {
		assert.Fail(err.Error())
		return
	}

	var result string

	ctx := context.Background()
	if err := llm.Invoke(ctx, "What is the weather in New York City?", &result); err != nil {
		assert.Fail(err.Error())
		return
	}

	assert.NotEmpty(result, "Expected result to not be empty")
}

func TestLLMWithToolsAndLineMessage(t *testing.T) {
	assert := assert.New(t)

	apiKey, ok := os.LookupEnv("WEATHER_API_KEY")
	if !ok {
		t.Skip("WEATHER_API_KEY environment variable is not set")
		return
	}

	cfg := WeatherAPIConfig{
		BaseURL: "https://api.openweathermap.org",
		APIKey:  apiKey,
		Timeout: 10 * time.Second,
	}

	tools := []Tool{
		NewWeatherTool(cfg),
	}

	llm, err := NewLLM("openai:gpt-4.1-mini",
		WithPrompt(`You are a helpful assistant for LINE messaging platform.

IMPORTANT: Return exactly ONE JSON object, not multiple objects.

Instructions:
1. For simple responses, use type "text" with content field filled and flexContent as empty string
2. For rich information, use type "flex" with flexContent field containing the Flex Message JSON and content as a brief description

Choose the most appropriate format and return ONLY ONE JSON object.

For weather information with visual request, use Flex Message format.`),
		WithTools(tools),
		WithStructuredOutput(LineMessage{}),
	)

	if err != nil {
		assert.Fail(err.Error())
		return
	}

	var result LineMessage

	ctx := context.Background()
	if err := llm.Invoke(ctx, "Show me the weather in New York City with a nice visual card format", &result); err != nil {
		assert.Fail(err.Error())
		return
	}

	assert.Equal("flex", result.Type, "Expected message type to be 'flex'")
	assert.NotEmpty(result.Flex.Flex, "Expected flexContent to be filled for rich content")
}
