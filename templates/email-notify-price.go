package templates

import (
	"bytes"
	"html/template"
	"log"
)

type InfoEmailNotifyPrice struct {
	Title         string
	Email         string
	Price         int64
	PricePrevious int64
	LinkProduct   string
}

const TEMPLATE_EMAIL_NOTIFY_PRICE = `
<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
</head>
<body>
	<p>Hi {{.Email}},</p>
	<p>Price of product you are tracking has changed from {{.PricePrevious}} to {{.Price}}.</p>
	<p>Click <a href="{{.LinkProduct}}">here</a> to view product.</p>
	<p>Thanks,</p>
</body>
</html>
`

func CreateEmailNotifyPriceTemplate(info InfoEmailNotifyPrice) string {
	tmpl, err := template.New("email").Parse(TEMPLATE_EMAIL_NOTIFY_PRICE)
	if err != nil {
		log.Println("Error parse template email notify price", err)
		return ""
	}

	var email bytes.Buffer
	if err := tmpl.Execute(&email, info); err != nil {
		return ""
	}

	return email.String()
}
