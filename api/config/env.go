package config

import "os"

const (
	ENV_PEPPER                 = "APP_PEPPER"
	ENV_MODE_DEV               = "APP_DEV"
	ENV_SMTP_EMAIL             = "APP_SMTP_EMAIL"
	ENV_SMTP_SERVER            = "APP_SMTP_SERVER"
	ENV_SMTP_USER              = "APP_SMTP_USER"
	ENV_SMTP_PASSWORD          = "APP_SMTP_PASSWORD"
	ENV_GOOGLE_OAUTH_CLIENT_ID = "APP_GOOGLE_OAUTH_CLIENT_ID"
)

func GetServerPepper() string {
	return os.Getenv(ENV_PEPPER)
}

func GetSMTPEmail() string {
	return os.Getenv(ENV_SMTP_EMAIL)
}

func GetSMTPServer() string {
	return os.Getenv(ENV_SMTP_SERVER) // e.g., "smtp.example.com:587"
}

func GetSMTPUser() string {
	return os.Getenv(ENV_SMTP_USER)
}

func GetSMTPPassword() string {
	return os.Getenv(ENV_SMTP_PASSWORD)
}

func GetPSQLConnection() string {
	if IsDeploymentProduction() {
		panic("no PSQL connection for production configured")
	}
	return "host=localhost port=5432 user=example password=dev-secret dbname=example sslmode=disable"
}

func IsDeploymentProduction() bool {
	return os.Getenv(ENV_MODE_DEV) == ""
}

func GetGoogleOAuthClientID() string {
	return os.Getenv(ENV_GOOGLE_OAUTH_CLIENT_ID)
}

func GetDomainFE() string {
	if IsDeploymentProduction() {
		return DOMAIN_PROD_FE
	}
	return DOMAIN_DEV_FE
}