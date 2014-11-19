package app

import (
	"bytes"
	"net/smtp"
	"strings"
	"text/template"

	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// returns url related to dashboard server
func GetDashboardUrl(path string) string {
	baseUrl := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathDashboardURL))
	return strings.TrimRight(baseUrl, "#/") + "/#/" + path
}

// returns url related to storefront server
func GetStorefrontUrl(path string) string {
	baseUrl := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathStorefrontURL))
	return strings.TrimRight(baseUrl, "#/") + "/#/" + path
}

// returns url related to foundation server
func GetFoundationUrl(path string) string {
	baseUrl := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathFoundationURL))
	return strings.TrimRight(baseUrl, "/") + "/" + path
}

// sends mail via smtp server specified in config
func SendMail(to string, subject string, body string) error {

	userName := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathMailUser))
	password := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathMailPassword))

	mailServer := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathMailServer))
	mailPort := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathMailPort))
	if mailPort != "" && mailPort != "0" {
		mailPort = ":" + mailPort
	} else {
		return nil
	}

	context := map[string]interface{}{
		"From":      utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathMailFrom)),
		"To":        to,
		"Subject":   subject,
		"Body":      body,
		"Signature": utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathMailSignature)),
	}

	emailTemplateBody := `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}
Content-Type: text/html

<p>{{.Body}}</p>

<p>{{.Signature}}</p>`

	emailTemplate := template.New("emailTemplate")
	emailTemplate, err := emailTemplate.Parse(emailTemplateBody)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	var doc bytes.Buffer
	err = emailTemplate.Execute(&doc, context)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	var auth smtp.Auth = nil
	if userName != "" {
		auth = smtp.PlainAuth("", userName, password, mailServer)
	}

	err = smtp.SendMail(mailServer+mailPort, auth, userName, []string{to}, doc.Bytes())
	return env.ErrorDispatch(err)
}
