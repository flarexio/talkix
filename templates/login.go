package templates

import (
	"fmt"
	"text/template"
)

func LoginTemplate(authURL string) *template.Template {
	flex := `
	{
	  "type": "bubble",
	  "body": {
	    "type": "box",
	    "layout": "vertical",
	    "contents": [
	      {
	        "type": "text",
	        "text": "{{ .Title }}",
	        "weight": "bold",
	        "size": "lg",
	        "align": "center",
	        "margin": "md"
	      },
	      {
	        "type": "text",
	        "text": "{{ .Description }}",
	        "size": "sm",
	        "color": "#888888",
	        "align": "center",
	        "wrap": true,
	        "margin": "md"
	      },
	      {
	        "type": "button",
	        "margin": "lg",
	        "action": {
	          "type": "uri",
	          "label": "Login with LINE",
	          "uri": "%s"
	        },
	        "style": "primary",
	        "color": "#1DB446"
	      }
	    ]
	  }
	}`

	flex = fmt.Sprintf(flex, authURL)

	tmpl, err := template.New("login").Parse(flex)
	if err != nil {
		panic(err.Error())
	}

	return tmpl
}

var LoginValuesSchema = map[string]any{
	"type":        []string{"object", "null"},
	"description": "Values for the login template",
	"properties": map[string]any{
		"Title":       map[string]any{"type": "string"},
		"Description": map[string]any{"type": "string"},
	},
	"required":             []string{"Title", "Description"},
	"additionalProperties": false,
}
