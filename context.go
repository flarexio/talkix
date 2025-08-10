package talkix

type ContextKey string

const (
	UserKey     ContextKey = "user"
	SessionKey  ContextKey = "session"
	OTPKey      ContextKey = "otp"
	MessagesKey ContextKey = "messages"
)
