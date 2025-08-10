package templates

import "text/template"

func WeatherTemplate() *template.Template {
	flex := `
	{
	  "type": "bubble",
	  "hero": {
	    "type": "image",
	    "url": "{{ .IconURL }}",
	    "size": "sm",
	    "aspectRatio": "1:1",
	    "aspectMode": "fit"
	  },
	  "body": {
	    "type": "box",
	    "layout": "vertical",
	    "spacing": "md",
	    "contents": [
	      {
	        "type": "text",
	        "text": "{{ .Location }}",
	        "weight": "bold",
	        "size": "lg",
	        "align": "center"
	      },
	      {
	        "type": "text",
	        "text": "{{ .Condition }}",
	        "size": "md",
	        "color": "#1DB446",
	        "align": "center"
	      },
	      {
	        "type": "box",
	        "layout": "horizontal",
	        "contents": [
	          {
	            "type": "text",
	            "text": "溫度",
	            "size": "sm",
	            "color": "#888888"
	          },
	          {
	            "type": "text",
	            "text": "{{ .Temperature }}",
	            "size": "sm",
	            "align": "end"
	          }
	        ]
	      },
	      {
	        "type": "box",
	        "layout": "horizontal",
	        "contents": [
	          {
	            "type": "text",
	            "text": "體感",
	            "size": "sm",
	            "color": "#888888"
	          },
	          {
	            "type": "text",
	            "text": "{{ .FeelsLike }}",
	            "size": "sm",
	            "align": "end"
	          }
	        ]
	      },
	      {
	        "type": "box",
	        "layout": "horizontal",
	        "contents": [
	          {
	            "type": "text",
	            "text": "濕度",
	            "size": "sm",
	            "color": "#888888"
	          },
	          {
	            "type": "text",
	            "text": "{{ .Humidity }}",
	            "size": "sm",
	            "align": "end"
	          }
	        ]
	      },
	      {
	        "type": "box",
	        "layout": "horizontal",
	        "contents": [
	          {
	            "type": "text",
	            "text": "風速",
	            "size": "sm",
	            "color": "#888888"
	          },
	          {
	            "type": "text",
	            "text": "{{ .WindSpeed }}",
	            "size": "sm",
	            "align": "end"
	          }
	        ]
	      },
	      {
	        "type": "box",
	        "layout": "vertical",
	        "margin": "md",
	        "contents": [
	          {
	            "type": "text",
	            "text": "{{ .ExtraInfo }}",
	            "size": "sm",
	            "color": "#0055FF",
	            "align": "start",
	            "wrap": true
	          }
	        ]
	      }
	    ]
	  },
	  "footer": {
	    "type": "box",
	    "layout": "vertical",
	    "contents": [
	      {
	        "type": "text",
	        "text": "Last updated {{ .LastUpdated }}",
	        "size": "xs",
	        "color": "#aaaaaa",
	        "align": "center"
	      }
	    ]
	  }
	}`

	tmpl, err := template.New("weather").Parse(flex)
	if err != nil {
		panic(err.Error())
	}

	return tmpl
}

var WeatherValuesSchema = map[string]any{
	"type":        []string{"object", "null"},
	"description": "Values for the weather template",
	"properties": map[string]any{
		"Location":    map[string]any{"type": "string"},
		"Condition":   map[string]any{"type": "string"},
		"IconURL":     map[string]any{"type": "string"},
		"Temperature": map[string]any{"type": "string"},
		"FeelsLike":   map[string]any{"type": "string"},
		"Humidity":    map[string]any{"type": "string"},
		"WindSpeed":   map[string]any{"type": "string"},
		"LastUpdated": map[string]any{"type": "string"},
		"ExtraInfo":   map[string]any{"type": "string"},
	},
	"required": []string{
		"Location", "Condition", "IconURL",
		"Temperature", "FeelsLike", "Humidity", "WindSpeed",
		"LastUpdated", "ExtraInfo",
	},
	"additionalProperties": false,
}
