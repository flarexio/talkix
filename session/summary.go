package session

import (
	"bytes"
	"context"
	"errors"
	"html/template"

	"github.com/flarexio/talkix/llm"
	"github.com/flarexio/talkix/llm/message"
)

const SYSTEM_PROMPT = `你是專業的對話摘要助手。請根據提供的資訊生成精簡準確的摘要。

要求：
1. 如果有先前摘要，請整合新對話內容更新摘要
2. 如果沒有先前摘要，請根據新對話內容建立摘要
3. 保持與原對話相同的語言
4. 摘要長度限制在50字以內
5. 只保留最核心的話題和關鍵結論
6. 使用簡潔的詞彙，避免冗詞贅字
7. 客觀陳述，不添加個人意見

{{if .PreviousSummary}}
先前摘要：
{{.PreviousSummary}}

{{end}}
本次對話：
{{.CurrentConversation}}

請直接輸出摘要內容，不要額外說明。`

var (
	instance   *llm.LLM
	promptTmpl *template.Template
)

func InitLLM(model string) error {
	tmpl, err := template.New("system_prompt").Parse(SYSTEM_PROMPT)
	if err != nil {
		return err
	}

	llm, err := llm.NewLLM(model)
	if err != nil {
		return err
	}

	instance = llm
	promptTmpl = tmpl
	return nil
}

func GenerateSummary(previousSummary string, conversation *Conversation) (string, error) {
	if instance == nil {
		return "", errors.New("llm not initialized")
	}

	var currentConversation string
	for _, msg := range conversation.Messages {
		if msg.Role == message.RoleHuman || msg.Role == message.RoleAI {
			currentConversation += msg.Role.String() + ": " + msg.Content + "\n"
		}
	}

	data := map[string]any{
		"PreviousSummary":     previousSummary,
		"CurrentConversation": currentConversation,
	}

	var prompt bytes.Buffer
	if err := promptTmpl.Execute(&prompt, data); err != nil {
		return "", err
	}

	messages := []message.Message{
		{
			Role:    message.RoleSystem,
			Content: prompt.String(),
		},
	}

	ctx := context.Background()
	messages, err := instance.InvokeWithMessages(ctx, messages)
	if err != nil {
		return "", err
	}

	if len(messages) == 0 {
		return "", errors.New("no messages returned from LLM")
	}

	summaryMsg := messages[len(messages)-1]
	if summaryMsg.Role != message.RoleAI {
		return "", errors.New("last message is not from AI")
	}

	return summaryMsg.Content, nil
}
