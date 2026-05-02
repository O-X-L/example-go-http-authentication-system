package config

import "time"

const (
	LISTENER        = "127.0.0.1:8080"
	SQL_DRIVER      = "postgres"
	DOMAIN_PROD_API = "api.example.oxl.app"
	DOMAIN_PROD_FE  = "www.example.oxl.app"
	DOMAIN_DEV_API  = "http://localhost:8080"
	DOMAIN_DEV_FE   = "http://localhost:8000"

	COOKIE_SESSION         = "sid"
	COOKIE_SESSION_PREAUTH = "sid_guest"
	COOKIE_CSRF_POSTAUTH   = "csrf"
	COOKIE_CSRF_PREAUTH    = "csrf_guest"

	HEADER_CSRF_PREAUTH  = "X-CSRF-Token-Guest"
	HEADER_CSRF_POSTAUTH = "X-CSRF-Token"

	SMTP_VISIBLE_SENDER = "Example App"
)

var (
	TIMEOUT_SESSION_FORCE        = 30 * 24 * time.Hour // forced logout even if rotated-session is still valid
	TIMEOUT_SESSION_IDLE         = 36 * time.Hour      // idle timeout
	TIMEOUT_SESSION_ROTATE       = 1 * time.Hour       // how often to rotate the tokens
	TIMEOUT_SESSION_GUEST        = 30 * time.Minute
	TIMEOUT_VERIFICATION_TOKEN   = 1 * time.Hour
	MAX_CONCURRENT_SESSIONS      = 5
	VERIFICATION_LINK_MAX_RESEND = 2
)
