package talkix

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flarexio/talkix/config"
	"github.com/flarexio/talkix/llm"
	"github.com/flarexio/talkix/llm/message"
)

func TestLLMWithLineMessage(t *testing.T) {
	assert := assert.New(t)

	llm, err := llm.NewLLM("openai:gpt-4.1-mini",
		llm.WithStructuredOutput(LineMessage{}),
	)

	if err != nil {
		assert.Fail(err.Error())
		return
	}

	msgs := []message.Message{
		message.HumanMessage("What is the capital of France?"),
	}

	ctx := context.Background()
	msgs, err = llm.Invoke(ctx, msgs)
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	resp := msgs[len(msgs)-1]

	var result LineMessage
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
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

	cfg := config.WeatherAPIConfig{
		BaseURL: "https://api.openweathermap.org",
		APIKey:  apiKey,
		Timeout: 10 * time.Second,
	}

	tools := []llm.Tool{
		NewWeatherTool(cfg),
	}

	llm, err := llm.NewLLM("openai:gpt-4.1-mini",
		llm.WithTools(tools),
	)

	if err != nil {
		assert.Fail(err.Error())
		return
	}

	msgs := []message.Message{
		message.HumanMessage("What is the weather in New York City?"),
	}

	ctx := context.Background()
	msgs, err = llm.Invoke(ctx, msgs)
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	resp := msgs[len(msgs)-1]

	result := resp.Content
	assert.NotEmpty(result, "Expected result to not be empty")
}
