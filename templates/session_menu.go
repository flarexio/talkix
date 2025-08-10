package templates

import (
	"text/template"
)

func SessionMenuTemplate() *template.Template {
	flex := `{
      "type": "bubble",
      "body": {
        "type": "box",
        "layout": "vertical",
        "contents": [
          {
            "type": "text",
            "text": "🗂️ 會話管理",
            "weight": "bold",
            "size": "lg",
            "align": "center"
          },
          {
            "type": "text",
            "text": "管理您的所有對話會話",
            "size": "sm",
            "color": "#888888",
            "align": "center",
            "margin": "md"
          },
          {
            "type": "separator",
            "margin": "lg"
          },
          {
            "type": "button",
            "action": {
              "type": "uri",
              "label": "📋 查看所有會話",
              "uri": "{{ .ListSessionsURL }}"
            },
            "style": "primary",
            "margin": "lg"
          }
        ]
      },
      "footer": {
        "type": "box",
        "layout": "vertical",
        "contents": [
          {
            "type": "text",
            "text": "⚠️ 僅支援一次性操作",
            "size": "xs",
            "color": "#888888",
            "align": "center"
          }
        ]
      }
    }`

	tmpl, err := template.New("session_menu").Parse(flex)
	if err != nil {
		panic(err.Error())
	}

	return tmpl
}

type SessionMenuValues struct {
	ListSessionsURL string
}

var SessionMenuValuesSchema = map[string]any{
	"type":                 []string{"object", "null"},
	"description":          "Values for the session menu template",
	"properties":           map[string]any{},
	"additionalProperties": false,
	"required":             []string{},
}
