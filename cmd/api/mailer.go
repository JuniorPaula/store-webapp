package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail"
)

//go:embed templates
var emailTemplateFS embed.FS

func (app *application) SendEmail(from, to, subject, tmpl string, data interface{}) error {
	templateToRender := fmt.Sprintf("templates/%s.html.tmpl", tmpl)

	t, err := template.New("email-html").ParseFS(emailTemplateFS, templateToRender)
	if err != nil {
		app.errorLog.Println(err.Error())
		return err
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println(err.Error())
		return err
	}

	formattedMesage := tpl.String()

	templateToRender = fmt.Sprintf("templates/%s.plain.tmpl", tmpl)
	t, err = template.New("email-plain").ParseFS(emailTemplateFS, templateToRender)
	if err != nil {
		app.errorLog.Println(err.Error())
		return err
	}

	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println(err.Error())
		return err
	}

	plainMessage := tpl.String()

	serve := mail.NewSMTPClient()
	serve.Host = app.config.smtp.host
	serve.Port = app.config.smtp.port
	serve.Username = app.config.smtp.username
	serve.Password = app.config.smtp.password
	serve.Encryption = mail.EncryptionTLS
	serve.KeepAlive = false
	serve.ConnectTimeout = 10 * time.Second
	serve.SendTimeout = 10 * time.Second

	smtpClient, err := serve.Connect()
	if err != nil {
		app.errorLog.Println(err.Error())
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(subject)
	email.SetBody(mail.TextHTML, formattedMesage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	err = email.Send(smtpClient)
	if err != nil {
		app.errorLog.Println(err.Error())
		return err
	}

	app.infoLog.Printf("Email sent to %s", to)

	return nil
}
