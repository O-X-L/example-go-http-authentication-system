package email_templates

import (
	"example_api/config"
	"fmt"
)

func GetRegistrationVerificationEmailContentEN(verificationLink string) (string, string, string) {
	subject := "Welcome to APP"
	bodyHTML := fmt.Sprintf(
		`Thank you for registering on this APP.<br/>
		To verify your email address click this link: <a href="%s">Verify Email</a><br/>
		If you have not registered on %s, you can ignore this email.`,
		verificationLink, config.DOMAIN_PROD_FE,
	)
	bodyPlaintext := fmt.Sprintf(
		`Thank you for registering on this APP.
		To verify your email address open this link: %s
		If you have not registered on %s, you can ignore this email.`,
		verificationLink, config.DOMAIN_PROD_FE,
	)
	return subject, bodyHTML, bodyPlaintext
}

func GetRegistrationWelcomeEmailContentEN() (string, string, string) {
	subject := "Welcome to APP"
	bodyHTML := fmt.Sprintf(
		`Thank you for registering on this APP.<br/>
		If you have not registered on %s, you can ignore this email.`,
		config.DOMAIN_PROD_FE,
	)
	bodyPlaintext := fmt.Sprintf(
		`Thank you for registering on this APP.
		If you have not registered on %s, you can ignore this email.`,
		config.DOMAIN_PROD_FE,
	)
	return subject, bodyHTML, bodyPlaintext
}
