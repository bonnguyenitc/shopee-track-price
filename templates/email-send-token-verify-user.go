package templates

import (
	"bytes"
	"log"
	"text/template"
)

type InfoEmailSendTokenVerifyUser struct {
	Title string
	Email string
	Token string
}

const TEMPLATE_EMAIL_SEND_TOKEN_VERIFY_USER = `
<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
</head>
<body>
	<p>Hi {{.Email}},</p>
	<p>Click <a href="http://localhost:3000/verify?token={{.Token}}">here</a> to verify your account.</p>
	<p>Token will be expired in 6 hours.</p>
	<p>If you did not request this, please ignore this email.</p>
	<p>Thanks,</p>
</body>
</html>
`

func CreateEmailSendTokenVerifyUserTemplate(info InfoEmailSendTokenVerifyUser) string {
	tmpl, err := template.New("email").Parse(TEMPLATE_EMAIL_SEND_TOKEN_VERIFY_USER)
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
