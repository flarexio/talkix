package talkix

type Config struct {
	Line     LineConfig     `yaml:"line"`
	Identity IdentityConfig `yaml:"identity"`
	Tools    ToolsConfig    `yaml:"tools"`
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
