package talkix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/yaml.v3"
)

// WeatherAPIConfig 天氣 API 配置
type WeatherAPIConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

func (cfg *WeatherAPIConfig) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		APIKey  string `yaml:"apiKey"`
		BaseURL string `yaml:"baseURL"`
		Timeout string `yaml:"timeout"`
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	cfg.APIKey = raw.APIKey

	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openweathermap.org"
	}

	cfg.Timeout = 10 * time.Second
	if raw.Timeout != "" {
		duration, err := time.ParseDuration(raw.Timeout)
		if err != nil {
			return err
		}

		cfg.Timeout = duration
	}

	return nil
}

// OpenWeatherMap 3.0 API 回應結構
type OpenWeatherResponse struct {
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Timezone string  `json:"timezone"`
	Current  struct {
		Dt         int64   `json:"dt"`
		Sunrise    int64   `json:"sunrise"`
		Sunset     int64   `json:"sunset"`
		Temp       float64 `json:"temp"`
		FeelsLike  float64 `json:"feels_like"`
		Pressure   int     `json:"pressure"`
		Humidity   int     `json:"humidity"`
		DewPoint   float64 `json:"dew_point"`
		UVI        float64 `json:"uvi"`
		Clouds     int     `json:"clouds"`
		Visibility int     `json:"visibility"`
		WindSpeed  float64 `json:"wind_speed"`
		WindDeg    int     `json:"wind_deg"`
		WindGust   float64 `json:"wind_gust,omitempty"`
		Weather    []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
	} `json:"current"`
}

// Geocoding API 回應結構 (用於獲取座標)
type GeocodingResponse []struct {
	Name       string            `json:"name"`
	LocalNames map[string]string `json:"local_names,omitempty"`
	Lat        float64           `json:"lat"`
	Lon        float64           `json:"lon"`
	Country    string            `json:"country"`
	State      string            `json:"state,omitempty"`
}

// 增強的天氣資料結構
type WeatherData struct {
	Location    string  `json:"location"`
	Country     string  `json:"country"`
	State       string  `json:"state,omitempty"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	Temperature float64 `json:"temperature"`
	FeelsLike   float64 `json:"feels_like"`
	Condition   string  `json:"condition"`
	Humidity    int     `json:"humidity"`
	Pressure    int     `json:"pressure"`
	WindSpeed   float64 `json:"wind_speed"`
	WindDeg     int     `json:"wind_deg"`
	Clouds      int     `json:"clouds"`
	UVI         float64 `json:"uvi"`
	Visibility  int     `json:"visibility"`
	IconURL     string  `json:"icon_url"`
	Sunrise     string  `json:"sunrise"`
	Sunset      string  `json:"sunset"`
	LastUpdated string  `json:"last_updated"`
}

func NewWeatherTool(config WeatherAPIConfig) Tool {
	return &weatherTool{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

type weatherTool struct {
	config WeatherAPIConfig
	client *http.Client
}

func (tool *weatherTool) Name() string {
	return "get_weather"
}

func (tool *weatherTool) Description() string {
	return "Get current weather information by latitude and longitude. Returns detailed weather data including temperature, weather conditions, humidity, wind speed, pressure, UV index, cloud coverage, visibility, sunrise and sunset times, and last updated time."
}

func (tool *weatherTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"latitude": map[string]any{
				"type":        "number",
				"description": "Latitude of the location (e.g., 25.0330)",
			},
			"longitude": map[string]any{
				"type":        "number",
				"description": "Longitude of the location (e.g., 121.5654)",
			},
			"units": map[string]any{
				"type":        "string",
				"description": "Temperature units: 'celsius', 'fahrenheit', or 'kelvin'",
				"enum":        []string{"celsius", "fahrenheit", "kelvin"},
				"default":     "celsius",
			},
		},
		"required": []string{"latitude", "longitude"},
	}
}

func (tool *weatherTool) Call(ctx context.Context, params map[string]any) (string, error) {
	lat, latOK := params["latitude"].(float64)
	lon, lonOK := params["longitude"].(float64)
	if !latOK || !lonOK {
		return "", errors.New("latitude and longitude parameters are required and must be numbers")
	}

	units := "metric"
	if unitsParam, ok := params["units"].(string); ok {
		switch unitsParam {
		case "fahrenheit":
			units = "imperial"
		case "kelvin":
			units = "standard"
		case "celsius":
			units = "metric"
		}
	}

	weatherResp, err := tool.FetchWeatherFromAPI(ctx, lat, lon, units)
	if err != nil {
		return "", fmt.Errorf("failed to fetch weather: %w", err)
	}

	weatherData := &WeatherData{
		Location:    "",
		Country:     "",
		State:       "",
		Latitude:    lat,
		Longitude:   lon,
		Timezone:    weatherResp.Timezone,
		Temperature: weatherResp.Current.Temp,
		FeelsLike:   weatherResp.Current.FeelsLike,
		Humidity:    weatherResp.Current.Humidity,
		Pressure:    weatherResp.Current.Pressure,
		WindSpeed:   weatherResp.Current.WindSpeed,
		WindDeg:     weatherResp.Current.WindDeg,
		Clouds:      weatherResp.Current.Clouds,
		UVI:         weatherResp.Current.UVI,
		Visibility:  weatherResp.Current.Visibility,
		Sunrise:     time.Unix(weatherResp.Current.Sunrise, 0).Format("15:04"),
		Sunset:      time.Unix(weatherResp.Current.Sunset, 0).Format("15:04"),
		LastUpdated: time.Unix(weatherResp.Current.Dt, 0).Format("2006-01-02 15:04:05"),
	}
	if len(weatherResp.Current.Weather) > 0 {
		weather := weatherResp.Current.Weather[0]
		weatherData.Condition = weather.Description
		weatherData.IconURL = fmt.Sprintf("https://openweathermap.org/img/wn/%s@2x.png", weather.Icon)
	}

	result, err := json.Marshal(weatherData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal weather data: %w", err)
	}
	return string(result), nil
}

// GeoLocation 地理位置資料
type GeoLocation struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	State     string  `json:"state,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (tool *weatherTool) GetCoordinates(ctx context.Context, location string) (*GeoLocation, error) {
	// 使用 Geocoding API
	geoURL := fmt.Sprintf("%s/geo/1.0/direct", tool.config.BaseURL)
	u, err := url.Parse(geoURL)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	query.Set("q", location)
	query.Set("limit", "1")
	query.Set("appid", tool.config.APIKey)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := tool.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}

	var geoResp GeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return nil, err
	}

	if len(geoResp) == 0 {
		return nil, errors.New("location not found")
	}

	loc := geoResp[0]
	return &GeoLocation{
		Name:      loc.Name,
		Country:   loc.Country,
		State:     loc.State,
		Latitude:  loc.Lat,
		Longitude: loc.Lon,
	}, nil
}

func (tool *weatherTool) FetchWeatherData(ctx context.Context, location, units string) (*WeatherData, error) {
	if tool.config.APIKey == "" {
		return nil, errors.New("weather API key is not configured")
	}

	// 第一步：使用 Geocoding API 獲取座標
	geoLoc, err := tool.GetCoordinates(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("failed to get coordinates: %w", err)
	}

	// 第二步：使用 One Call API 3.0 獲取天氣資料
	weatherResp, err := tool.FetchWeatherFromAPI(ctx, geoLoc.Latitude, geoLoc.Longitude, units)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather: %w", err)
	}

	// 轉換為標準格式
	weatherData := &WeatherData{
		Location:    geoLoc.Name,
		Country:     geoLoc.Country,
		State:       geoLoc.State,
		Latitude:    geoLoc.Latitude,
		Longitude:   geoLoc.Longitude,
		Timezone:    weatherResp.Timezone,
		Temperature: weatherResp.Current.Temp,
		FeelsLike:   weatherResp.Current.FeelsLike,
		Humidity:    weatherResp.Current.Humidity,
		Pressure:    weatherResp.Current.Pressure,
		WindSpeed:   weatherResp.Current.WindSpeed,
		WindDeg:     weatherResp.Current.WindDeg,
		Clouds:      weatherResp.Current.Clouds,
		UVI:         weatherResp.Current.UVI,
		Visibility:  weatherResp.Current.Visibility,
		Sunrise:     time.Unix(weatherResp.Current.Sunrise, 0).Format("15:04"),
		Sunset:      time.Unix(weatherResp.Current.Sunset, 0).Format("15:04"),
		LastUpdated: time.Unix(weatherResp.Current.Dt, 0).Format("2006-01-02 15:04:05"),
	}

	if len(weatherResp.Current.Weather) > 0 {
		weather := weatherResp.Current.Weather[0]
		weatherData.Condition = weather.Description
		weatherData.IconURL = fmt.Sprintf("https://openweathermap.org/img/wn/%s@2x.png", weather.Icon)
	}

	return weatherData, nil
}

func (tool *weatherTool) FetchWeatherFromAPI(ctx context.Context, lat, lon float64, units string) (*OpenWeatherResponse, error) {
	// 使用 One Call API 3.0
	weatherURL := fmt.Sprintf("%s/data/3.0/onecall", tool.config.BaseURL)
	u, err := url.Parse(weatherURL)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	query.Set("lat", fmt.Sprintf("%.6f", lat))
	query.Set("lon", fmt.Sprintf("%.6f", lon))
	query.Set("appid", tool.config.APIKey)
	query.Set("units", units)
	query.Set("exclude", "minutely,hourly,daily,alerts") // 只要當前天氣
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := tool.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var weatherResp OpenWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, err
	}

	return &weatherResp, nil
}
