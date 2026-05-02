package config

import "strings"

func validEnvServerPepper() bool {
	ss := GetServerPepper()
	return ss != "" && len(ss) > 20
}

func validEnvSMTP() bool {
	email := GetSMTPEmail()
	server := GetSMTPServer()
	user := GetSMTPUser()
	pwd := GetSMTPPassword()
	return email != "" && server != "" && strings.Contains(server, ":") && user != "" && pwd != ""
}

func validateGoogleOAuthClientID() bool {
	return strings.HasSuffix(GetGoogleOAuthClientID(), ".apps.googleusercontent.com")
}

func EnsureValidConfig() bool {
	return validEnvServerPepper() && validEnvSMTP() && validateGoogleOAuthClientID()
}
