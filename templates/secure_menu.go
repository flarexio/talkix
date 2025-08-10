package templates

import (
	"text/template"
)

func SecureMenuTemplate() *template.Template {
	flex := `{
      "type": "bubble",
      "body": {
        "type": "box",
        "layout": "vertical",
        "contents": [
          {
            "type": "text",
            "text": "🔐 安全操作選單",
            "weight": "bold",
            "size": "lg",
            "align": "center"
          },
          {
            "type": "text",
            "text": "請選擇您要執行的操作",
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
              "label": "👤 查看個人資料",
              "uri": "{{ .ViewProfileURL }}"
            },
            "style": "secondary",
            "margin": "lg"
          },
          {
            "type": "button",
            "action": {
              "type": "uri",
              "label": "⚙️ 編輯設定",
              "uri": "{{ .EditSettingsURL }}"
            },
            "style": "secondary",
            "margin": "sm"
          },
          {
            "type": "button",
            "action": {
              "type": "uri",
              "label": "🗑️ 刪除資料",
              "uri": "{{ .DeleteDataURL }}"
            },
            "style": "secondary",
            "margin": "sm"
          }
        ]
      },
      "footer": {
        "type": "box",
        "layout": "vertical",
        "contents": [
          {
            "type": "text",
            "text": "⚠️ 連結將在 3 分鐘後失效",
            "size": "xs",
            "color": "#888888",
            "align": "center"
          }
        ]
      }
    }`

	tmpl, err := template.New("secure_menu").Parse(flex)
	if err != nil {
		panic(err.Error())
	}

	return tmpl
}

type SecureMenuValues struct {
	ViewProfileURL  string
	EditSettingsURL string
	DeleteDataURL   string
}
