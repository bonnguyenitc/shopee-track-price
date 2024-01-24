package templates

import (
	"bytes"
	"log"
	"text/template"
)

type InfoEmailSendTokenResetPassword struct {
	Title    string
	Email    string
	UrlToken string
}

const TEMPLATE_EMAIL_SEND_TOKEN_RESET_PASSWORD = `
<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
</head>
<body>
	<p>Hi {{.Email}},</p>
	<p>Click <a href="{{.UrlToken}}">here</a> to reset your password.</p>
	<p>Token will be expired in 6 hours.</p>
	<p>If you did not request this, please ignore this email.</p>
	<p>Thanks,</p>
</body>
</html>
`

func CreateEmailSendTokenResetPasswordTemplate(info InfoEmailSendTokenResetPassword) string {
	tmpl, err := template.New("email").Parse(TEMPLATE_EMAIL_SEND_TOKEN_RESET_PASSWORD)
	if err != nil {
		log.Println("Error parse template email send token verify user", err)
		return ""
	}

	var email bytes.Buffer
	if err := tmpl.Execute(&email, info); err != nil {
		return ""
	}

	return email.String()
}
