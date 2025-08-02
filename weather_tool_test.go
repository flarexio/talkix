package talkix

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flarexio/talkix/config"
)

func TestWeatherTool(t *testing.T) {
	assert := assert.New(t)

	apiKey, ok := os.LookupEnv("WEATHER_API_KEY")
	if !ok {
		t.Skip("WEATHER_API_KEY environment variable is not set")
		return
	}

	cfg := config.WeatherAPIConfig{
		APIKey:  apiKey,
		BaseURL: "https://api.openweathermap.org",
		Timeout: 10 * time.Second,
	}

	tool := NewWeatherTool(cfg)
	weatherTool := tool.(*weatherTool)

	ctx := context.Background()
	data, err := weatherTool.FetchWeatherData(ctx, "Taichung", "metric")
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	fmt.Println(data)
}
