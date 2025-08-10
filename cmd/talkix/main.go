package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/flarexio/core/policy"
	"github.com/flarexio/talkix"
	"github.com/flarexio/talkix/auth"
	"github.com/flarexio/talkix/config"
	"github.com/flarexio/talkix/identity"
	"github.com/flarexio/talkix/llm"
	"github.com/flarexio/talkix/persistence/kv"
	"github.com/flarexio/talkix/session"
	"github.com/flarexio/talkix/transport/http"
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

	var cfg config.Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return err
	}

	session.InitLLM(cfg.LLM.Summary.Model)

	otp := auth.NewOTPStore()

	opts := badger.DefaultOptions(filepath.Join(path, cfg.LLM.Persistence.Name))
	if cfg.LLM.Persistence.InMemory {
		opts = badger.DefaultOptions("").WithInMemory(true)
	}

	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	users := kv.NewUserRepository(db)
	sessions := kv.NewSessionRepository(db)

	tools := []llm.Tool{
		talkix.NewWeatherTool(cfg.LLM.Tools.Weather),
	}

	for _, server := range cfg.LLM.Tools.MCPServers {
		mcpTools, err := registerMCPServer(ctx, server)
		if err != nil {
			return err
		}

		tools = append(tools, mcpTools...)
	}

	svc, err := talkix.NewAIService(cfg, tools, otp,
		users, sessions,
	)
	if err != nil {
		return err
	}

	// svc := talkix.NewSimpleService(cfg, otp, users, sessions)

	name := svc.Name()
	svc = talkix.LoggingMiddleware(name)(svc)

	directUser := identity.DirectUserEndpoint(path, cfg.Identity)

	r := gin.Default()
	{
		endpoint := talkix.ReplyMessageEndpoint(svc)

		if err := line.Init(cfg); err != nil {
			return err
		}

		handler := line.MessageHandler(endpoint, directUser)

		r.POST("/webhook/line", handler)
	}

	sessionSvc := talkix.NewSessionService(users, sessions)
	sessionSvc = talkix.SessionLoggingMiddleware()(sessionSvc)

	http.Init(cfg.JWT)

	otpAuth := http.OTPAuthorizator(otp, directUser)
	{
		r.GET("/users/:user/session/list", otpAuth("list_sessions"), http.SessionViewHandler())
	}

	permissionsPath := filepath.Join(path, "permissions.json")
	policy, err := policy.NewRegoPolicy(ctx, permissionsPath)
	if err != nil {
		return err
	}

	jwtAuth := http.JWTAuthorizator(policy, directUser)
	{
		// GET /users/:user/sessions
		{
			endpoint := talkix.ListSessionsEndpoint(sessionSvc)
			r.GET("/users/:user/sessions", jwtAuth("talkix::sessions.read"), http.ListSessionsHandler(endpoint))
		}

		// GET /users/:user/sessions/:session
		{
			endpoint := talkix.SessionEndpoint(sessionSvc)
			r.GET("/users/:user/sessions/:session", jwtAuth("talkix::sessions.read"), http.SessionHandler(endpoint))
		}

		// POST /users/:user/sessions
		{
			endpoint := talkix.CreateSessionEndpoint(sessionSvc)
			r.POST("/users/:user/sessions", jwtAuth("talkix::sessions.create"), http.CreateSessionHandler(endpoint))
		}

		// PATCH /users/:user/sessions/:session
		{
			endpoint := talkix.SwitchSessionEndpoint(sessionSvc)
			r.PATCH("/users/:user/sessions/:session", jwtAuth("talkix::sessions.update"), http.SwitchSessionHandler(endpoint))
		}

		// DELETE /users/:user/sessions/:session
		{
			endpoint := talkix.DeleteSessionEndpoint(sessionSvc)
			r.DELETE("/users/:user/sessions/:session", jwtAuth("talkix::sessions.delete"), http.DeleteSessionHandler(endpoint))
		}
	}

	go r.Run(":" + strconv.Itoa(cmd.Int("port")))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sign := <-quit

	log.Info("graceful shutdown", zap.String("signal", sign.String()))
	return nil
}

func NewMCPTool(tool mcp.Tool, fn callToolFn) llm.Tool {
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

func registerMCPServer(ctx context.Context, cfg config.MCPServerConfig) ([]llm.Tool, error) {
	var (
		c   *client.Client
		err error
	)

	switch cfg.Transport {
	case config.TransportTypeStdio:
		c, err = client.NewStdioMCPClient(
			cfg.Command,
			cfg.Environment,
			cfg.Arguments...,
		)

	case config.TransportTypeSSE:
		c, err = client.NewSSEMCPClient(cfg.URL)

	case config.TransportTypeStreamableHTTP:
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

	tools := make([]llm.Tool, 0)

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
