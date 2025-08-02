package message

import (
	"encoding/json"
	"fmt"
)

type Role string

const (
	RoleSystem Role = "system"
	RoleHuman  Role = "human"
	RoleAI     Role = "ai"
	RoleTool   Role = "tool"
)

func (r Role) String() string {
	switch r {
	case RoleSystem:
		return "System"
	case RoleHuman:
		return "Human"
	case RoleAI:
		return "AI"
	case RoleTool:
		return "Tool"
	default:
		return "Unknown"
	}
}

type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

func (m *Message) PrettyPrint() {
	fmt.Printf("--- %sMessage ---\n", m.Role)

	if m.ToolCallID != "" {
		fmt.Printf("Tool Call ID: %s\n", m.ToolCallID)
	}

	if bs := []byte(m.Content); !json.Valid(bs) {
		fmt.Printf("%s\n", m.Content)
	} else {
		var contentMap map[string]any
		if err := json.Unmarshal(bs, &contentMap); err != nil {
			fmt.Printf("%s\n", m.Content)
		} else {
			contentJSON, _ := json.MarshalIndent(contentMap, "", "  ")
			fmt.Printf("%s\n", contentJSON)
		}
	}

	if len(m.ToolCalls) > 0 {
		fmt.Printf("Tool Calls:\n")
		for _, tc := range m.ToolCalls {
			fmt.Printf("- Name: %s (id: %s)\n", tc.Name, tc.ID)
			fmt.Println("  Arguments:")

			argsJSON, err := json.MarshalIndent(tc.Arguments, "", "  ")
			if err != nil {
				fmt.Printf("%v\n", tc.Arguments)
			} else {
				fmt.Printf("%s\n", argsJSON)
			}
		}
	}

	fmt.Print("-------------------\n\n")
}

type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"args"`
}

func SystemMessage(content string) Message {
	return Message{
		Role:    RoleSystem,
		Content: content,
	}
}

func HumanMessage(content string) Message {
	return Message{
		Role:    RoleHuman,
		Content: content,
	}
}

func AIMessage(content string, toolCalls ...ToolCall) Message {
	return Message{
		Role:      RoleAI,
		Content:   content,
		ToolCalls: toolCalls,
	}
}
