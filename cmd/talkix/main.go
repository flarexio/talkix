package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/flarexio/talkix"
	"github.com/flarexio/talkix/transport/line"
)

func main() {
	cmd := &cli.Command{
		Name:  "talkix",
		Usage: "Run the Talkix service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "Path to the talkix file",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "Port to run the service on",
				Value: 8080,
			},
		},
		Action: run,
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	path := cmd.String("path")
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		path = filepath.Join(homeDir, ".flarex", "talkix")
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer log.Sync()

	zap.ReplaceGlobals(log)

	f, err := os.Open(filepath.Join(path, "config.yaml"))
	if err != nil {
		return err
	}
	defer f.Close()

	var cfg talkix.Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return err
	}

	tools := []talkix.Tool{
		talkix.NewWeatherTool(cfg.Tools.Weather),
	}

	for _, server := range cfg.Tools.MCPServers {
		mcpTools, err := registerMCPServer(ctx, server)
		if err != nil {
			return err
		}

		tools = append(tools, mcpTools...)
	}

	svc, err := talkix.NewAIService(cfg.Line, tools)
	if err != nil {
		return err
	}

	svc = talkix.LoggingMiddleware("ai")(svc)

	endpoint := talkix.ReplyMessageEndpoint(svc)

	if err := line.Init(cfg.Line); err != nil {
		return err
	}

	directIdentityUserEndpoint, err := DirectIdentityUserEndpoint(path, cfg.Identity)
	if err != nil {
		return err
	}

	handler := line.MessageHandler(endpoint, directIdentityUserEndpoint)

	r := gin.Default()
	r.POST("/webhook/line", handler)

	go r.Run(":" + strconv.Itoa(cmd.Int("port")))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sign := <-quit

	log.Info("graceful shutdown", zap.String("signal", sign.String()))
	return nil
}

func NewMCPTool(tool mcp.Tool, fn callToolFn) talkix.Tool {
	return &mcpTool{tool, fn}
}

type callToolFn func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)

type mcpTool struct {
	tool mcp.Tool
	fn   callToolFn
}

func (t *mcpTool) Name() string {
	return t.tool.Name
}

func (t *mcpTool) Description() string {
	return t.tool.Description
}

func (t *mcpTool) Parameters() map[string]any {
	params := make(map[string]any)
	params["type"] = t.tool.InputSchema.Type
	params["properties"] = t.tool.InputSchema.Properties
	params["required"] = t.tool.InputSchema.Required
	return params
}

func (t *mcpTool) Call(ctx context.Context, params map[string]any) (string, error) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      t.tool.Name,
			Arguments: params,
		},
	}

	result, err := t.fn(ctx, req)
	if err != nil {
		return "", err
	}

	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		return "", errors.New("unexpected content type")
	}

	return content.Text, nil
}

func registerMCPServer(ctx context.Context, cfg talkix.MCPServerConfig) ([]talkix.Tool, error) {
	var (
		c   *client.Client
		err error
	)

	switch cfg.Transport {
	case talkix.TransportTypeStdio:
		c, err = client.NewStdioMCPClient(
			cfg.Command,
			cfg.Environment,
			cfg.Arguments...,
		)

	case talkix.TransportTypeSSE:
		c, err = client.NewSSEMCPClient(cfg.URL)

	case talkix.TransportTypeStreamableHTTP:
		c, err = client.NewStreamableHttpClient(cfg.URL)

	default:
		return nil, errors.New("unsupported transport")
	}

	if err != nil {
		return nil, err
	}

	if err := c.Start(ctx); err != nil {
		return nil, err
	}

	req := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "talkix",
				Version: "1.0.0",
			},
		},
	}

	if _, err := c.Initialize(ctx, req); err != nil {
		c.Close()
		return nil, err
	}

	fn := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return c.CallTool(ctx, req)
	}

	tools := make([]talkix.Tool, 0)

	var cursor mcp.Cursor
	for {
		req := mcp.ListToolsRequest{
			PaginatedRequest: mcp.PaginatedRequest{
				Params: mcp.PaginatedParams{
					Cursor: cursor,
				},
			},
		}

		results, err := c.ListTools(ctx, req)
		if err != nil {
			continue
		}

		for _, mcpTool := range results.Tools {
			tool := NewMCPTool(mcpTool, fn)
			tools = append(tools, tool)
		}

		cursor = results.NextCursor
		if cursor == "" {
			break
		}
	}

	if len(tools) == 0 {
		return nil, errors.New("no tools found")
	}

	return tools, nil
}

func DirectIdentityUserEndpoint(path string, cfg talkix.IdentityConfig) (line.DirectIdentityUser, error) {
	certFile := filepath.Join(path, "certs", cfg.CertFile)
	keyFile := filepath.Join(path, "certs", cfg.KeyFile)
	caFile := filepath.Join(path, "certs", cfg.CaFile)

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return func(username string) (*talkix.User, error) {
		serverURL := cfg.ServerURL
		resource := "users/" + username

		resp, err := client.Get(serverURL + "/" + resource)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errMsg, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			return nil, errors.New(string(errMsg))
		}

		var user *talkix.User
		if err := yaml.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, err
		}

		return user, nil
	}, nil
}
