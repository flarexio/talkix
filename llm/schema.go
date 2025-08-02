package llm

type Schema interface {
	Name() string
	Description() string
	Schema() map[string]any
}
