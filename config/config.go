package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Line     LineConfig     `yaml:"line"`
	Identity IdentityConfig `yaml:"identity"`
	LLM      LLMConfig      `yaml:"llm"`
}

type LineConfig struct {
	Messaging struct {
		ChannelToken  string `yaml:"channelToken"`
		ChannelSecret string `yaml:"channelSecret"`
		WebhookURL    string `yaml:"webhookURL"`
	} `yaml:"messaging"`
	Login struct {
		AuthURL string `yaml:"authURL"`
	} `yaml:"login"`
}

type IdentityConfig struct {
	ServerURL string `yaml:"serverURL"`
	CaFile    string `yaml:"caFile"`
	CertFile  string `yaml:"certFile"`
	KeyFile   string `yaml:"keyFile"`
}

type LLMConfig struct {
	Model       string      `yaml:"model"`
	Prompt      string      `yaml:"prompt"`
	Persistence Persistence `yaml:"persistence"`
	Tools       ToolsConfig `yaml:"tools"`
}

type PersistenceDriver string

const (
	InMemory PersistenceDriver = "inmem"
	Badger   PersistenceDriver = "badger"
	SQLite   PersistenceDriver = "sqlite"
)

type Persistence struct {
	Driver   PersistenceDriver `yaml:"driver"`
	Name     string            `yaml:"name"`
	Host     string            `yaml:"host"`
	Port     int               `yaml:"port"`
	Username string            `yaml:"username"`
	Password string            `yaml:"password"`
	InMemory bool              `yaml:"inmem"`
}

type ToolsConfig struct {
	MCPServers map[string]MCPServerConfig `yaml:"mcpServers"`
	Weather    WeatherAPIConfig           `yaml:"weather"`
}

type TransportType string

const (
	TransportTypeStdio          TransportType = "stdio"
	TransportTypeSSE            TransportType = "sse"
	TransportTypeStreamableHTTP TransportType = "streamable-http"
	TransportTypeNATS           TransportType = "nats"
)

type MCPServerConfig struct {
	Transport   TransportType `yaml:"transport"`
	Command     string        `yaml:"command"`
	URL         string        `yaml:"url"`
	Arguments   []string      `yaml:"args"`
	Environment []string      `yaml:"env"`
}

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
