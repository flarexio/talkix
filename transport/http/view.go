package http

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/flarexio/talkix/user"
)

//go:embed sessions.html
var tmplFS embed.FS

var sessionsPageTmpl = template.Must(template.ParseFS(tmplFS, "sessions.html"))

func SessionViewHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtx, ok := c.Get("user")
		if !ok {
			c.String(http.StatusInternalServerError, "user not found in context")
			c.Abort()
			return
		}

		u, ok := userCtx.(*user.User)
		if !ok {
			c.String(http.StatusInternalServerError, "invalid user context")
			c.Abort()
			return
		}

		token, ok := c.Get("jwt")
		if !ok {
			c.String(http.StatusInternalServerError, "JWT not found in context")
			c.Abort()
			return
		}

		c.Status(http.StatusOK)
		sessionsPageTmpl.Execute(c.Writer, gin.H{
			"JWT":  token,
			"User": u.Profile.Username,
		})
	}
}
