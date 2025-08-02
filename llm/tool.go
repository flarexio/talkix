package llm

import "context"

type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	Call(ctx context.Context, params map[string]any) (string, error)
}
