package email_templates

import (
	"fmt"
	"example_api/config"
)

func GetPasswordResetVerificationEmailContentEN(verificationLink string) (string, string, string) {
	subject := "Password Reset"
	bodyHTML := fmt.Sprintf(
		`A password-reset was issued for your account.<br/>
		To create a new password click this link: <a href="%s">Verify Email</a><br/>
		If you have not issues that password-reset on %s, you can ignore this email.`,
		verificationLink, config.GetDomainFE(),
	)
	bodyPlaintext := fmt.Sprintf(
		`A password-reset was issued for your account.
		To create a new password click this link: %s
		If you have not issues that password-reset on %s, you can ignore this email.`,
		verificationLink, config.GetDomainFE(),
	)
	return subject, bodyHTML, bodyPlaintext
}
