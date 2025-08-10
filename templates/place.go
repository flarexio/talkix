package templates

import "text/template"

func PlaceTemplate() *template.Template {
	flex := `
	{
	  "type": "carousel",
	  "contents": [
	    {{- range $i, $p := . }}
	    {{- if $i}},{{end}}
	    {
	      "type": "bubble",
	      "body": {
	        "type": "box",
	        "layout": "vertical",
	        "contents": [
	          {
	            "type": "text",
	            "text": "{{ $p.Name }}",
	            "weight": "bold",
	            "size": "xl"
	          },
	          {
	            "type": "box",
	            "layout": "baseline",
	            "margin": "md",
	            "contents": [
	              {{- $icons := starIcons $p.Rating -}}
	              {{- range $j, $icon := $icons }}
	              {
	                "type": "icon",
	                "size": "sm",
	                "url": "{{ $icon }}"
	              }{{ if ne (add $j 1) (len $icons) }},{{ end }}
	              {{- end }},
	              {
	                "type": "text",
	                "text": "{{ $p.Rating }}",
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
	                    "text": "{{ $p.Address }}",
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
	              "uri": "https://www.google.com/maps/place/?q=place_id:{{ $p.PlaceID }}"
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

	tmpl, err := template.New("place").Funcs(funcMap).Parse(flex)
	if err != nil {
		panic(err.Error())
	}

	return tmpl
}

var PlaceValuesSchema = map[string]any{
	"type":        []string{"array", "null"},
	"description": "Array of place objects (restaurants, shops, attractions, etc.)",
	"items": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"Name": map[string]any{"type": "string"},
			"PlaceID": map[string]any{
				"type":        "string",
				"description": "Google Maps Place ID for the place",
			},
			"Rating":  map[string]any{"type": "number"},
			"Address": map[string]any{"type": "string"},
		},
		"required":             []string{"Name", "PlaceID", "Rating", "Address"},
		"additionalProperties": false,
	},
}
