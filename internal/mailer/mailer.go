package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// Below we declare a new variable with the type embed FS (embedded file system) to hold
// our email templates. This has a comment directive in the format `//go:embed <path>`
// IMMEDIATELY ABOVE it, which indicates to Go that we want to store the contents of the
// ./templates directory in the templateFS embedded file stem variable.

//go:embed "templates"
var templateFS embed.FS

// Define a Mailer struct which contains a mail.Dialer instance(used to connect to an
// SMTP server) and the sender information for your emails(the name and address you
// want the email to be from, such as "Alice Smith <alice@example.com>")
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	// Initialize a new mail.Dialer instance with given SMTP server settings. We
	// also configure this to use a 5-second timeout whenever we send an email.
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	// Return a mailer instance containing the dialer and sender information
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Define a Send() method on the Mailer type. This takes the recipient email address
// as the first parameter, the name of the file containing the templates, and any
// dynamic dta for the templates as an any parameter
func (m Mailer) Send(recipient, templateFile string, data any) error {
	// Use the ParseFS() method to parse the required template file from the embedded
	// file system
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.}
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Follow the same pattern to execute the "plainBody" template and store the result
	// int the plainBody variable.
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// And likewise with html body
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "plainBody", data)
	if err != nil {
		return err
	}

	// Use the mail.NewMessage() function to initialize a new mail.Message instance.
	// Then we use the SetHeader() method to set the email recipient, sender and subject
	// headers, the SetBody() metod to se t the plain-text body, and the AddAlternative()
	// method to set the HTML body. It's important to note that AddAlternative() should
	// always be called *after* setBody()
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Call the DialAndSend() method on the dialer, passing in the message to send. This
	// opens a connection to the SMTP server, send the message, then closes the
	// connection. If there is a timeout, it will return a "dial tcp: i/o timeout"
	// error

	// Try sending the email up to three times before aborting and returning the final
	// error. We sleep for 500 ms between each attempt.
	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if nil == err {
			return nil
		}
		time.Sleep(500 * time.Millisecond)

	}
	return err

}
