package main

import (
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func emailSigninLink(email string) error {
	signinToken, err := createSigninToken(email)
	if err != nil {
		return err
	}
	signinURL := appBaseURL + "/sl?t=" + signinToken

	from := mail.NewEmail("Hako", "fred@atriumph.com")
	to := mail.NewEmail("", email)
	subject := "[Hako] Sign in link"
	content := mail.NewContent("text/plain", "Click the following link to sign in: "+signinURL)
	m := mail.NewV3MailInit(from, subject, to, content)

	request := sendgrid.GetRequest(sendgridApiKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	_, err = sendgrid.API(request)
	return err
}
