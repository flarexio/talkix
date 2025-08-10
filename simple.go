package talkix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/flarexio/talkix/auth"
	"github.com/flarexio/talkix/config"
	"github.com/flarexio/talkix/llm/message"
	"github.com/flarexio/talkix/session"
	"github.com/flarexio/talkix/templates"
	"github.com/flarexio/talkix/user"
)

func NewSimpleService(cfg config.Config,
	otp *auth.OTPStore,
	users user.Repository, sessions session.Repository,
) Service {
	templates := map[string]*template.Template{
		"login":        templates.LoginTemplate(cfg.Line.Login.AuthURL),
		"secure_menu":  templates.SecureMenuTemplate(),
		"session_menu": templates.SessionMenuTemplate(),
	}

	return &simpleService{
		cfg:       cfg,
		templates: templates,
		otp:       otp,
		users:     users,
		sessions:  sessions,
	}
}

type simpleService struct {
	cfg       config.Config
	templates map[string]*template.Template
	otp       *auth.OTPStore
	users     user.Repository
	sessions  session.Repository
}

func (svc *simpleService) Name() string {
	return "simple"
}

func (svc *simpleService) ReplyMessage(ctx context.Context, msg Message) (Message, error) {
	userCtx, ok := ctx.Value(UserKey).(*user.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}

	u, err := svc.users.Find(userCtx.ID)
	if err != nil {
		u = userCtx
	}

	u.Profile = userCtx.Profile
	u.Verified = userCtx.Verified

	var s *session.Session
	if u.SelectedSessionID == "" {
		s = session.NewSession(u.ID)
		u.AddSessionID(s.ID)
	} else {
		found, err := svc.sessions.Find(u.SelectedSessionID)
		if err != nil {
			return nil, err
		}

		s = found
	}

	m, ok := msg.(*TextMessage)
	if !ok {
		return nil, errors.New("invalid message type")
	}

	c := session.NewConversation()
	c.SetIO(m.Text, "Copy cat: "+m.Text)
	c.AddMessage([]message.Message{
		{
			Role:    message.RoleHuman,
			Content: m.Text,
		},
		{
			Role:    message.RoleAI,
			Content: "Copy cat: " + m.Text,
		},
	}...)
	s.AddConversation(c)

	svc.sessions.Save(s)
	svc.users.Save(u)

	switch m.Text {
	case "LOGIN":
		return svc.handleLogin()

	case "MENU":
		return svc.handleSecureMenu(userCtx.ID)

	case "PROFILE":
		return svc.handleProfileAccess(userCtx.ID)

	case "SESSION":
		return svc.handleSessionMenu(u)

	default:
		return NewTextMessage("Copy cat: " + m.Text), nil
	}
}

func (svc *simpleService) handleLogin() (Message, error) {
	tmpl, ok := svc.templates["login"]
	if !ok {
		return nil, errors.New("login template not found")
	}

	values := map[string]string{
		"Title":       "Please Login to Continue",
		"Description": "You need to login to access this feature.",
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, values); err != nil {
		return nil, err
	}

	return NewFlexMessage(
		"Please Login to Continue",
		buf.Bytes(),
	), nil
}

func (svc *simpleService) handleSecureMenu(userID string) (Message, error) {
	viewOTP, err := svc.otp.GenerateOTP(userID, "view_profile", map[string]any{
		"section": "personal_info",
	})
	if err != nil {
		return nil, err
	}

	editOTP, err := svc.otp.GenerateOTP(userID, "edit_settings", map[string]any{
		"category": "notifications",
	})
	if err != nil {
		return nil, err
	}

	deleteOTP, err := svc.otp.GenerateOTP(userID, "delete_data", map[string]any{
		"confirm_required": true,
	})
	if err != nil {
		return nil, err
	}

	tmpl, ok := svc.templates["secure_menu"]
	if !ok {
		return nil, errors.New("secure_menu template not found")
	}

	// 準備模板數據
	values := templates.SecureMenuValues{
		ViewProfileURL:  fmt.Sprintf("%s/otp/action?token=%s", svc.cfg.BaseURL, viewOTP),
		EditSettingsURL: fmt.Sprintf("%s/otp/action?token=%s", svc.cfg.BaseURL, editOTP),
		DeleteDataURL:   fmt.Sprintf("%s/otp/action?token=%s", svc.cfg.BaseURL, deleteOTP),
	}

	// 執行模板
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, values); err != nil {
		return nil, err
	}

	return NewFlexMessage(
		"安全操作選單",
		buf.Bytes(),
	), nil
}

func (svc *simpleService) handleProfileAccess(userID string) (Message, error) {
	// 生成快速訪問個人資料的 OTP
	otp, err := svc.otp.GenerateOTP(userID, "quick_profile", map[string]any{
		"quick_access": true,
		"timestamp":    fmt.Sprintf("%d", time.Now().Unix()),
	})
	if err != nil {
		return nil, err
	}

	flexJSON := fmt.Sprintf(`{
		"type": "bubble",
		"body": {
			"type": "box",
			"layout": "vertical",
			"contents": [
				{
					"type": "text",
					"text": "📋 個人資料",
					"weight": "bold",
					"size": "lg",
					"align": "center"
				},
				{
					"type": "text",
					"text": "點擊下方按鈕安全訪問您的個人資料",
					"size": "sm",
					"color": "#666666",
					"align": "center",
					"wrap": true,
					"margin": "md"
				},
				{
					"type": "button",
					"action": {
						"type": "uri",
						"label": "🔐 安全訪問個人資料",
						"uri": "%s/otp/action?token=%s"
					},
					"style": "primary",
					"color": "#1DB446",
					"margin": "lg"
				},
				{
					"type": "text",
					"text": "🔒 此連結使用一次性密碼保護\n⏰ 3分鐘內有效",
					"size": "xs",
					"color": "#888888",
					"align": "center",
					"wrap": true,
					"margin": "lg"
				}
			]
		}
	}`, svc.cfg.BaseURL, otp)

	return NewFlexMessage(
		"個人資料訪問",
		[]byte(flexJSON),
	), nil
}

func (svc *simpleService) handleSessionMenu(u *user.User) (Message, error) {
	listOTP, err := svc.otp.GenerateOTP(u.ID, "list_sessions", nil)
	if err != nil {
		return nil, err
	}

	tmpl, ok := svc.templates["session_menu"]
	if !ok {
		return nil, errors.New("session_menu template not found")
	}

	values := templates.SessionMenuValues{
		ListSessionsURL: fmt.Sprintf("%s/users/%s/session/list?token=%s", svc.cfg.BaseURL, u.Profile.Username, listOTP),
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, values); err != nil {
		return nil, err
	}

	return NewFlexMessage(
		"Session 管理選單",
		buf.Bytes(),
	), nil
}
