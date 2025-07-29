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

func RestaurantTemplate() *template.Template {
	flex := `
	{
	  "type": "carousel",
	  "contents": [
	    {{- range $i, $r := . }}
	    {{- if $i}},{{end}}
	    {
	      "type": "bubble",
	      "body": {
	        "type": "box",
	        "layout": "vertical",
	        "contents": [
	          {
	            "type": "text",
	            "text": "{{ $r.Name }}",
	            "weight": "bold",
	            "size": "xl"
	          },
	          {
	            "type": "box",
	            "layout": "baseline",
	            "margin": "md",
	            "contents": [
	              {{- $icons := starIcons $r.Rating -}}
	              {{- range $j, $icon := $icons }}
	              {
	                "type": "icon",
	                "size": "sm",
	                "url": "{{ $icon }}"
	              }{{ if ne (add $j 1) (len $icons) }},{{ end }}
	              {{- end }},
	              {
	                "type": "text",
	                "text": "{{ $r.Rating }}",
	                "size": "sm",
	                "color": "#999999",
	                "margin": "md",
	                "flex": 0
	              }
	            ]
	          },
	          {
	            "type": "box",
	            "layout": "vertical",
	            "margin": "lg",
	            "spacing": "sm",
	            "contents": [
	              {
	                "type": "box",
	                "layout": "baseline",
	                "spacing": "sm",
	                "contents": [
	                  {
	                    "type": "text",
	                    "text": "Place",
	                    "color": "#aaaaaa",
	                    "size": "sm",
	                    "flex": 1
	                  },
	                  {
	                    "type": "text",
	                    "text": "{{ $r.Address }}",
	                    "wrap": true,
	                    "color": "#666666",
	                    "size": "sm",
	                    "flex": 5
	                  }
	                ]
	              },
	              {
	                "type": "box",
	                "layout": "baseline",
	                "spacing": "sm",
	                "contents": [
	                  {
	                    "type": "text",
	                    "text": "Time",
	                    "color": "#aaaaaa",
	                    "size": "sm",
	                    "flex": 1
	                  },
	                  {
	                    "type": "text",
	                    "text": "{{ if $r.OpenTime }}{{ $r.OpenTime }}{{ else }}無營業資訊{{ end }}",
	                    "wrap": true,
	                    "color": "#666666",
	                    "size": "sm",
	                    "flex": 5
	                  }
	                ]
	              }
	            ]
	          }
	        ]
	      },
	      "footer": {
	        "type": "box",
	        "layout": "vertical",
	        "spacing": "sm",
	        "contents": [
	          {
	            "type": "button",
	            "style": "link",
	            "height": "sm",
	            "action": {
	              "type": "uri",
	              "label": "在地圖上查看",
	              "uri": "https://www.google.com/maps/place/?q=place_id:{{ $r.PlaceID }}"
	            }
	          }
	        ]
	      }
	    }
	    {{- end }}
	  ]
	}`

	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"len": func(arr []string) int { return len(arr) },
		"starIcons": func(rating float64) []string {
			gold := "https://developers-resource.landpress.line.me/fx/img/review_gold_star_28.png"
			gray := "https://developers-resource.landpress.line.me/fx/img/review_gray_star_28.png"
			count := int(rating + 0.5)
			icons := make([]string, 5)
			for i := 0; i < 5; i++ {
				if i < count {
					icons[i] = gold
				} else {
					icons[i] = gray
				}
			}
			return icons
		},
	}

	tmpl, err := template.New("restaurant").Funcs(funcMap).Parse(flex)
	if err != nil {
		panic(err.Error())
	}

	return tmpl
}

var RestaurantValuesSchema = map[string]any{
	"type":        []string{"array", "null"},
	"description": "Array of restaurant objects for the restaurant template",
	"items": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"Name": map[string]any{"type": "string"},
			"PlaceID": map[string]any{
				"type":        "string",
				"description": "Google Maps Place ID for the restaurant",
			},
			"Rating":   map[string]any{"type": "number"},
			"Address":  map[string]any{"type": "string"},
			"OpenTime": map[string]any{"type": "string"},
		},
		"required":             []string{"Name", "PlaceID", "Rating", "Address", "OpenTime"},
		"additionalProperties": false,
	},
}
