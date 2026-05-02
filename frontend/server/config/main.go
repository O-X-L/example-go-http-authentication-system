package config

import (
	"os"
	"time"
)

const (
	SQL_DRIVER   = "postgres"
	ENV_MODE_DEV = "APP_DEV"

	DOMAIN_PROD_API = "api.example.oxl.app"
	DOMAIN_DEV_API  = "http://localhost:8080"

	COOKIE_SESSION_PREAUTH = "sid_guest"
	COOKIE_SESSION         = "sid"
	COOKIE_CSRF_PREAUTH    = "csrf_guest"
)

var (
	TIMEOUT_SESSION_GUEST = 30 * time.Minute
)

func IsDeploymentProduction() bool {
	return os.Getenv(ENV_MODE_DEV) == ""
}

func GetPSQLConnection() string {
	if IsDeploymentProduction() {
		panic("no PSQL connection for production configured")
	}
	return "host=localhost port=5432 user=example password=dev-secret dbname=example sslmode=disable"
}

func GetAPIDomain() string {
	if !IsDeploymentProduction() {
		return DOMAIN_DEV_API
	}
	return DOMAIN_PROD_API
}
