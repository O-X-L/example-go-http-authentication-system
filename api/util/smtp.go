package util

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"time"

	"example_api/config"
	email_templates "example_api/templates/email"
)

// SendMailFunc allows overriding smtp.SendMail for testing purposes
var SendMailFunc = smtp.SendMail

// SendEmail sends an async email.
func SendEmail(recipient string, subject string, bodyHTML string, bodyPlaintext string) error {
	from := config.GetSMTPEmail()
	pass := config.GetSMTPPassword()
	user := config.GetSMTPUser()
	server := config.GetSMTPServer()

	host, _, err := net.SplitHostPort(server)
	if err != nil {
		host = server
	}

	auth := smtp.PlainAuth("", user, pass, host)
	to := []string{recipient}

	var msg bytes.Buffer
	boundary := "example-email-mime-boundary-1337"

	// HEADERS
	dateHeader := time.Now().Format(time.RFC1123Z)
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", dateHeader))
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", config.SMTP_VISIBLE_SENDER, from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")

	// PLAIN TEXT PART
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(fmt.Sprintf("%s\r\n", bodyPlaintext))

	// HTML PART
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(fmt.Sprintf("<html><body>%s</body></html>\r\n", bodyHTML))

	// END OF MESSAGE
	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	err = SendMailFunc(server, auth, from, to, msg.Bytes())
	if err != nil {
		log.Printf("Failed to send email to %s: %v", recipient, err)
	}
	return err
}

// SendRegistrationVerificationEmail sends an async verification email to the user.
func SendRegistrationVerificationEmail(recipient string, verificationLink string) error {
	if !config.IsDeploymentProduction() {
		log.Printf("Sending registration-email to %s: %s", recipient, verificationLink)
	}
	subject, bodyHTML, bodyPlaintext := email_templates.GetRegistrationVerificationEmailContentEN(verificationLink)
	return SendEmail(recipient, subject, bodyHTML, bodyPlaintext)
}

func SendRegistrationWelcomeEmail(recipient string) error {
	subject, bodyHTML, bodyPlaintext := email_templates.GetRegistrationWelcomeEmailContentEN()
	return SendEmail(recipient, subject, bodyHTML, bodyPlaintext)
}
